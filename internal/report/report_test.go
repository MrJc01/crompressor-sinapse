package report

import (
	"strings"
	"testing"
	"time"

	"github.com/MrJc01/crompressor-sinapse/internal/analyzer"
	"github.com/MrJc01/crompressor-sinapse/internal/cdc"
	"github.com/MrJc01/crompressor-sinapse/internal/codebook"
)

func sampleResult() analyzer.DirResult {
	return analyzer.DirResult{
		Path:         "./testdata/samples",
		TotalFiles:   3,
		TotalBytes:   15360,
		TotalChunks:  120,
		UniqueChunks: 80,
		GlobalDedup:  0.333,
		AvgChunkSize: 128,
		CDCOptions:   cdc.DefaultOptions(),
		Duration:     42 * time.Millisecond,
		CodebookStats: codebook.Stats{
			TotalEntries: 80,
			TotalHits:    40,
			TotalMisses:  80,
			HitRate:      0.333,
		},
		Files: []analyzer.FileResult{
			{Name: "go_code.go", SizeBytes: 5120, TotalChunks: 40, UniqueChunks: 35, DedupRate: 0.125, AvgChunkSize: 128},
			{Name: "json_log.json", SizeBytes: 5120, TotalChunks: 40, UniqueChunks: 25, DedupRate: 0.375, AvgChunkSize: 128},
			{Name: "portuguese.txt", SizeBytes: 5120, TotalChunks: 40, UniqueChunks: 30, DedupRate: 0.25, AvgChunkSize: 128},
		},
	}
}

func TestFormatTerminal(t *testing.T) {
	r := sampleResult()
	output := FormatTerminal(r)

	if !strings.Contains(output, "SINAPSE CDC ANALYSIS REPORT") {
		t.Error("Terminal output deveria conter título")
	}
	if !strings.Contains(output, "120") {
		t.Error("Terminal output deveria conter total de chunks")
	}

	t.Logf("Terminal output:\n%s", output)
}

func TestFormatMarkdown(t *testing.T) {
	r := sampleResult()
	output := FormatMarkdown(r)

	if !strings.Contains(output, "# Crompressor-Sinapse") {
		t.Error("Markdown deveria conter header H1")
	}
	if !strings.Contains(output, "go_code.go") {
		t.Error("Markdown deveria listar arquivos")
	}
	if !strings.Contains(output, "Codebook") {
		t.Error("Markdown deveria ter seção Codebook")
	}

	t.Logf("Markdown length: %d chars", len(output))
}

func TestFormatJSON(t *testing.T) {
	r := sampleResult()
	output, err := FormatJSON(r)
	if err != nil {
		t.Fatalf("FormatJSON falhou: %v", err)
	}

	if !strings.Contains(output, "\"TotalChunks\": 120") {
		t.Error("JSON deveria conter TotalChunks")
	}

	t.Logf("JSON length: %d chars", len(output))
}
