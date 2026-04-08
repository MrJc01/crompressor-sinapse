package daemon

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDaemonHealthEndpoint(t *testing.T) {
	server := NewDaemonServer("8080", "http://fake-backend:8080")
	
	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	server.router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Código de status Health errado: esperado %v, obtido %v", http.StatusOK, status)
	}

	if !strings.Contains(rr.Body.String(), "online") {
		t.Errorf("Corpo da resposta inesperado: %v", rr.Body.String())
	}
}

func TestDaemonCompletions_MissingPrompt(t *testing.T) {
	server := NewDaemonServer("8080", "http://fake-backend:8080")
	
	payload := `{"prompt": ""}` // Empty
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer([]byte(payload)))
	rr := httptest.NewRecorder()

	server.router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Código de status deveria ser Bad Request: obtido %v", status)
	}
}

func TestDaemonCompletions_FallbackExecution(t *testing.T) {
	// Com um backend Falso (porta morta), o Proxy falha graciosamente e retorna mensagem Offline.
	server := NewDaemonServer("8080", "http://127.0.0.1:9999") 
	
	payload := `{"prompt": "Resuma a arquitetura"}`
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer([]byte(payload)))
	rr := httptest.NewRecorder()

	server.router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("O daemon engole falha do Llama e devolve a degradação, logo Server code 200 esperado, obtido %v", status)
	}

	var response ResponsePayload
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Falha unmarshal resposta Mux")
	}

	if response.Bypassed != -1 || !strings.Contains(response.Content, "Offline") {
		t.Errorf("A resposta mock de degradação estrutural falhou em se manifestar")
	}
}
