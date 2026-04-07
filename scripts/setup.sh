#!/bin/bash
# ══════════════════════════════════════════════════════════
# Crompressor-Sinapse — Setup
# Verifica ambiente, instala dependências, compila binário.
# ══════════════════════════════════════════════════════════
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
BIN_DIR="$PROJECT_DIR/bin"
REPORTS_DIR="$PROJECT_DIR/reports"

echo "╔══════════════════════════════════════════════╗"
echo "║   Crompressor-Sinapse — Setup                ║"
echo "╚══════════════════════════════════════════════╝"
echo ""

# ── 1. Verificar Go ──
echo "▶ Verificando Go..."
if ! command -v go &>/dev/null; then
    echo "❌ Go não encontrado. Instale Go 1.22+ primeiro."
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}')
echo "  ✅ $GO_VERSION"

# ── 2. Verificar diretórios ──
echo "▶ Criando diretórios..."
mkdir -p "$BIN_DIR" "$REPORTS_DIR"
echo "  ✅ bin/ e reports/ criados"

# ── 3. Dependências Go ──
echo "▶ Resolvendo dependências..."
cd "$PROJECT_DIR"
go mod tidy 2>&1 | head -5
echo "  ✅ go mod tidy concluído"

# ── 4. Compilar binário ──
echo "▶ Compilando binário..."
go build -v -o "$BIN_DIR/sinapse" ./cmd/sinapse/ 2>&1 | tail -3
echo "  ✅ $BIN_DIR/sinapse compilado"

# ── 5. Verificar binário ──
echo "▶ Verificando binário..."
"$BIN_DIR/sinapse" --help >/dev/null 2>&1
echo "  ✅ Binário executa corretamente"

# ── Resumo ──
echo ""
echo "═══════════════════════════════════════════════"
echo "  ✅ Setup concluído com sucesso!"
echo ""
echo "  Binário: $BIN_DIR/sinapse"
echo "  Go:      $GO_VERSION"
echo "  OS:      $(uname -s)/$(uname -m)"
echo "═══════════════════════════════════════════════"
