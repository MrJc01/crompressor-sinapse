// Package analyzer coleta métricas de análise CDC sobre arquivos e diretórios.
package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/MrJc01/crompressor-sinapse/internal/cdc"
	"github.com/MrJc01/crompressor-sinapse/internal/codebook"
)

// FileResult contém o resultado da análise de um arquivo.
type FileResult struct {
	Name         string
	SizeBytes    int64
	TotalChunks  int
	UniqueChunks int
	DedupRate    float64
	AvgChunkSize float64
	MinChunkSize uint32
	MaxChunkSize uint32
}

// DirResult contém o resultado consolidado da análise de um diretório.
type DirResult struct {
	Path          string
	Files         []FileResult
	TotalFiles    int
	TotalBytes    int64
	TotalChunks   int
	UniqueChunks  int
	GlobalDedup   float64
	AvgChunkSize  float64
	CodebookStats codebook.Stats
	Duration      time.Duration
	CDCOptions    cdc.Options
}

// AnalyzeFile analisa um único arquivo com CDC e retorna métricas.
func AnalyzeFile(path string, opts cdc.Options) (FileResult, error) {
	info, err := os.Stat(path)
	if err != nil {
		return FileResult{}, err
	}

	chunks, err := cdc.TokenizeFile(path, opts)
	if err != nil {
		return FileResult{}, err
	}

	stats := cdc.StatsFor(chunks)

	return FileResult{
		Name:         filepath.Base(path),
		SizeBytes:    info.Size(),
		TotalChunks:  stats.TotalChunks,
		UniqueChunks: stats.UniqueChunks,
		DedupRate:    stats.DedupRate,
		AvgChunkSize: stats.AvgChunkSize,
		MinChunkSize: stats.MinChunkSize,
		MaxChunkSize: stats.MaxChunkSize,
	}, nil
}

// AnalyzeDir analisa todos os arquivos de um diretório com um Codebook compartilhado.
// O Codebook compartilhado permite medir deduplicação CROSS-FILE.
func AnalyzeDir(dir string, opts cdc.Options) (DirResult, error) {
	start := time.Now()

	entries, err := os.ReadDir(dir)
	if err != nil {
		return DirResult{}, fmt.Errorf("erro ao ler diretório: %w", err)
	}

	cb := codebook.New()
	result := DirResult{
		Path:       dir,
		CDCOptions: opts,
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		path := filepath.Join(dir, entry.Name())

		// Analisar arquivo individualmente
		fr, err := AnalyzeFile(path, opts)
		if err != nil {
			continue
		}

		// Registrar chunks no Codebook global (cross-file dedup)
		chunks, err := cdc.TokenizeFile(path, opts)
		if err != nil {
			continue
		}
		cb.InsertAll(chunks)

		result.Files = append(result.Files, fr)
		result.TotalFiles++
		result.TotalBytes += fr.SizeBytes
		result.TotalChunks += fr.TotalChunks
	}

	// Métricas globais (cross-file)
	cbStats := cb.Stats()
	result.UniqueChunks = cb.Len()
	result.CodebookStats = cbStats
	result.Duration = time.Since(start)

	if result.TotalChunks > 0 {
		result.GlobalDedup = 1.0 - float64(result.UniqueChunks)/float64(result.TotalChunks)
		result.AvgChunkSize = float64(result.TotalBytes) / float64(result.TotalChunks)
	}

	return result, nil
}
