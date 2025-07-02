package main

import (
	"crypto/sha256" // Para gerar hashes únicos para cache.
	"encoding/hex"  // Para codificar bytes de hash em uma string hexadecimal.
	"encoding/json" // Para codificação e decodificação JSON.
	"fmt"
	"log"      // Para mensagens de log.
	"net/http" // Para criar o servidor HTTP e lidar com requisições.
	"os"       // Para acessar variáveis de ambiente.

	"github.com/joho/godotenv" // Biblioteca para carregar variáveis de ambiente de um arquivo .env.
)

// Server struct contém as dependências para o servidor HTTP, incluindo o banco de dados e o serviço Gemini.
type Server struct {
	dbService         *DBService         // Serviço para interações com o banco de dados (cache).
	geminiPuzzleService *GeminiPuzzleService // Serviço para interagir com a API Gemini.
}

func main() {
	// Carrega variáveis de ambiente de um arquivo .env. Isso é principalmente para desenvolvimento local.
	// Em ambientes de produção (como GCP Cloud Run ou AWS EC2), as variáveis de ambiente
	// devem ser definidas diretamente na configuração de implantação.
	if err := godotenv.Load(); err != nil {
		log.Println("Nenhum arquivo .env encontrado, assumindo que as variáveis de ambiente estão definidas diretamente.")
	}

	// Recupera a string de conexão do banco de dados das variáveis de ambiente.
	dbConnStr := os.Getenv("DATABASE_URL")
	if dbConnStr == "" {
		log.Fatal("Variável de ambiente DATABASE_URL não definida. Por favor, forneça sua string de conexão PostgreSQL.")
	}

	// Recupera a chave da API Gemini das variáveis de ambiente.
	geminiAPIKey := os.Getenv("GEMINI_API_KEY")
	if geminiAPIKey == "" {
		log.Fatal("Variável de ambiente GEMINI_API_KEY não definida. Por favor, forneça sua chave da API Gemini.")
	}

	// Inicializa o serviço de banco de dados.
	dbService, err := NewDBService(dbConnStr)
	if err != nil {
		log.Fatalf("Falha ao inicializar o serviço de banco de dados: %v", err)
	}
	defer dbService.Close() // Garante que a conexão com o banco de dados seja fechada quando a função principal sair.

	// Inicializa o serviço de quebra-cabeças Gemini com a chave da API.
	geminiPuzzleService := NewGeminiPuzzleService(geminiAPIKey)

	// Cria uma nova instância de servidor, injetando os serviços inicializados.
	server := &Server{
		dbService:         dbService,
		geminiPuzzleService: geminiPuzzleService,
	}

	// Registra o manipulador HTTP para o endpoint /generate-puzzle.
	http.HandleFunc("/generate-puzzle", server.generatePuzzleHandler)

	// Determina a porta para escutar. Padrão para 8080 se não especificado nas variáveis de ambiente.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Servidor iniciando na porta %s...", port)
	// Inicia o servidor HTTP. log.Fatal fará com que o programa seja encerrado se o servidor falhar ao iniciar.
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// generatePuzzleHandler é o manipulador HTTP para requisições de geração de quebra-cabeças.
// Ele lida com a lógica de cache: verifica o cache, chama o Gemini se não encontrado e salva no cache.
func (s *Server) generatePuzzleHandler(w http.ResponseWriter, r *http.Request) {
	// Garante que apenas requisições POST sejam permitidas.
	if r.Method != http.MethodPost {
		http.Error(w, "Apenas requisições POST são permitidas para este endpoint.", http.StatusMethodNotAllowed)
		return
	}

	var req PuzzleRequest
	// Decodifica o corpo da requisição JSON para a struct PuzzleRequest.
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Payload de requisição inválido: %v", err), http.StatusBadRequest)
		return
	}

	// Serializa a struct da requisição de volta para JSON para criar uma representação de bytes consistente para hashing.
	// Isso garante que o hash seja o mesmo para requisições idênticas, independentemente da ordem de iteração do mapa, etc.
	reqBytes, err := json.Marshal(req)
	if err != nil {
		log.Printf("Erro ao serializar a requisição para hashing: %v", err)
		http.Error(w, "Erro interno do servidor: Falha ao processar a requisição.", http.StatusInternalServerError)
		return
	}

	// Gera um hash SHA256 dos bytes da requisição. Este hash serve como chave de cache.
	hasher := sha256.New()
	hasher.Write(reqBytes)
	requestHash := hex.EncodeToString(hasher.Sum(nil)) // Converte o hash para uma string hexadecimal.

	// Tenta recuperar uma resposta em cache do banco de dados.
	cachedResponse, err := s.dbService.GetCachedPuzzle(requestHash)
	if err != nil {
		log.Printf("Erro ao verificar o cache para o hash %s: %v", requestHash, err)
		// Registra o erro, mas continua o processamento; uma falha na verificação do cache não deve bloquear a requisição.
	}

	// Se uma resposta em cache for encontrada, retorne-a imediatamente.
	if cachedResponse != nil {
		w.Header().Set("Content-Type", "application/json") // Define o tipo de conteúdo da resposta.
		w.Write(cachedResponse)                            // Escreve os dados JSON em cache na resposta.
		log.Printf("Retornada resposta em cache para o hash: %s", requestHash)
		return // Encerra o processamento da requisição aqui.
	}

	// Se nenhuma resposta em cache, chama a API Gemini para gerar um novo quebra-cabeça.
	geminiResponse, err := s.geminiPuzzleService.GeneratePuzzle(req)
	if err != nil {
		log.Printf("Erro ao gerar quebra-cabeça da API Gemini para a requisição %+v: %v", req, err)
		http.Error(w, fmt.Sprintf("Falha ao gerar quebra-cabeça: %v", err), http.StatusInternalServerError)
		return
	}

	// Após obter uma resposta com sucesso do Gemini, salve-a no cache.
	err = s.dbService.SaveCachedPuzzle(requestHash, reqBytes, geminiResponse)
	if err != nil {
		log.Printf("Erro ao salvar quebra-cabeça no cache para o hash %s: %v", requestHash, err)
		// Registra o erro, mas continua a retornar a resposta; uma falha ao salvar no cache não deve bloquear o usuário.
	}

	// Define o tipo de conteúdo e escreve a resposta do Gemini de volta para o cliente.
	w.Header().Set("Content-Type", "application/json")
	w.Write(geminiResponse)
	log.Printf("Nova resposta gerada e salva no cache para o hash: %s", requestHash)
}
