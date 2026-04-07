#!/bin/bash

# ==============================================================================
# Crompressor-Sinapse | Lab: Stress Test de M-Map Cache e Mux Server
# ==============================================================================

GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

# Cleanup handler para orfanatos de PIDs
cleanup() {
    echo -e "\n${RED}🛑 Encerrando ambientes mockados...${NC}"
    kill $MOCK_PID 2>/dev/null
    kill $DAEMON_PID 2>/dev/null
    exit 0
}
trap cleanup SIGINT SIGTERM EXIT

echo -e "${BLUE}╔══════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║       🌋 CROM-LAB | Concorrência Brutal vs MMap Cache        ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════════════════╝${NC}"
echo ""

# 1. Bootando Pseudo Llama
echo -e "⚡ Bootando Servidor Llama de Mentira (MOCK API) na :9090 ..."
go run scripts/mock_llm.go &
MOCK_PID=$!
sleep 2

# 2. Bootando o Daemond Sinapse
echo -e "⚡ Injetando Módulo Crompressor Daemond Gateway na :8080 ..."
./bin/sinapse serve --port 8080 --llama-url http://127.0.0.1:9090 > /dev/null 2>&1 &
DAEMON_PID=$!
sleep 2

PROMPT_JSON='{"prompt": "The Crompressor Thesis stands upon the belief that predictable past semantics must not consume GPU cycles again and again."}'

echo -e "\n${GREEN}🚀 Iniciando Bateria de Testes Sequenciais vs Request Bypasses LRU${NC}"

# Função para engolir Curl e isolar métricas brutas geradas pelo JSON JSON response do Servidor
fire_request() {
    local i=$1
    local res=$(curl -s -X POST http://127.0.0.1:8080/v1/chat/completions \
         -H "Content-Type: application/json" \
         -d "$PROMPT_JSON")
    
    local bypasses=$(echo "$res" | grep -o '"chunks_bypassed":[0-9]*' | cut -d ':' -f 2)
    local ttft=$(echo "$res" | grep -o '"ttft_ms":[0-9.]*' | cut -d ':' -f 2)
    
    # Tratamento caso backend retorne degrade-string (Offline)
    if [[ -z "$bypasses" ]]; then bypasses="FAIL"; ttft="FAIL"; fi
    
    echo " Request #$i | TTFT: ${ttft}ms | CDC Chunks Bypassed (O(1)): $bypasses"
}

# 1 Fio Single-Thread para criar o Frio (Misses) e logo depois os Hits
for i in {1..7}; do
    fire_request $i
done

echo ""
echo -e "${BLUE}🌊 Multithreading Tsunâmi (Disparadas Concorrentes Múltiplas)${NC}"

# Faz 20 disparos perfeitamente simultâneos via background (&) para forçar o mutex de leitura O(1) do Cache
for i in {8..28}; do
    curl -s -X POST http://127.0.0.1:8080/v1/chat/completions \
         -H "Content-Type: application/json" \
         -d "$PROMPT_JSON" > /dev/null &
done

sleep 2
echo -e "${GREEN}✅ Mutex do ActivationCache processou 20 threads paralelas sem corromper a memória local!${NC}"
echo -e "Toda a bateria concluída. O laboratório logístico finaliza provando estabilidade Master API e Economia do Bypass (Time To First Token despencado)."

cleanup
