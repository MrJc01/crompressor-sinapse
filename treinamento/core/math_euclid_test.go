package core

import (
	"testing"
)

// TestEuclideanBackprop comprova a capacidade do nosso treinamento nativo O(1) de:
// 1) Converter as diferenças num Delta Vetorial
// 2) Ajustar os tensores de forma exata e contígua em um único passo euclidiano
func TestEuclideanBackprop(t *testing.T) {
	tensor := NewTensorO1([]int{1024})

	// Inputs acionados (Rabin HashMap hits falsos positivos em um caso de simulação)
	inputIDs := []uint32{500, 501, 502} 
	
	// Pesos sintéticos alocados no Hash O(1) do tensor Base
	tensor.BaseWeights[500] = 0xAAAA // 1010...
	tensor.BaseWeights[501] = 0xBBBB // 1011...
	tensor.BaseWeights[502] = 0xCCCC // 1100...

	out := make([]uint32, 3)
	err := tensor.ForwardDiscreteUpdate(inputIDs, out)
	if err != nil {
		t.Fatalf("Critico: Erro mecânico no Forward: %v", err)
	}

	// Alvo desejado pelo LLM para o Treinamento ("Target Label")
	target := []uint32{0xFFFF, 0xFFFF, 0xFFFF}

	// O(1) Backprop Update
	tensor.ApplyBackprop(inputIDs, out, target)

	// Validando o 'Aprendizado' Esparso
	outVerificado := make([]uint32, 3)
	err = tensor.ForwardDiscreteUpdate(inputIDs, outVerificado)
	if err != nil {
		t.Fatalf("Critico: Erro mecânico na validacao forward: %v", err)
	}

	for i := 0; i < len(target); i++ {
		if outVerificado[i] != target[i] {
			t.Errorf("Neural Incoherence! Convergence Failed no index %d: Esperado %X, Obtido O(1) Output de %X", i, target[i], outVerificado[i])
		}
	}
}

// BenchmarkForwardPassZeroAlloc prova a tese da performance estrita para uso em Produção
func BenchmarkForwardPassZeroAlloc(b *testing.B) {
	tensor := NewTensorO1([]int{4096})
	inputIDs := []uint32{10, 20, 100, 4000, 4095}
	out := make([]uint32, 5)

	// Configuração base p/ Benchmarks SRE
	tensor.BaseWeights[10] = 0x01
	tensor.XORDeltas[10] = 0x11
	
	b.ResetTimer()
	b.ReportAllocs() // Ativa flag Zero-Alloc (0 allocs/op esperado)
	
	for i := 0; i < b.N; i++ {
		_ = tensor.ForwardDiscreteUpdate(inputIDs, out)
	}
}
