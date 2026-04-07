#!/bin/bash
# ══════════════════════════════════════════════════════════
# Crompressor-Sinapse — CI Pipeline
# Pipeline completo: setup → test → bench → analyze → report
# Execute com: bash scripts/ci.sh
# ══════════════════════════════════════════════════════════
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
START_TIME=$(date +%s)

echo ""
echo "╔══════════════════════════════════════════════════╗"
echo "║  🧠 Crompressor-Sinapse — CI Pipeline Completo   ║"
echo "║     \"Quando a compressão se torna cognição.\"     ║"
echo "╚══════════════════════════════════════════════════╝"
echo ""

# ── Etapa 1: Setup ──
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  📦 ETAPA 1/5: Setup"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
bash "$SCRIPT_DIR/setup.sh"
echo ""

# ── Etapa 2: Testes ──
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  🧪 ETAPA 2/5: Testes"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
bash "$SCRIPT_DIR/test.sh"
echo ""

# ── Etapa 3: Benchmarks ──
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  ⚡ ETAPA 3/5: Benchmarks"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
bash "$SCRIPT_DIR/bench.sh"
echo ""

# ── Etapa 4: Análise CDC ──
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  🔬 ETAPA 4/5: Análise CDC"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
bash "$SCRIPT_DIR/analyze.sh"
echo ""

# ── Etapa 5: Relatório Consolidado ──
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  📊 ETAPA 5/5: Relatório Final"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
bash "$SCRIPT_DIR/report.sh"
echo ""

# ── Resumo Final ──
END_TIME=$(date +%s)
ELAPSED=$((END_TIME - START_TIME))

echo ""
echo "╔══════════════════════════════════════════════════╗"
echo "║  ✅ CI PIPELINE CONCLUÍDO COM SUCESSO            ║"
echo "╠══════════════════════════════════════════════════╣"
echo "║                                                  ║"
echo "║  Tempo total: ${ELAPSED}s                              ║"
echo "║                                                  ║"
echo "║  Relatórios gerados em: ./reports/               ║"
echo "║                                                  ║"
echo "╚══════════════════════════════════════════════════╝"
echo ""

# Listar relatórios gerados
echo "📁 Relatórios disponíveis:"
ls -la "$(dirname "$SCRIPT_DIR")/reports/"*.md 2>/dev/null | awk '{print "  " $NF}' || echo "  Nenhum"
echo ""
