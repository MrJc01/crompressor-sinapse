package core

import (
	"math/bits"
)

// CalculateEuclideanDelta quantifica o abismo vetorial local. Em ambientes discretos, 
// a distância pode ser representada pelo delta estrutural de bits (XOR).
func CalculateEuclideanDelta(predicted, target uint32) uint32 {
	return predicted ^ target
}

// ApplyBackprop aplica as correções O(1).
// Recebendo os IDs da janela CDC, a matriz ativada gerada e o alvo esperado (target),
// nós extraímos a diferença Delta Euclidiana e injetamos como máscara local.
func (t *TensorO1) ApplyBackprop(inputIDs []uint32, predicted, target []uint32) {
	LogForensic("Camada Aplicação (Backprop)", "Atualizando XORDeltas Esparsas", "activations", len(inputIDs))
	
	for i := 0; i < len(inputIDs); i++ {
		idx := inputIDs[i]
		if int(idx) < len(t.XORDeltas) {
			// Achar a máscara diferencial exata
			mask := CalculateEuclideanDelta(predicted[i], target[i])
			
			// Como o objetivo é sobrescrever via máscara (P2P Mesh Gossip update style), 
			// injetamos a variação sobre a diferença corrente.
			t.XORDeltas[idx] ^= mask
		}
	}
}

// EuclideanDistanceAnalytical auxilia P&D na checagem e assertividade do roteamento
// de entropia usando Contagem de Bits ativada (Hamming distance sobre modelo Euclidiano discreto).
func EuclideanDistanceAnalytical(x, y uint32) int {
	diff := x ^ y
	return bits.OnesCount32(diff)
}
