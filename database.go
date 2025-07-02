package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq" // Driver PostgreSQL para database/sql
)

// DBService lida com todas as operações de banco de dados, especificamente para cache de respostas de quebra-cabeças.
type DBService struct {
	db *sql.DB // O pool de conexão do banco de dados subjacente
}

// NewDBService inicializa um novo DBService estabelecendo uma conexão com o banco de dados PostgreSQL.
// Ele recebe uma string de conexão (ex: "postgres://user:password@host:port/database_name?sslmode=disable")
// e retorna um ponteiro para DBService ou um erro se a conexão falhar.
func NewDBService(connStr string) (*DBService, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("falha ao abrir a conexão com o banco de dados: %w", err)
	}

	// Ping no banco de dados para verificar se a conexão está ativa e válida.
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("falha ao conectar ao banco de dados: %w", err)
	}

	log.Println("Conectado com sucesso ao banco de dados PostgreSQL.")
	return &DBService{db: db}, nil
}

// Close fecha a conexão com o banco de dados. É importante adiar esta chamada
// na função principal para garantir a limpeza adequada dos recursos.
func (s *DBService) Close() error {
	log.Println("Fechando conexão com o banco de dados.")
	return s.db.Close()
}

// GetCachedPuzzle recupera uma resposta de quebra-cabeça em cache do banco de dados usando um hash de requisição.
// Retorna os dados da resposta em cache como um slice de bytes, se encontrado, ou nil se não encontrado (sql.ErrNoRows).
// Qualquer outro erro de banco de dados será retornado.
func (s *DBService) GetCachedPuzzle(requestHash string) ([]byte, error) {
	var responseData []byte // Variável para armazenar os dados da resposta recuperados
	query := "SELECT response_data FROM cached_puzzles WHERE request_hash = $1"

	// QueryRow executa uma consulta que deve retornar no máximo uma linha.
	// Scan copia as colunas da linha correspondente para a variável responseData.
	err := s.db.QueryRow(query, requestHash).Scan(&responseData)
	if err == sql.ErrNoRows {
		return nil, nil // Nenhuma linha encontrada, indicando um cache miss
	}
	if err != nil {
		return nil, fmt.Errorf("falha ao obter quebra-cabeça em cache para o hash %s: %w", requestHash, err)
	}
	log.Printf("Cache hit para o hash: %s", requestHash)
	return responseData, nil
}

// SaveCachedPuzzle salva uma resposta de quebra-cabeça no cache do banco de dados.
// Ele recebe o hash da requisição, os parâmetros da requisição original e os dados da resposta do Gemini.
// Ele usa um UPSERT (ON CONFLICT DO UPDATE) para inserir um novo registro ou atualizar um existente
// se um registro com o mesmo request_hash já existir.
func (s *DBService) SaveCachedPuzzle(requestHash string, requestParams []byte, responseData []byte) error {
	query := `
		INSERT INTO cached_puzzles (request_hash, request_params, response_data, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (request_hash) DO UPDATE SET
			request_params = EXCLUDED.request_params,
			response_data = EXCLUDED.response_data,
			created_at = EXCLUDED.created_at
	`
	// Exec executa uma consulta sem retornar nenhuma linha.
	_, err := s.db.Exec(query, requestHash, requestParams, responseData, time.Now())
	if err != nil {
		return fmt.Errorf("falha ao salvar quebra-cabeça em cache para o hash %s: %w", requestHash, err)
	}
	log.Printf("Cache salvo para o hash: %s", requestHash)
	return nil
}
