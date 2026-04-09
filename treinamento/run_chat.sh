#!/bin/bash
cd /home/j/Documentos/GitHub/crompressor-sinapse/treinamento

echo "[*] Limpeza Ambiental Audivel (Auditando Ghost Processes)..."
netstat -tulnp 2>/dev/null | grep -i go || echo "Servidores base desativados. Rede P2P livre."

echo "[*] Compilando Engine Chatbot LLM Generativo..."
mkdir -p ./bin
go build -o ./bin/conversacional ./cmd/conversacional

echo "[*] Inicializando Neural Shell O(1)..."
./bin/conversacional
