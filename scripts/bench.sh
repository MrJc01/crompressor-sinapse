#!/bin/bash
# ══════════════════════════════════════════════════════════
# Crompressor-Sinapse — Benchmark
# Roda benchmarks de performance CDC e gera relatório.
# ══════════════════════════════════════════════════════════
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
REPORTS_DIR="$PROJECT_DIR/reports"
TIMESTAMP=$(date +%Y-%m-%d_%H-%M-%S)
REPORT_FILE="$REPORTS_DIR/bench_${TIMESTAMP}.md"
RAW_FILE="$REPORTS_DIR/bench_raw_${TIMESTAMP}.txt"

mkdir -p "$REPORTS_DIR"

echo "╔══════════════════════════════════════════════╗"
echo "║   Crompressor-Sinapse — Benchmark            ║"
echo "╚══════════════════════════════════════════════╝"
echo ""

cd "$PROJECT_DIR"

# ── 1. Rodar benchmarks ──
echo "▶ Rodando benchmarks (5s por teste)..."
BENCH_OUTPUT=$(go test -bench=. -benchmem -benchtime=5s ./... 2>&1) || true
echo "$BENCH_OUTPUT" | tee "$RAW_FILE"

# ── 2. Gerar relatório MD ──
cat > "$REPORT_FILE" <<EOF
# Crompressor-Sinapse — Relatório de Benchmark

> Gerado em: $(date '+%Y-%m-%d %H:%M:%S')

## Ambiente

| Parâmetro | Valor |
|-----------|-------|
| **CPU** | $(grep "model name" /proc/cpuinfo 2>/dev/null | head -1 | cut -d: -f2 | xargs || echo "N/A") |
| **RAM** | $(free -h 2>/dev/null | awk '/Mem:/{print $2}' || echo "N/A") |
| **Go** | $(go version | awk '{print $3}') |
| **OS** | $(uname -sr) |
| **Arch** | $(uname -m) |

## Resultados

\`\`\`
$BENCH_OUTPUT
\`\`\`

## Análise

Os benchmarks medem:
- **Tokenize_1KB**: Throughput do CDC para blocos pequenos
- **Tokenize_100KB**: Throughput do CDC para blocos grandes
- **RabinRoll**: Velocidade do rolling hash por byte
- **HashBytes**: Velocidade do hash FNV para chunks
- **Codebook_Insert**: Velocidade de inserção no Codebook

---

> *Crompressor-Sinapse — "Nós não comprimimos dados. Nós indexamos o universo."*
EOF

echo ""
echo "═══════════════════════════════════════════════"
echo "  Raw: $RAW_FILE"
echo "  Relatório: $REPORT_FILE"
echo "═══════════════════════════════════════════════"
