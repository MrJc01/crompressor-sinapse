#!/bin/bash
# ══════════════════════════════════════════════════════════
# Crompressor-Sinapse — Test
# Roda todos os testes com race detector e gera relatório.
# ══════════════════════════════════════════════════════════
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
REPORTS_DIR="$PROJECT_DIR/reports"
TIMESTAMP=$(date +%Y-%m-%d_%H-%M-%S)
REPORT_FILE="$REPORTS_DIR/test_${TIMESTAMP}.md"
COVERAGE_FILE="$REPORTS_DIR/coverage.out"

mkdir -p "$REPORTS_DIR"

echo "╔══════════════════════════════════════════════╗"
echo "║   Crompressor-Sinapse — Test Suite           ║"
echo "╚══════════════════════════════════════════════╝"
echo ""

cd "$PROJECT_DIR"

# ── 1. Rodar testes ──
echo "▶ Rodando testes com race detector..."
TEST_OUTPUT=$(go test -v -race -cover -coverprofile="$COVERAGE_FILE" ./... 2>&1) || true
echo "$TEST_OUTPUT"

# ── 2. Extrair métricas ──
TOTAL_TESTS=$(echo "$TEST_OUTPUT" | grep -c "^--- " || true)
PASSED=$(echo "$TEST_OUTPUT" | grep -c "^--- PASS" || true)
FAILED=$(echo "$TEST_OUTPUT" | grep -c "^--- FAIL" || true)
COVERAGE=$(echo "$TEST_OUTPUT" | grep -oP 'coverage: \K[0-9.]+' | tail -1 || echo "0")

# ── 3. Gerar relatório MD ──
cat > "$REPORT_FILE" <<EOF
# Crompressor-Sinapse — Relatório de Testes

> Gerado em: $(date '+%Y-%m-%d %H:%M:%S')

## Resumo

| Métrica | Valor |
|---------|-------|
| **Total de Testes** | $TOTAL_TESTS |
| **Passou** | ✅ $PASSED |
| **Falhou** | ❌ $FAILED |
| **Cobertura** | ${COVERAGE}% |
| **Race Detector** | Ativo |
| **Go** | $(go version | awk '{print $3}') |

## Output Completo

\`\`\`
$TEST_OUTPUT
\`\`\`

---

> *Crompressor-Sinapse — "Quando a compressão se torna cognição."*
EOF

echo ""
echo "═══════════════════════════════════════════════"
echo "  Testes: $PASSED passou, $FAILED falhou"
echo "  Cobertura: ${COVERAGE}%"
echo "  Relatório: $REPORT_FILE"
echo "═══════════════════════════════════════════════"
