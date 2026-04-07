package daemon

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/MrJc01/crompressor-sinapse/internal/cdc"
	"github.com/MrJc01/crompressor-sinapse/internal/inference"
)

// DaemonServer envelopa o ambiente HTTP Mux Server do Go com as engrenagens Neurais em estado aquecido (Persistent).
type DaemonServer struct {
	router      *http.ServeMux
	llamaBridge *inference.LlamaClient // Aponta para um Llama backend real providenciando inferências.
	cdcOpts     cdc.Options
	Port        string
}

// RequestPayload intercepta os requests das chamadas do cliente
type RequestPayload struct {
	Prompt string `json:"prompt"`
}

// ResponsePayload devolve ao cliente respostas processadas com métricas de O(1) Embeddings Bypass
type ResponsePayload struct {
	Content      string  `json:"content"`
	Bypassed     int     `json:"chunks_bypassed"`
	Computed     int     `json:"chunks_computed"`
	TTFTMs       float64 `json:"ttft_ms"`
	ProcessSpeed string  `json:"speed_metrics"`
}

// NewDaemonServer instancia um Node ativo para Inferir sob demanda remotamente via RESTful.
func NewDaemonServer(port string, llamaURL string) *DaemonServer {
	// A grande mágica O(1): Instancia-se o LRU Cache Memory Persistente. 
	// Diferentes requests de usuários compartilharão do mesmo Cache Neural em Bypass.
	sharedActivationCache := inference.NewActivationCache(20000)

	server := &DaemonServer{
		router:      http.NewServeMux(),
		llamaBridge: inference.NewLlamaClient(llamaURL, sharedActivationCache),
		cdcOpts:     cdc.DefaultOptions(),
		Port:        port,
	}

	server.RegisterRoutes()
	return server
}

// RegisterRoutes cadastra as extremidades primárias operacionais
func (ds *DaemonServer) RegisterRoutes() {
	ds.router.HandleFunc("POST /v1/chat/completions", ds.handleCompletions)
	ds.router.HandleFunc("GET /health", ds.handleHealth)
}

func (ds *DaemonServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"online", "version": "1.0-sinapse-daemon"}`))
}

func (ds *DaemonServer) handleCompletions(w http.ResponseWriter, r *http.Request) {
	var payload RequestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, `{"error":"invalid_payload"}`, http.StatusBadRequest)
		return
	}

	if payload.Prompt == "" {
		http.Error(w, `{"error":"prompt_empty"}`, http.StatusBadRequest)
		return
	}

	// Gateway Process
	proxyResult, err := ds.llamaBridge.ExecutePromptProxy(payload.Prompt, ds.cdcOpts)
	if err != nil {
		// Degradamos de forma segura para não apagar erro de conexão ao BackEnd em vez de 500 fatal.
		log.Printf("[Daemon] Proxy falhou: %v", err)
		proxyResult = &inference.EngineProxyResult{
			CapturedResponse: "Mock Error Fallback: Backend GGUF Offline ou inatingível. Cheque porta do Llama.",
			BypassedChunks:   -1,
			ComputedChunks:   -1,
		}
	}

	resp := ResponsePayload{
		Content:  proxyResult.CapturedResponse,
		Bypassed: proxyResult.BypassedChunks,
		Computed: proxyResult.ComputedChunks,
		TTFTMs:   float64(proxyResult.WallTime.Milliseconds()),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Start inicia o Servidor na bloqueia a Thread corrente rodando.
func (ds *DaemonServer) Start() error {
	log.Printf("🚀 Crompressor-Sinapse Daemon em background rodando O(1) Neural State...")
	log.Printf("📡 Endpoint da API Aberto na Porta: %s", ds.Port)
	log.Printf("🔗 Proxyfying LLM para C++ LlamaCore em: %s", ds.llamaBridge.BaseURL)
	return http.ListenAndServe(fmt.Sprintf(":%s", ds.Port), ds.router)
}
