# Crompressor-Sinapse Monorepo Makefile

.PHONY: all build-cliente test-cliente clean

all: build-cliente

build-cliente:
	@echo "Construindo o Gateway Daemon Cliente O(1)..."
	@cd cliente && go build -o bin/sinapse ./cmd/sinapse
	@echo "Build Cliente Concluída!"

test-cliente:
	@echo "Rodando a Bateria SRE de Testes Concorrentes no Core (Fases 1-6)..."
	@cd cliente && go test -v -race ./...

clean:
	@echo "Limpando artefatos de build..."
	@rm -rf cliente/bin/sinapse
	@echo "Limpeza concluída."
