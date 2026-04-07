.PHONY: build test bench analyze report ci clean help

# ── Variáveis ──
BIN       := ./bin/sinapse
GOFLAGS   := -v
REPORTS   := ./reports
TIMESTAMP := $(shell date +%Y-%m-%d_%H-%M-%S)

# ── Targets Principais ──

help: ## Mostra esta ajuda
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Compila o binário CLI
	@mkdir -p bin
	go build $(GOFLAGS) -o $(BIN) ./cmd/sinapse/

test: ## Roda todos os testes com race detector
	go test -v -race -cover -coverprofile=$(REPORTS)/coverage.out ./...

bench: build ## Roda benchmarks e gera relatório
	go test -bench=. -benchmem -benchtime=5s ./... | tee $(REPORTS)/bench_raw_$(TIMESTAMP).txt

analyze: build ## Analisa testdata e gera relatório
	$(BIN) analyze --input ./testdata/samples/ --output $(REPORTS)/analysis_$(TIMESTAMP).md

report: ## Gera relatório consolidado
	$(BIN) report --reports-dir $(REPORTS) --output $(REPORTS)/full_report_$(TIMESTAMP).md

ci: ## Pipeline completo: build → test → bench → analyze → report
	@bash scripts/ci.sh

clean: ## Remove binários e relatórios
	rm -rf bin/
	rm -f $(REPORTS)/*.out $(REPORTS)/*_raw_*.txt

# ── Atalhos ──

setup: ## Configura ambiente
	@bash scripts/setup.sh

run: build ## Compila e roda help
	$(BIN) --help
