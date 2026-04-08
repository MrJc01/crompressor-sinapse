// Package report formata relatórios de análise CDC em múltiplos formatos.
package report

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/MrJc01/crompressor-sinapse/internal/analyzer"
)

// ── Formatação Terminal (Box Art) ──

// FormatTerminal formata um DirResult como box art colorida para o terminal.
func FormatTerminal(r analyzer.DirResult) string {
	var b strings.Builder
	w := 55

	line := strings.Repeat("═", w-2)
	b.WriteString(fmt.Sprintf("╔%s╗\n", line))
	b.WriteString(fmt.Sprintf("║%s║\n", center("SINAPSE CDC ANALYSIS REPORT", w-2)))
	b.WriteString(fmt.Sprintf("╠%s╣\n", line))

	b.WriteString(fmt.Sprintf("║  %-20s %28s  ║\n", "Path:", r.Path))
	b.WriteString(fmt.Sprintf("║  %-20s %28d  ║\n", "Files:", r.TotalFiles))
	b.WriteString(fmt.Sprintf("║  %-20s %24s  ║\n", "Total Size:", formatBytes(r.TotalBytes)))
	b.WriteString(fmt.Sprintf("║  %-20s %28d  ║\n", "Total Chunks:", r.TotalChunks))
	b.WriteString(fmt.Sprintf("║  %-20s %28d  ║\n", "Unique Chunks:", r.UniqueChunks))
	b.WriteString(fmt.Sprintf("║  %-20s %27.1f%%  ║\n", "Dedup Rate:", r.GlobalDedup*100))
	b.WriteString(fmt.Sprintf("║  %-20s %24.0f B  ║\n", "Avg Chunk Size:", r.AvgChunkSize))
	b.WriteString(fmt.Sprintf("║  %-20s %28v  ║\n", "Duration:", r.Duration.Round(time.Millisecond)))

	b.WriteString(fmt.Sprintf("║%s║\n", strings.Repeat(" ", w-2)))
	b.WriteString(fmt.Sprintf("║  %-20s %28s  ║\n", "CDC Strategy:", "Rabin"))
	b.WriteString(fmt.Sprintf("║  %-20s %28d  ║\n", "Min Chunk:", r.CDCOptions.MinSize))
	b.WriteString(fmt.Sprintf("║  %-20s %28d  ║\n", "Avg Chunk:", r.CDCOptions.AvgSize))
	b.WriteString(fmt.Sprintf("║  %-20s %28d  ║\n", "Max Chunk:", r.CDCOptions.MaxSize))

	b.WriteString(fmt.Sprintf("╚%s╝\n", line))

	return b.String()
}

// ── Formatação Markdown ──

// FormatMarkdown formata um DirResult como relatório Markdown completo.
func FormatMarkdown(r analyzer.DirResult) string {
	var b strings.Builder
	ts := time.Now().Format("2006-01-02 15:04:05")

	b.WriteString("# Crompressor-Sinapse — Relatório de Análise CDC\n\n")
	b.WriteString(fmt.Sprintf("> Gerado em: %s\n\n", ts))
	b.WriteString("---\n\n")

	// Resumo
	b.WriteString("## Resumo\n\n")
	b.WriteString("| Métrica | Valor |\n")
	b.WriteString("|---------|-------|\n")
	b.WriteString(fmt.Sprintf("| **Path** | `%s` |\n", r.Path))
	b.WriteString(fmt.Sprintf("| **Arquivos** | %d |\n", r.TotalFiles))
	b.WriteString(fmt.Sprintf("| **Tamanho Total** | %s |\n", formatBytes(r.TotalBytes)))
	b.WriteString(fmt.Sprintf("| **Total Chunks** | %d |\n", r.TotalChunks))
	b.WriteString(fmt.Sprintf("| **Chunks Únicos** | %d |\n", r.UniqueChunks))
	b.WriteString(fmt.Sprintf("| **Taxa de Dedup** | %.1f%% |\n", r.GlobalDedup*100))
	b.WriteString(fmt.Sprintf("| **Tamanho Médio Chunk** | %.0f bytes |\n", r.AvgChunkSize))
	b.WriteString(fmt.Sprintf("| **Duração** | %v |\n", r.Duration.Round(time.Millisecond)))
	b.WriteString("\n")

	// Config CDC
	b.WriteString("## Configuração CDC\n\n")
	b.WriteString("| Parâmetro | Valor |\n")
	b.WriteString("|-----------|-------|\n")
	b.WriteString(fmt.Sprintf("| Estratégia | Rabin Rolling Hash |\n"))
	b.WriteString(fmt.Sprintf("| Min Size | %d bytes |\n", r.CDCOptions.MinSize))
	b.WriteString(fmt.Sprintf("| Avg Size | %d bytes |\n", r.CDCOptions.AvgSize))
	b.WriteString(fmt.Sprintf("| Max Size | %d bytes |\n", r.CDCOptions.MaxSize))
	b.WriteString("\n")

	// Por arquivo
	if len(r.Files) > 0 {
		b.WriteString("## Detalhamento por Arquivo\n\n")
		b.WriteString("| Arquivo | Tamanho | Chunks | Únicos | Dedup | Avg Size |\n")
		b.WriteString("|---------|---------|--------|--------|-------|----------|\n")
		for _, f := range r.Files {
			b.WriteString(fmt.Sprintf("| `%s` | %s | %d | %d | %.1f%% | %.0f B |\n",
				f.Name, formatBytes(f.SizeBytes), f.TotalChunks,
				f.UniqueChunks, f.DedupRate*100, f.AvgChunkSize))
		}
		b.WriteString("\n")
	}

	// Codebook
	b.WriteString("## Codebook (Cross-File)\n\n")
	b.WriteString("| Métrica | Valor |\n")
	b.WriteString("|---------|-------|\n")
	b.WriteString(fmt.Sprintf("| Entradas | %d |\n", r.CodebookStats.TotalEntries))
	b.WriteString(fmt.Sprintf("| Hits | %d |\n", r.CodebookStats.TotalHits))
	b.WriteString(fmt.Sprintf("| Misses | %d |\n", r.CodebookStats.TotalMisses))
	b.WriteString(fmt.Sprintf("| Hit Rate | %.1f%% |\n", r.CodebookStats.HitRate*100))
	b.WriteString("\n")

	// Info do sistema
	b.WriteString("## Ambiente\n\n")
	b.WriteString("| Parâmetro | Valor |\n")
	b.WriteString("|-----------|-------|\n")
	b.WriteString(fmt.Sprintf("| Go | %s |\n", runtime.Version()))
	b.WriteString(fmt.Sprintf("| OS/Arch | %s/%s |\n", runtime.GOOS, runtime.GOARCH))
	b.WriteString(fmt.Sprintf("| CPUs | %d |\n", runtime.NumCPU()))
	b.WriteString("\n---\n\n")
	b.WriteString("> *\"Nós não comprimimos dados. Nós indexamos o universo.\"*\n")

	return b.String()
}

// ── Formatação JSON ──

// FormatJSON serializa um DirResult como JSON formatado.
func FormatJSON(r analyzer.DirResult) (string, error) {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ── Helpers ──

func center(s string, width int) string {
	if len(s) >= width {
		return s
	}
	pad := (width - len(s)) / 2
	return strings.Repeat(" ", pad) + s + strings.Repeat(" ", width-len(s)-pad)
}

func formatBytes(b int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case b >= GB:
		return fmt.Sprintf("%.2f GB", float64(b)/float64(GB))
	case b >= MB:
		return fmt.Sprintf("%.2f MB", float64(b)/float64(MB))
	case b >= KB:
		return fmt.Sprintf("%.2f KB", float64(b)/float64(KB))
	default:
		return fmt.Sprintf("%d B", b)
	}
}
