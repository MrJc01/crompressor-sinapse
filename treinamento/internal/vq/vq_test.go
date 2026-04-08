package vq

import (
	"fmt"
	"math"
	"testing"
)

// Teste estrutural das leis O(1) matemáticas
// Simularemos a injeção estocástica de floats de C++ LLMs e observaremos a rede aprender usando apenas index IDs.

func TestVQLayer_Forward_And_Learn(t *testing.T) {
	dimensao := 8
	vocabulario := 16
	learningRate := float32(0.5)

	layer := NewVQLayer(vocabulario, dimensao)

	// Simula a ativação (Batch) flutuante vindo da Camada N-1 
	// (Simulando uma Frase da IA: "[0.9, 0.9, 0.9...]")
	batchZ := []float32{0.9, 0.9, 0.9, 0.9, 0.9, 0.9, 0.9, 0.9}

	// 1. FORWARD INICIAL ANTES DO TREINAMENTO
	vectorQuantizado, initialID, err := layer.Forward(batchZ)
	if err != nil {
		t.Fatalf("Erro no processamento vq forward: %v", err)
	}

	distanciaInicial := l2Distance(batchZ, vectorQuantizado)
	fmt.Printf("➜ Antes do Treino do Codebook:\n")
	fmt.Printf("A Matriz Caótica encontrou o ID %d. \nDistância Numérica: %.4f\n\n", initialID, distanciaInicial)

	// 2. BACKPROPAGATION (Modificação de Pesos)
	// Treinamos forçando a VQ a absorver os vetores
	for epoca := 1; epoca <= 5; epoca++ {
		layer.UpdateCodebook(initialID, batchZ, learningRate)
		
		vetorAtual, _, _ := layer.Forward(batchZ)
		erro := l2Distance(batchZ, vetorAtual)
		fmt.Printf("Época %d - Distância Resídual caindo: %.4f\n", epoca, erro)
	}

	vetorFinal, _, _ := layer.Forward(batchZ)
	distanciaFinal := l2Distance(batchZ, vetorFinal)
	
	fmt.Printf("\n➜ Após 5 Épocas de Deltas Constantes O(1):\n")
	fmt.Printf("Distância Numérica: %.4f\n", distanciaFinal)

	if distanciaFinal > distanciaInicial {
		t.Fatalf("As leis da matemática falharam, perda subiu: %.4f > %.4f", distanciaFinal, distanciaInicial)
	}
	
	if math.Abs(float64(distanciaFinal)) > 0.1 {
		t.Errorf("Codebook não convergiu propriamente.")
	}
	// Sucesso Educativo Comprovado!
}
