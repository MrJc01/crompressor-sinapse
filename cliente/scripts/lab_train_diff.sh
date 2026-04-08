#!/bin/bash

# ==============================================================================
# Crompressor-Sinapse | Lab: Evolução XOR Esparsa de VRAM
# ==============================================================================

# Cores para Terminal
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}╔══════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║      🔬 CROM-LAB | Curva de Sparsity e Compressão O(1)       ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════════════════╝${NC}"
echo ""

make_run() {
    local b=$1
    local s=$2
    
    echo -n -e "▶ Simulando ${YELLOW}[Codebook: $b Blocos, Limiar $s]${NC} ... "
    # Extrai as saídas vitais processando o texto via AWK pra montar a tabela em tempo real.
    local saida=$(./bin/sinapse train -b $b -s $s)
    local vram_save=$(echo "$saida" | grep "Taxa de Sparsity" | awk '{print $NF}')
    local time_cost=$(echo "$saida" | grep "Treino Finalizado em" | awk '{print $NF}')
    
    # Se bater erro ou timeout a string pode vir suja.
    if [[ -z "$vram_save" ]]; then
        vram_save="FAIL%"
    fi

    echo -e "${GREEN}Concluído em $time_cost  =>  VRAM Economizada (XOR): $vram_save${NC}"
}

echo -e "Analisando comportamento sobre Blocos Estáticos de IA..."
echo "Avaliando cenários de Micro-Skills (Apenas ~2% dos pesos afetados) e Macro-Skills (~35%)."
echo ""

make_run 128 0.02
make_run 128 0.15
make_run 128 0.35
echo "---"
make_run 512 0.02
make_run 512 0.15
make_run 512 0.35
echo "---"
make_run 2048 0.02
make_run 2048 0.15
make_run 2048 0.35

echo ""
echo -e "${GREEN}🎯 O Laboratório Terminou.${NC} Observe que independentemente do tamanho massivo da IA, a retenção restrita ao escopo (threshold) salva proporcionalmente a VRAM do ambiente GGUF."
