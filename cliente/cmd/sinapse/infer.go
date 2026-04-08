package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/MrJc01/crompressor-sinapse/internal/inference"
)

// inferCmd levanta o Simulador de Inferência acoplado ao Cache CDC.
func inferCmd() *cobra.Command {
	var iteracoes int

	cmd := &cobra.Command{
		Use:   "infer [prompt]",
		Short: "Roda uma inferência iterativa comprovando bypass neural diferencial",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			prompt := args[0]
			opts := cdcOpts() // Compartilha a configuração declarada nas flags globais instanciadas no main.go

			fmt.Println("🚀 Inicializando LLM Mock Simulator e LRU Cache...")
			cache := inference.NewActivationCache(2048)
			wrapper := inference.NewSimulatedWrapper(cache, 768)

			fmt.Printf("Turno Base (Frio) ──\n")
			fmt.Printf(" Prompt: %s\n", truncateStr(prompt, 50))
			
			// 1° Turno frio. O Computador fará todo o processamento massivo simulado
			outFrio := wrapper.ProcessPrompt(prompt, opts)
			printInferenceMetrics(outFrio, 1)

			basePrompt := prompt
			// Iterações adicionando pequenos diferenciais no fim do buffer de texto para atestar economia de Token Time
			for i := 1; i <= iteracoes; i++ {
				fmt.Printf("\nTurno Quente (Diferencial) #%d ──\n", i)
				
				// Simulando prompt similar onde uma sub-edição entra
				basePrompt = fmt.Sprintf("%s Edit Delta %v+", basePrompt, i)
				fmt.Printf(" Prompt: %s\n", truncateStr(basePrompt, 50))
				
				outQuente := wrapper.ProcessPrompt(basePrompt, opts)
				printInferenceMetrics(outQuente, float64(outFrio.TimeToken.Milliseconds()))
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&iteracoes, "iteracoes", "i", 2, "Quantidade de repetições aplicando deltas em cima do Contexto em Memória")
	return cmd
}

func printInferenceMetrics(out inference.InferenceOutput, frioBase float64) {
	fmt.Printf(" 📈 Bypassed: %-3d chunks\n", out.Bypassed)
	fmt.Printf(" ⚙️ Computed: %-3d chunks\n", out.Computed)
	fmt.Printf(" ⏱️ TTFT:    %v\n", out.TimeToken)

	if out.Bypassed > 0 && frioBase > 0 {
		economia := 100.0 - (float64(out.TimeToken.Milliseconds()) / frioBase * 100.0)
		fmt.Printf(" 📉 Economia (Latency Drop): %.1f%%\n", economia)
	}
}

func truncateStr(s string, limit int) string {
	if len(s) > limit {
		return s[:limit] + "..."
	}
	return s
}
