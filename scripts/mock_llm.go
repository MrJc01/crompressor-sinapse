package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/completion", func(w http.ResponseWriter, r *http.Request) {
		// Simula um Llama compute rápido (50ms)
		time.Sleep(50 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		
		resp := map[string]string{
			"content": "Conclusão neural processada pelo Mock Server em 50ms.",
		}
		json.NewEncoder(w).Encode(resp)
	})

	fmt.Println("🤖 Mock LLM Llama Server iniciado na Porta :9090 (Aguardando proxy Daemon)")
	http.ListenAndServe(":9090", nil)
}
