package main

import "encoding/json"

// PuzzleRequest representa o payload da requisição recebida do aplicativo Dart.
// Contém os parâmetros para gerar um quebra-cabeça.
type PuzzleRequest struct {
	GameType   string   `json:"gameType"`   // Ex: "crossword", "wordsearch"
	Difficulty string   `json:"difficulty"` // Ex: "easy", "medium", "hard"
	Topics     []string `json:"topics"`     // Lista de tópicos para o quebra-cabeça
	Language   string   `json:"language"`   // Idioma do quebra-cabeça
}

// CrosswordWord representa uma única palavra dentro de um quebra-cabeça de palavras cruzadas.
type CrosswordWord struct {
	Word      string `json:"word"`      // A palavra real
	Clue      string `json:"clue"`      // A dica para a palavra
	StartRow  int    `json:"startRow"`  // Linha inicial (base 0) na grade
	StartCol  int    `json:"startCol"`  // Coluna inicial (base 0) na grade
	Direction string `json:"direction"` // "across" (horizontal) ou "down" (vertical)
}

// CrosswordData encapsula todos os dados específicos de um quebra-cabeça de palavras cruzadas.
type CrosswordData struct {
	GridSize struct { // Dimensões da grade de palavras cruzadas
		Rows int `json:"rows"`
		Cols int `json:"cols"`
	} `json:"gridSize"`
	Words []CrosswordWord `json:"words"` // Lista de palavras nas palavras cruzadas
}

// WordSearchData encapsula todos os dados específicos de um caça-palavras.
type WordSearchData struct {
	GridSize struct { // Dimensões da grade do caça-palavras
		Rows int `json:"rows"`
		Cols int `json:"cols"`
	} `json:"gridSize"`
	WordsToFind []string `json:"wordsToFind"` // Lista de palavras a serem encontradas no caça-palavras
}

// GeminiPuzzleResponse representa a resposta estruturada esperada da API Gemini,
// que será retornada ao aplicativo Dart.
type GeminiPuzzleResponse struct {
	GameType       string          `json:"gameType"`              // Tipo de jogo (crossword ou wordsearch)
	Difficulty     string          `json:"difficulty"`            // Nível de dificuldade
	Topics         []string        `json:"topics"`                // Tópicos usados para o quebra-cabeça
	CrosswordData  *CrosswordData  `json:"crosswordData,omitempty"`  // Dados de palavras cruzadas (se o gameType for crossword)
	WordSearchData *WordSearchData `json:"wordSearchData,omitempty"` // Dados de caça-palavras (se o gameType for wordsearch)
}

// GeminiContentPart representa uma parte do conteúdo dentro de uma requisição/resposta da API Gemini.
// Para geração de texto, geralmente contém o campo 'text'.
type GeminiContentPart struct {
	Text string `json:"text"`
}

// GeminiContent representa o bloco de conteúdo em uma requisição/resposta da API Gemini.
type GeminiContent struct {
	Parts []GeminiContentPart `json:"parts"`
}

// GeminiGenerationConfig define a configuração para o processo de geração da API Gemini.
// Inclui o formato da resposta (schema) e parâmetros de geração.
type GeminiGenerationConfig struct {
	ResponseMimeType string          `json:"responseMimeType"`   // Tipo MIME esperado da resposta (ex: "application/json")
	ResponseSchema   json.RawMessage `json:"responseSchema"`     // Schema JSON para a estrutura de saída desejada
	Temperature      float64         `json:"temperature"`        // Controla a aleatoriedade na geração
	TopP             float64         `json:"topP"`               // Controla a diversidade via amostragem de núcleo
	TopK             int             `json:"topK"`               // Controla a diversidade via amostragem top-k
}

// GeminiRequest representa o payload completo enviado à API Gemini.
type GeminiRequest struct {
	Contents         []GeminiContent        `json:"contents"`         // O conteúdo de entrada (ex: prompt)
	GenerationConfig GeminiGenerationConfig `json:"generationConfig"` // Configuração para geração
}

// GeminiCandidate representa uma resposta candidata gerada pela API Gemini.
type GeminiCandidate struct {
	Content GeminiContent `json:"content"` // O conteúdo gerado
}

// GeminiAPIResponse representa a resposta completa recebida da API Gemini.
type GeminiAPIResponse struct {
	Candidates []GeminiCandidate `json:"candidates"` // Lista de candidatos gerados
}
