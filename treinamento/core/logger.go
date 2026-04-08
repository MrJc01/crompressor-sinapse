package core

import (
	"log/slog"
	"os"
)

// Logger central para auditoria forense do Crompressor.
// Foco em ser de altíssima performance para não criar gargalos no treinamento O(1).
var Logger *slog.Logger

func init() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	Logger = slog.New(handler)
}

// LogForensic padroniza a saída do engenheiro SRE e arquiteto de ML
// estruturando o ponto exato da arquitetura em que ocorre o evento.
func LogForensic(layer, msg string, args ...any) {
	Logger.Debug("["+layer+"] "+msg, args...)
}
