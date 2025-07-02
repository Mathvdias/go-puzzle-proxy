package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

// geminiAPIURL é o endpoint para o modelo Gemini 2.0 Flash.
const geminiAPIURL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent"

// GeminiPuzzleService lida com as interações com a API Gemini.
type GeminiPuzzleService struct {
	apiKey string // A chave da API Gemini, mantida secreta no servidor.
}

// NewGeminiPuzzleService cria e retorna uma nova instância de GeminiPuzzleService.
// Requer que a chave da API Gemini seja passada durante a inicialização.
func NewGeminiPuzzleService(apiKey string) *GeminiPuzzleService {
	return &GeminiPuzzleService{apiKey: apiKey}
}

// GeneratePuzzle constrói o prompt e o schema apropriados, então chama a API Gemini
// para gerar um quebra-cabeça com base nos parâmetros de requisição fornecidos.
// Retorna a resposta JSON bruta do Gemini como um slice de bytes ou um erro.
func (s *GeminiPuzzleService) GeneratePuzzle(req PuzzleRequest) ([]byte, error) {
	// Validação básica para a chave da API.
	if s.apiKey == "" || s.apiKey == "YOUR_GEMINI_API_KEY_HERE" {
		return nil, fmt.Errorf("GEMINI_API_KEY não definida ou é o valor padrão. Por favor, defina-a como uma variável de ambiente")
	}

	// Determina a string do tipo de jogo para o prompt.
	gameTypeString := ""
	if req.GameType == "crossword" {
		gameTypeString = "crossword puzzle"
	} else { // Assumindo req.GameType == "wordsearch"
		gameTypeString = "word search puzzle"
	}

	// Normaliza a string de dificuldade para minúsculas.
	difficultyString := strings.ToLower(req.Difficulty)

	// Constrói a string de tópicos para o prompt.
	topicsString := ""
	if len(req.Topics) > 0 {
		topicsString = fmt.Sprintf("about %s", strings.Join(req.Topics, ", "))
	} else {
		topicsString = "general knowledge" // Tópico padrão se nenhum for fornecido.
	}

	prompt := ""
	var schemaBytes []byte // Usaremos um slice de bytes temporário para o schema JSON

	// Lógica para construir o prompt e o schema de resposta com base no tipo de jogo.
	if req.GameType == "crossword" {
		prompt = fmt.Sprintf(`
			Generate a %s %s in %s.
			%s
			Provide a grid of 8x8 to 10x10.
			Return the data as a JSON object with 'gameType' (crossword), 'difficulty', 'topics', and 'crosswordData'.
			'crosswordData' should contain 'gridSize' (rows, cols) and an array of 'words'.
			Each 'word' object should have 'word', 'clue', 'startRow', 'startCol' (0-indexed), and 'direction' ('across' or 'down').
			Ensure words fit the grid and intersect correctly without gaps. All cells in a word must be valid letters.
			Prioritize well-formed and solvable puzzles.
		`, difficultyString, gameTypeString, req.Language, topicsString)

		// Schema JSON específico para palavras cruzadas.
		schemaBytes = []byte(`{
			"type": "OBJECT",
			"properties": {
				"gameType": {
					"type": "STRING",
					"enum": ["crossword"]
				},
				"difficulty": {
					"type": "STRING",
					"enum": ["easy", "medium", "hard"]
				},
				"topics": {
					"type": "ARRAY",
					"items": {"type": "STRING"}
				},
				"crosswordData": {
					"type": "OBJECT",
					"properties": {
						"gridSize": {
							"type": "OBJECT",
							"properties": {
								"rows": {"type": "INTEGER"},
								"cols": {"type": "INTEGER"}
							},
							"required": ["rows", "cols"]
						},
						"words": {
							"type": "ARRAY",
							"items": {
								"type": "OBJECT",
								"properties": {
									"word": {"type": "STRING"},
									"clue": {"type": "STRING"},
									"startRow": {"type": "INTEGER"},
									"startCol": {"type": "INTEGER"},
									"direction": {
										"type": "STRING",
										"enum": ["across", "down"]
									}
								},
								"required": ["word", "clue", "startRow", "startCol", "direction"]
							}
						}
					},
					"required": ["gridSize", "words"]
				}
			},
			"required": ["gameType", "difficulty", "topics"]
		}`)
	} else { // Caça-palavras
		prompt = fmt.Sprintf(`
			Generate a %s %s in %s.
			%s
			Provide a grid size based on difficulty: Easy (10x10), Medium (12x12), Hard (15x15).
			Return the data as a JSON object with 'gameType' (wordsearch), 'difficulty', 'topics', and 'wordSearchData'.
			'wordSearchData' should contain 'gridSize' (rows, cols) and a list of 'wordsToFind'.
			**Crucially, do NOT generate the full grid of letters. ONLY provide gridSize and wordsToFind.**
			The 'wordsToFind' list should contain 10-15 unique words (depending on difficulty) that are relevant to the topics and suitable for a word search puzzle (e.g., no spaces, only letters, common vocabulary).
			Ensure these words are always in the uppercase.
			Prioritize well-formed words and a good mix for the chosen difficulty.
		`, difficultyString, gameTypeString, req.Language, topicsString)

		// Schema JSON específico para caça-palavras.
		schemaBytes = []byte(`{
			"type": "OBJECT",
			"properties": {
				"gameType": {
					"type": "STRING",
					"enum": ["wordsearch"]
				},
				"difficulty": {
					"type": "STRING",
					"enum": ["easy", "medium", "hard"]
				},
				"topics": {
					"type": "ARRAY",
					"items": {"type": "STRING"}
				},
				"wordSearchData": {
					"type": "OBJECT",
					"properties": {
						"gridSize": {
							"type": "OBJECT",
							"properties": {
								"rows": {"type": "INTEGER"},
								"cols": {"type": "INTEGER"}
							},
							"required": ["rows", "cols"]
						},
						"wordsToFind": {
							"type": "ARRAY",
							"items": {"type": "STRING"}
						}
					},
					"required": ["gridSize", "wordsToFind"]
				}
			},
			"required": ["gameType", "difficulty", "topics"]
		}`)
	}

	// NOVO: Parse o schema JSON em um map e então marshal de volta para RawMessage.
	// Isso garante que o json.RawMessage contenha JSON válido.
	var parsedSchema map[string]interface{}
	if err := json.Unmarshal(schemaBytes, &parsedSchema); err != nil {
		return nil, fmt.Errorf("falha ao parsear o schema JSON para map: %w", err)
	}
	responseSchema, err := json.Marshal(parsedSchema)
	if err != nil {
		return nil, fmt.Errorf("falha ao serializar o map do schema para json.RawMessage: %w", err)
	}

	// Constrói o payload da requisição para a API Gemini.
	geminiReq := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiContentPart{
					{Text: prompt},
				},
			},
		},
		GenerationConfig: GeminiGenerationConfig{
			ResponseMimeType: "application/json",
			ResponseSchema:   responseSchema, // Usa o json.RawMessage validado
			Temperature:      0.7, // Ajuste conforme necessário para criatividade vs. consistência.
			TopP:             0.9,
			TopK:             40,
		},
	}

	// Serializa a struct Go para um slice de bytes JSON para o corpo da requisição HTTP.
	jsonReqBody, err := json.Marshal(geminiReq)
	if err != nil {
		return nil, fmt.Errorf("falha ao serializar a requisição Gemini: %w", err)
	}

	log.Printf("Chamando a API Gemini com prompt (truncado): %s...", prompt[:min(len(prompt), 100)]) // Registra um prompt truncado para brevidade.
	client := &http.Client{} // Cria um novo cliente HTTP.

	// Cria uma nova requisição POST para o endpoint da API Gemini, incluindo a chave da API na string de consulta.
	httpReq, err := http.NewRequest("POST", fmt.Sprintf("%s?key=%s", geminiAPIURL, s.apiKey), bytes.NewBuffer(jsonReqBody))
	if err != nil {
		return nil, fmt.Errorf("falha ao criar requisição HTTP: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json") // Define o cabeçalho do tipo de conteúdo.

	// Executa a requisição HTTP.
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("falha ao fazer requisição HTTP para Gemini: %w", err)
	}
	defer resp.Body.Close() // Garante que o corpo da resposta seja fechado após a leitura.

	// Lê o corpo completo da resposta.
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("falha ao ler o corpo da resposta Gemini: %w", err)
	}

	// Verifica códigos de status HTTP diferentes de 200 do Gemini.
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API Gemini falhou com status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Deserializa a resposta da API Gemini para a struct GeminiAPIResponse.
	var geminiAPIResp GeminiAPIResponse
	if err := json.Unmarshal(bodyBytes, &geminiAPIResp); err != nil {
		return nil, fmt.Errorf("falha ao deserializar a resposta da API Gemini: %w. Resposta bruta: %s", err, string(bodyBytes))
	}

	// Valida se a resposta contém candidatos e conteúdo.
	if len(geminiAPIResp.Candidates) == 0 || len(geminiAPIResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("A resposta da API Gemini estava vazia ou inesperada. Resposta bruta: %s", string(bodyBytes))
	}

	// Extrai o texto gerado (string JSON) da resposta Gemini.
	jsonString := geminiAPIResp.Candidates[0].Content.Parts[0].Text
	log.Println("Resposta da API Gemini recebida com sucesso.")
	return []byte(jsonString), nil // Retorna a string JSON bruta do Gemini.
}

// min é uma função auxiliar para obter o mínimo de dois inteiros.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
