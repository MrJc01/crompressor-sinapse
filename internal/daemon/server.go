package daemon

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/MrJc01/crompressor-sinapse/internal/cdc"
	"github.com/MrJc01/crompressor-sinapse/internal/inference"
	"github.com/MrJc01/crompressor-sinapse/internal/p2p"
)

// DaemonServer envelopa o ambiente HTTP Mux Server do Go com as engrenagens Neurais em estado aquecido (Persistent).
type DaemonServer struct {
	router      *http.ServeMux
	llamaBridge *inference.LlamaClient
	cdcOpts     cdc.Options
	p2pNode     *p2p.Node  // Integrando a malha do enxame distribuído
	Port        string
	ID          string     // Nó ID Local
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

func NewDaemonServer(port string, llamaURL string, nodeID string, peers []string) *DaemonServer {
	sharedActivationCache := inference.NewActivationCache(20000)

	server := &DaemonServer{
		router:      http.NewServeMux(),
		llamaBridge: inference.NewLlamaClient(llamaURL, sharedActivationCache),
		cdcOpts:     cdc.DefaultOptions(),
		p2pNode:     p2p.NewNode(nodeID, peers, sharedActivationCache),
		Port:        port,
		ID:          nodeID,
	}

	server.RegisterRoutes()
	return server
}

func (ds *DaemonServer) RegisterRoutes() {
	ds.router.HandleFunc("POST /v1/chat/completions", ds.handleCompletions)
	ds.router.HandleFunc("GET /health", ds.handleHealth)
	ds.router.HandleFunc("POST /p2p/gossip", ds.p2pNode.HandleGossip)
}

func (ds *DaemonServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"status":"online", "version": "2.0-sinapse-p2p", "node_id": "%s"}`, ds.ID)))
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

	// Gatilho de inteligência local, o proxyResult traz as métricas de novos hashes
	proxyResult, err := ds.llamaBridge.ExecutePromptProxy(payload.Prompt, ds.cdcOpts)
	if err != nil {
		log.Printf("[Daemon] Proxy falhou: %v", err)
		proxyResult = &inference.EngineProxyResult{
			CapturedResponse: "Mock Error Fallback: Backend GGUF Offline ou inatingível. Cheque porta do Llama.",
			BypassedChunks:   -1,
			ComputedChunks:   -1,
		}
	} else {
		// Hook Gossip: Se gerou novos calculos locais pesados (computed > 0), compartilha o resultado via P2P
        // O `ExecutePromptProxy` precisaria retornar os novos vetores e hashes, 
        // mas para fins laboratoriais, iremos captar o prompt cru e broadcastear o 1º Token Token.
        if proxyResult.ComputedChunks > 0 {
             chunks := cdc.Tokenize([]byte(payload.Prompt), ds.cdcOpts)
             if len(chunks) > 0 {
                  // Emitindo de graça para os vizinhos o Vetor recém materializado pela LLM local
                  ds.p2pNode.BroadcastHashedVector(chunks[0].Hash, []float32{1.0, 1.5, 2.0})
             }
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
