#!/bin/bash

# ==============================================================================
# Crompressor-Sinapse | Lab: Rede Mesh P2P O(1)
# ==============================================================================

CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
GRAY='\033[1;30m'
NC='\033[0m'

cleanup() {
    echo -e "\n${YELLOW}🛑 Encerrando Cluster Mesh...${NC}"
    kill $MOCK_PID 2>/dev/null
    kill $N1_PID 2>/dev/null
    kill $N2_PID 2>/dev/null
    kill $N3_PID 2>/dev/null
    exit 0
}
trap cleanup SIGINT SIGTERM EXIT

echo -e "${CYAN}╔══════════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║     🌐 CROM-LAB | Cluster P2P Gossip e Sincronia de Bypasses ║${NC}"
echo -e "${CYAN}╚══════════════════════════════════════════════════════════════╝${NC}\n"

echo -e "⚡ Injetando Mock Llama Backend na Porta :9090..."
go run scripts/mock_llm.go > /dev/null 2>&1 &
MOCK_PID=$!
sleep 1

# Topology: 
# Node01 conhece Node02
# Node02 conhece Node03
# Node03 conhece Node01

echo -e "⚡ Subindo Node01 (:8081) [Peer: 8082]"
./bin/sinapse serve -p 8081 -l http://127.0.0.1:9090 -n NO-01 --peers http://127.0.0.1:8082 > logs_node1.txt 2>&1 &
N1_PID=$!

echo -e "⚡ Subindo Node02 (:8082) [Peer: 8083]"
./bin/sinapse serve -p 8082 -l http://127.0.0.1:9090 -n NO-02 --peers http://127.0.0.1:8083 > logs_node2.txt 2>&1 &
N2_PID=$!

echo -e "⚡ Subindo Node03 (:8083) [Peer: 8081]"
./bin/sinapse serve -p 8083 -l http://127.0.0.1:9090 -n NO-03 --peers http://127.0.0.1:8081 > logs_node3.txt 2>&1 &
N3_PID=$!

sleep 3
echo -e "\n${GREEN}🚀 O Cluster de Inteligência Distribuída está operante!${NC}\n"

PROMPT='{"prompt": "Tese Distribuida: Fragmentos O(1) roteam pela malha."}'

# 1. Bate no Node 1
echo -e "${GRAY}[Ação] Cliente envia Prompt complexo para o Node01 ...${NC}"
res1=$(curl -s -X POST http://127.0.0.1:8081/v1/chat/completions -H "Content-Type: application/json" -d "$PROMPT")
tt1=$(echo "$res1" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data.get('ttft_ms', '0'))" 2>/dev/null)
by1=$(echo "$res1" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data.get('chunks_bypassed', '0'))" 2>/dev/null)

echo -e "  → Node01 TTFT: ${tt1}ms | Bypassed Chunks: ${by1} (Miss Cache local)"

# Dá um tempinho para o protocolo Epidêmico (Gossip) rodar pelos 3 nós
echo -e "\n${GRAY}[Ação] Aguardando 1 segundo para a malha Gossip propagar Delta Vectors via HTTP...${NC}"
sleep 1 

# 2. Bate no Node 3 (Que está 2 saltos longe do Node 1) e ve se ele se beneficiou do bypass
echo -e "\n${GRAY}[Ação] OUTRO Cliente (no Japão) envia O MESMO prompt conectando-se ao Node03 ...${NC}"
res3=$(curl -s -X POST http://127.0.0.1:8083/v1/chat/completions -H "Content-Type: application/json" -d "$PROMPT")
tt3=$(echo "$res3" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data.get('ttft_ms', '0'))" 2>/dev/null)
by3=$(echo "$res3" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data.get('chunks_bypassed', '0'))" 2>/dev/null)

if [[ "$by3" -gt "0" ]]; then
    echo -e "  → Node03 TTFT: ${tt3}ms | Bypassed Chunks: ${by3} ${GREEN}(HIT GOSSIP! O NO-03 Bypasou a IA usando a neuro divergência calculada no NO-01!)${NC}"
else
    echo -e "  → Node03 TTFT: ${tt3}ms | Bypassed: ${by3} ${YELLOW}(Falhou sincronia)${NC}"
    echo "Logs Node03:"
    cat logs_node3.txt
fi

echo -e "\n${CYAN}🎯 Laboratório Mesh Finalizado.${NC}"
