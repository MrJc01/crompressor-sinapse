#!/bin/bash
# ══════════════════════════════════════════════════════════
# Crompressor-Sinapse — Analyze
# Analisa testdata com CDC e gera relatório de métricas.
# ══════════════════════════════════════════════════════════
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
BIN="$PROJECT_DIR/bin/sinapse"
REPORTS_DIR="$PROJECT_DIR/reports"
TIMESTAMP=$(date +%Y-%m-%d_%H-%M-%S)
REPORT_FILE="$REPORTS_DIR/analysis_${TIMESTAMP}.md"

mkdir -p "$REPORTS_DIR"

echo "╔══════════════════════════════════════════════╗"
echo "║   Crompressor-Sinapse — CDC Analyze          ║"
echo "╚══════════════════════════════════════════════╝"
echo ""

# ── 1. Verificar binário ──
if [ ! -f "$BIN" ]; then
    echo "▶ Binário não encontrado. Compilando..."
    cd "$PROJECT_DIR"
    mkdir -p bin
    go build -o "$BIN" ./cmd/sinapse/
fi

# ── 2. Analisar testdata (terminal) ──
echo "▶ Analisando ./testdata/samples/ ..."
echo ""
"$BIN" analyze --input "$PROJECT_DIR/testdata/samples/" --format terminal
echo ""

# ── 3. Gerar relatório MD ──
echo "▶ Gerando relatório Markdown..."
"$BIN" analyze --input "$PROJECT_DIR/testdata/samples/" --format markdown --output "$REPORT_FILE"

echo ""
echo "═══════════════════════════════════════════════"
echo "  ✅ Análise concluída"
echo "  Relatório: $REPORT_FILE"
echo "═══════════════════════════════════════════════"
