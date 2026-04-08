#!/bin/bash

# ==============================================================================
# Crompressor-Sinapse | Interactive Mock CLI Chat 
# ==============================================================================

# Cores para o Terminal
CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
GRAY='\033[1;30m'
NC='\033[0m' # No Color

# Endpoint do Daemon Sinapse (Fase 4 local backend)
DAEMON_URL="http://127.0.0.1:8080/v1/chat/completions"

echo -e "${CYAN}╔══════════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║     🧠 Crompressor-Sinapse | Chat Interativo (Terminal)      ║${NC}"
echo -e "${CYAN}╚══════════════════════════════════════════════════════════════╝${NC}"
echo -e "${GRAY}(Digite 'sair' ou 'exit' para interromper a sessão.)${NC}\n"

# Handler para desligar daemons abertos quando sair do Chat
cleanup() {
    echo -e "${GRAY}Desligando servidores alocados em Background...${NC}"
    kill $MOCK_PID 2>/dev/null
    kill $DAEMON_PID 2>/dev/null
}
trap cleanup EXIT SIGINT SIGTERM

# Helper boot automático das engines se 8080 não estiver rodando (Zero config pra o usuário)
if ! curl -s -f -o /dev/null "$DAEMON_URL"; then
    echo -e "${YELLOW}Bootando Mock Llama em Background (:9090)...${NC}"
    go run scripts/mock_llm.go > /dev/null 2>&1 &
    MOCK_PID=$!
    sleep 1

    echo -e "${YELLOW}Bootando Gateway Daemon MUX em Background (:8080)...${NC}"
    ./bin/sinapse serve --port 8080 --llama-url http://127.0.0.1:9090 > /dev/null 2>&1 &
    DAEMON_PID=$!
    sleep 2
    
    echo -e "${GREEN}✅ Todos os motores ligados. Contexto Persistente Global ativado.${NC}\n"
fi

while true; do
  echo -e -n "\n${GREEN}Você:${NC} "
  read -r user_prompt

  # Checks de quebra break
  if [[ "$user_prompt" == "sair" || "$user_prompt" == "exit" ]]; then
      echo -e "${CYAN}Encerrando canal Sinapse. Até mais!${NC}"
      break
  fi
  
  if [[ -z "$user_prompt" ]]; then
      continue
  fi

  # Escapar double quotes pra injeção no curl json
  escaped_prompt="${user_prompt//\"/\\\"}"
  json_payload="{\"prompt\": \"$escaped_prompt\"}"

  echo -e "${GRAY}[Processando via Daemon O(1)...]${NC}"

  # Realizando POST silencioso e guardando resposta raw
  raw_response=$(curl -s -X POST "$DAEMON_URL" \
       -H "Content-Type: application/json" \
       -d "$json_payload")

  # Verificação base se o servidor não está morto
  if [[ -z "$raw_response" ]]; then
      echo -e "\n${YELLOW}⚠️  Llama/Daemon Desligado ou Erro de conexão Local.${NC}"
      continue
  fi

  # Usando Python in-line para processar JSON puro e sem dependências do 'jq' no Linux bash de dev
  content=$(echo "$raw_response" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data.get('content', ''))" 2>/dev/null)
  bypassed=$(echo "$raw_response" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data.get('chunks_bypassed', '0'))" 2>/dev/null)
  ttft=$(echo "$raw_response" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data.get('ttft_ms', '0.0'))" 2>/dev/null)
  
  if [[ -z "$content" || "$bypassed" == "FAIL" || "$bypassed" == "-1" ]]; then
      # Recupera Fallbacks graciosos
      content=$(echo "$raw_response" | grep -o '"content":"[^"]*' | cut -d '"' -f 4)
      echo -e "\n${YELLOW}Sinapse Offline Fallback:${NC} $content"
      continue
  fi

  # Resposta Final exibindo O(1) metrics
  echo -e "\n${CYAN}Sinapse AI:${NC} $content"
  echo -e "${GRAY} ⚡ Métricas do Cache: TTFT ${ttft}ms | CDC Bypass LRU O(1): ${bypassed} Chunks economizados.${NC}"

done
