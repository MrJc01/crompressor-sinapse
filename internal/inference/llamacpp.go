package inference

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/MrJc01/crompressor-sinapse/internal/cdc"
)

// LlamaClient encapsula as chamadas diretas a um Servidor ativo Native Llama.cpp ou similar (API Compatível).
// Nós isolaremos a inferência via Proxy para que o Daemon Go possa agir sem exigir CGO rígido no Build Stage.
type LlamaClient struct {
	BaseURL    string
	ClientHTTP *http.Client
	Cache      *ActivationCache
}

// LlamaRequest modelo de requisição openai padrão para completions
type LlamaRequest struct {
	Prompt      string  `json:"prompt"`
	Temperature float64 `json:"temperature,omitempty"`
	MaxTokens   int     `json:"n_predict,omitempty"`
}

// LlamaResponse representa a saída json do backend Llama cpp
type LlamaResponse struct {
	Content string `json:"content"`
}

// EngineProxyResult carrega os dados sobre o processo final na API real Llama e métricas de bypass no front local
type EngineProxyResult struct {
	CapturedResponse string
	BypassedChunks   int
	ComputedChunks   int
	WallTime         time.Duration
}

// NewLlamaClient devolve a conexão estruturada com os Caches acoplados
func NewLlamaClient(endpoint string, cache *ActivationCache) *LlamaClient {
	return &LlamaClient{
		BaseURL: endpoint,
		ClientHTTP: &http.Client{
			Timeout: time.Second * 300, // Long polling de IA
		},
		Cache: cache,
	}
}

// ExecutePromptProxy engulhe o prompt enviado pelo cliente, tritura os Chunks via CDC,
// salva as simulações no LRU Local para quantificação e despacha o pedido ao LLM Llama real via HTTP.
func (lc *LlamaClient) ExecutePromptProxy(prompt string, opts cdc.Options) (*EngineProxyResult, error) {
	start := time.Now()

	chunks := cdc.Tokenize([]byte(prompt), opts)
	res := &EngineProxyResult{}

	// Bypassing simulado no Gateway Layer do Proxy
	for _, chunk := range chunks {
		if _, hit := lc.Cache.Get(chunk.Hash); hit {
			res.BypassedChunks++
		} else {
			res.ComputedChunks++
			// Persistimos fake vectors p/ alimentar simulação TTFT Cache em Requests Futuros
			lc.Cache.Put(chunk.Hash, []float32{0.0, 1.0, 2.0}) 
		}
	}

	reqBody := LlamaRequest{
		Prompt:      prompt,
		Temperature: 0.1,
		MaxTokens:   512,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("falha ao encodar requisição Llama: %v", err)
	}

	req, err := http.NewRequest("POST", lc.BaseURL+"/completion", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	httpRes, err := lc.ClientHTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("a conexão ao backend llama falhou: %v", err)
	}
	defer httpRes.Body.Close()

	if httpRes.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("backend llama devolveu erro http: %d", httpRes.StatusCode)
	}

	bodyBytes, _ := io.ReadAll(httpRes.Body)
	var llamaDecoded LlamaResponse
	if err := json.Unmarshal(bodyBytes, &llamaDecoded); err != nil {
		return nil, fmt.Errorf("falha unmarshal server proxy json: %v", err)
	}

	res.CapturedResponse = llamaDecoded.Content
	res.WallTime = time.Since(start)

	return res, nil
}
