package daemon

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

// TestDaemonCompletions_RaceCondition envia dezenas de sessões concorrentes de proxy contra a malha Shared Memory do Cache Neural.
func TestDaemonCompletions_RaceCondition(t *testing.T) {
	// A Fake Llama backend handler so the mux doesn't fail natively
	fakeLlama := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"content":"success"}`))
	}))
	defer fakeLlama.Close()

	// Inicia a aplicação usando Porta 8000 dummy
	server := NewDaemonServer("8800", fakeLlama.URL)

	payloadBytes := []byte(`{"prompt": "Multithreaded CDC Semantic Context Collision Detection!"}`)

	// Disparo concorrente
	workers := 50
	var wg sync.WaitGroup
	wg.Add(workers)

	// Usaremos requests em Goroutines massivos para atestar Data Raça em Go Tools Map
	for i := 0; i < workers; i++ {
		go func(routineID int) {
			defer wg.Done()
			
			req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer(payloadBytes))
			rr := httptest.NewRecorder()

			server.router.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusOK {
				t.Errorf("[Goroutine %d] Fallback Crítico Inesperado, status = %v", routineID, status)
			}
			
			var response ResponsePayload
			if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
				t.Errorf("[Goroutine %d] Broken JSON response.", routineID)
			}
		}(i)
	}

	// Aguardando fechamento do stress massivo
	wg.Wait()
}
