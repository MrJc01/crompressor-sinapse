// Crompressor-Sinapse CLI
// Binário unificado com subcomandos: tokenize, analyze, report
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/MrJc01/crompressor-sinapse/internal/analyzer"
	"github.com/MrJc01/crompressor-sinapse/internal/cdc"
	"github.com/MrJc01/crompressor-sinapse/internal/report"
)

var (
	// Flags globais
	verbose bool
	output  string
	format  string

	// Flags CDC
	minSize int
	maxSize int
	avgSize int
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "sinapse",
		Short: "Crompressor-Sinapse — Motor de Tokenização Semântica CDC",
		Long: `Crompressor-Sinapse aplica Content-Defined Chunking (CDC)
para tokenização semântica de texto e código-fonte.

"Quando a compressão se torna cognição."`,
	}

	// Flags globais
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Saída detalhada")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "", "Arquivo de saída (default: stdout)")
	rootCmd.PersistentFlags().StringVarP(&format, "format", "f", "terminal", "Formato: terminal, markdown, json")

	// Subcomandos
	rootCmd.AddCommand(tokenizeCmd())
	rootCmd.AddCommand(analyzeCmd())
	rootCmd.AddCommand(reportCmd())
	rootCmd.AddCommand(inferCmd())
	rootCmd.AddCommand(trainCmd())
	rootCmd.AddCommand(serveCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// ── Subcomando: tokenize ──

func tokenizeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tokenize",
		Short: "Tokeniza arquivo ou diretório com CDC Rabin",
		RunE: func(cmd *cobra.Command, args []string) error {
			input, _ := cmd.Flags().GetString("input")
			if input == "" {
				return fmt.Errorf("--input é obrigatório")
			}

			opts := cdcOpts()

			info, err := os.Stat(input)
			if err != nil {
				return fmt.Errorf("erro ao acessar %s: %w", input, err)
			}

			if info.IsDir() {
				results, err := cdc.TokenizeDir(input, opts)
				if err != nil {
					return err
				}
				for name, chunks := range results {
					stats := cdc.StatsFor(chunks)
					fmt.Printf("%-30s  %5d chunks  %5d únicos  dedup=%.1f%%\n",
						name, stats.TotalChunks, stats.UniqueChunks, stats.DedupRate*100)
				}
			} else {
				chunks, err := cdc.TokenizeFile(input, opts)
				if err != nil {
					return err
				}
				stats := cdc.StatsFor(chunks)
				fmt.Printf("Arquivo: %s\n", input)
				fmt.Printf("Chunks:  %d (únicos: %d, dedup: %.1f%%)\n",
					stats.TotalChunks, stats.UniqueChunks, stats.DedupRate*100)
				fmt.Printf("Avg:     %.0f bytes, Min: %d, Max: %d\n",
					stats.AvgChunkSize, stats.MinChunkSize, stats.MaxChunkSize)
			}

			return nil
		},
	}

	cmd.Flags().String("input", "", "Arquivo ou diretório de entrada")
	cmd.Flags().IntVar(&minSize, "min-size", 32, "Tamanho mínimo do chunk")
	cmd.Flags().IntVar(&maxSize, "max-size", 1024, "Tamanho máximo do chunk")
	cmd.Flags().IntVar(&avgSize, "avg-size", 128, "Tamanho médio desejado")

	return cmd
}

// ── Subcomando: analyze ──

func analyzeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Analisa diretório e gera relatório de métricas CDC",
		RunE: func(cmd *cobra.Command, args []string) error {
			input, _ := cmd.Flags().GetString("input")
			if input == "" {
				return fmt.Errorf("--input é obrigatório")
			}

			opts := cdcOpts()
			result, err := analyzer.AnalyzeDir(input, opts)
			if err != nil {
				return err
			}

			// Formatar output
			var out string
			switch format {
			case "markdown", "md":
				out = report.FormatMarkdown(result)
			case "json":
				out, err = report.FormatJSON(result)
				if err != nil {
					return err
				}
			default:
				out = report.FormatTerminal(result)
			}

			// Escrever
			if output != "" {
				if err := os.WriteFile(output, []byte(out), 0644); err != nil {
					return fmt.Errorf("erro ao escrever %s: %w", output, err)
				}
				fmt.Printf("✅ Relatório salvo em: %s\n", output)
			} else {
				fmt.Print(out)
			}

			return nil
		},
	}

	cmd.Flags().String("input", "", "Diretório de entrada")
	cmd.Flags().IntVar(&minSize, "min-size", 32, "Tamanho mínimo do chunk")
	cmd.Flags().IntVar(&maxSize, "max-size", 1024, "Tamanho máximo do chunk")
	cmd.Flags().IntVar(&avgSize, "avg-size", 128, "Tamanho médio desejado")

	return cmd
}

// ── Subcomando: report ──

func reportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Consolida relatórios existentes",
		RunE: func(cmd *cobra.Command, args []string) error {
			reportsDir, _ := cmd.Flags().GetString("reports-dir")
			if reportsDir == "" {
				reportsDir = "./reports"
			}

			entries, err := os.ReadDir(reportsDir)
			if err != nil {
				return fmt.Errorf("erro ao ler %s: %w", reportsDir, err)
			}

			fmt.Printf("📊 Relatórios em %s:\n\n", reportsDir)
			for _, e := range entries {
				if !e.IsDir() {
					info, _ := e.Info()
					fmt.Printf("  %-40s  %s\n", e.Name(), formatSize(info.Size()))
				}
			}

			return nil
		},
	}

	cmd.Flags().String("reports-dir", "./reports", "Diretório de relatórios")

	return cmd
}

// ── Helpers ──

func cdcOpts() cdc.Options {
	return cdc.Options{
		MinSize: minSize,
		MaxSize: maxSize,
		AvgSize: avgSize,
	}
}

func formatSize(b int64) string {
	switch {
	case b >= 1024*1024:
		return fmt.Sprintf("%.1f MB", float64(b)/(1024*1024))
	case b >= 1024:
		return fmt.Sprintf("%.1f KB", float64(b)/1024)
	default:
		return fmt.Sprintf("%d B", b)
	}
}
