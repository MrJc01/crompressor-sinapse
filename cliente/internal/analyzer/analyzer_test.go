package analyzer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/MrJc01/crompressor-sinapse/internal/cdc"
)

func createTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestAnalyzeFile(t *testing.T) {
	dir := t.TempDir()
	path := createTempFile(t, dir, "test.go", `package main

import "fmt"

func main() {
	fmt.Println("Crompressor-Sinapse: análise de arquivo")
	for i := 0; i < 100; i++ {
		fmt.Printf("iteração %d\n", i)
	}
}
`)

	result, err := AnalyzeFile(path, cdc.DefaultOptions())
	if err != nil {
		t.Fatalf("AnalyzeFile falhou: %v", err)
	}

	if result.TotalChunks == 0 {
		t.Error("TotalChunks não deveria ser zero")
	}

	t.Logf("Arquivo: %s, %d bytes, %d chunks, %.1f%% dedup, avg=%.0f bytes",
		result.Name, result.SizeBytes, result.TotalChunks,
		result.DedupRate*100, result.AvgChunkSize)
}

func TestAnalyzeDir(t *testing.T) {
	dir := t.TempDir()

	// Criar arquivos com conteúdo parcialmente repetido (simular dedup cross-file)
	createTempFile(t, dir, "main.go", `package main

import "fmt"

func main() {
	fmt.Println("Hello from main")
}
`)
	createTempFile(t, dir, "helper.go", `package main

import "fmt"

func helper() {
	fmt.Println("Hello from helper")
}
`)
	createTempFile(t, dir, "data.json", `{
	"name": "crompressor-sinapse",
	"version": "1.0.0",
	"description": "CDC tokenization engine"
}
`)

	result, err := AnalyzeDir(dir, cdc.Options{MinSize: 16, MaxSize: 128, AvgSize: 32})
	if err != nil {
		t.Fatalf("AnalyzeDir falhou: %v", err)
	}

	if result.TotalFiles != 3 {
		t.Errorf("Esperado 3 arquivos, obtido %d", result.TotalFiles)
	}

	t.Logf("Dir: %d arquivos, %d bytes, %d chunks (%d únicos), dedup=%.1f%%, tempo=%v",
		result.TotalFiles, result.TotalBytes, result.TotalChunks,
		result.UniqueChunks, result.GlobalDedup*100, result.Duration)
}
