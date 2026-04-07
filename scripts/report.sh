#!/bin/bash
# ══════════════════════════════════════════════════════════
# Crompressor-Sinapse — Report
# Consolida todos os relatórios em um relatório final.
# ══════════════════════════════════════════════════════════
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
REPORTS_DIR="$PROJECT_DIR/reports"
TIMESTAMP=$(date +%Y-%m-%d_%H-%M-%S)
FULL_REPORT="$REPORTS_DIR/full_report_${TIMESTAMP}.md"

mkdir -p "$REPORTS_DIR"

echo "╔══════════════════════════════════════════════╗"
echo "║   Crompressor-Sinapse — Full Report          ║"
echo "╚══════════════════════════════════════════════╝"
echo ""

# ── Gerar relatório consolidado ──
cat > "$FULL_REPORT" <<HEADER
# 📊 Crompressor-Sinapse — Relatório Consolidado

> Gerado em: $(date '+%Y-%m-%d %H:%M:%S')
> Máquina: $(uname -n) ($(uname -sr), $(uname -m))

---

## Ambiente

| Parâmetro | Valor |
|-----------|-------|
| **CPU** | $(grep "model name" /proc/cpuinfo 2>/dev/null | head -1 | cut -d: -f2 | xargs || echo "N/A") |
| **RAM** | $(free -h 2>/dev/null | awk '/Mem:/{print $2}' || echo "N/A") |
| **Disco Livre** | $(df -h "$PROJECT_DIR" 2>/dev/null | awk 'NR==2{print $4}' || echo "N/A") |
| **Go** | $(go version 2>/dev/null | awk '{print $3}' || echo "N/A") |
| **OS** | $(uname -sr) |

---

HEADER

# Encontrar relatórios mais recentes de cada tipo
LATEST_TEST=$(ls -t "$REPORTS_DIR"/test_*.md 2>/dev/null | head -1 || true)
LATEST_BENCH=$(ls -t "$REPORTS_DIR"/bench_*.md 2>/dev/null | head -1 || true)
LATEST_ANALYSIS=$(ls -t "$REPORTS_DIR"/analysis_*.md 2>/dev/null | head -1 || true)

# Incluir relatório de teste
if [ -n "$LATEST_TEST" ] && [ -f "$LATEST_TEST" ]; then
    echo "  ✅ Incluindo: $(basename "$LATEST_TEST")"
    echo "## Testes" >> "$FULL_REPORT"
    echo "" >> "$FULL_REPORT"
    # Extrair conteúdo sem o header H1
    sed '1,/^---$/d' "$LATEST_TEST" >> "$FULL_REPORT" 2>/dev/null || cat "$LATEST_TEST" >> "$FULL_REPORT"
    echo -e "\n---\n" >> "$FULL_REPORT"
else
    echo "  ⚠️  Nenhum relatório de teste encontrado"
fi

# Incluir relatório de benchmark
if [ -n "$LATEST_BENCH" ] && [ -f "$LATEST_BENCH" ]; then
    echo "  ✅ Incluindo: $(basename "$LATEST_BENCH")"
    echo "## Benchmarks" >> "$FULL_REPORT"
    echo "" >> "$FULL_REPORT"
    sed '1,/^---$/d' "$LATEST_BENCH" >> "$FULL_REPORT" 2>/dev/null || cat "$LATEST_BENCH" >> "$FULL_REPORT"
    echo -e "\n---\n" >> "$FULL_REPORT"
else
    echo "  ⚠️  Nenhum relatório de benchmark encontrado"
fi

# Incluir relatório de análise
if [ -n "$LATEST_ANALYSIS" ] && [ -f "$LATEST_ANALYSIS" ]; then
    echo "  ✅ Incluindo: $(basename "$LATEST_ANALYSIS")"
    echo "## Análise CDC" >> "$FULL_REPORT"
    echo "" >> "$FULL_REPORT"
    sed '1,/^---$/d' "$LATEST_ANALYSIS" >> "$FULL_REPORT" 2>/dev/null || cat "$LATEST_ANALYSIS" >> "$FULL_REPORT"
    echo -e "\n---\n" >> "$FULL_REPORT"
else
    echo "  ⚠️  Nenhum relatório de análise encontrado"
fi

# Footer
cat >> "$FULL_REPORT" <<FOOTER

---

> *Crompressor-Sinapse — "Quando a compressão se torna cognição."*
FOOTER

echo ""
echo "═══════════════════════════════════════════════"
echo "  ✅ Relatório consolidado gerado"
echo "  Arquivo: $FULL_REPORT"
echo "═══════════════════════════════════════════════"
