package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/MrJc01/crompressor-sinapse/internal/neural"
)

// trainCmd roda o simulador de aprendizado Neural XOR Diferencial provando matematicamente o uso estrito e mínimo de VRAM.
func trainCmd() *cobra.Command {
	var sparsity float64
	var blocks int

	cmd := &cobra.Command{
		Use:   "train",
		Short: "Simula o Forward Neural gravando apenas o treinamento esparso Diferencial XOR (Treinamento O(1))",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("🧠 Crompressor-Sinapse — Treinador XOR Diferencial O(1)")
			
			// Aproximadamente simulando um bloco nativo de um Pseudo Modelo LLM base Llama em Go (Ex: 1024 blocos).
			fmt.Printf("\n ⏳ Alocando Codebook Imutável Pre-Trained (Blocks: %d)...\n", blocks)
			codebook := neural.NewWeightCodebook(blocks)
			
			trainer := neural.NewTrainer(codebook, sparsity)

			fmt.Println("\n 🔄 Rodando o Backpropagation Epoch e quantizando a Skill Diferencial...")
			deltaMask, metrics := trainer.SimulateEpoch()

			// Reporting Terminal Output
			fmt.Printf("\n ✅ Treino Finalizado em %v\n", metrics.Duration)
			fmt.Println(" ══════════════════════════════════════════════════════════════")
			fmt.Printf(" 📡 Treinamento Tradicional (Backprop Normal)\n")
			fmt.Printf("    Tamanho Físico Gerado (RAM/Disk): %s\n", formatSize(int64(metrics.BaseCodebookSize)))
			fmt.Println(" ══════════════════════════════════════════════════════════════")
			fmt.Printf(" 🌌 Crompressor XOR Delta (Máscara + Int8 Quant)\n")
			fmt.Printf("    Tamanho Físico Gerado (RAM/Disk): %s\n", formatSize(int64(metrics.ComputedDeltaSize)))
			fmt.Printf("    Taxa de Sparsity (VRAM Save):     %.2f%%\n", metrics.SparsityRate)
			
			// Executar O(1) Fake Load Time
			fmt.Println("\n ⚡ Simulando Carregamento da nova Skill para a Interface GGUF (Forward Efetivo)")
			effWeights := neural.Reconstruct(codebook, deltaMask)
			fmt.Printf("    Mesclados O(1): %d elementos neurais instanciados com sucesso.\n\n", len(effWeights))
			
			return nil
		},
	}

	cmd.Flags().Float64VarP(&sparsity, "sparsity", "s", 0.05, "Limiar de aprendizado/Threshold. % da área da matriz que pode ser modificada por turno (Default 5%)")
	cmd.Flags().IntVarP(&blocks, "blocks", "b", 256, "Tamanho multiplicador da rede imutável")

	return cmd
}
