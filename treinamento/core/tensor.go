package core

import (
	"errors"
)

// TensorO1 opera puramente em quantização discreta buscando alocação em bloco O(1).
// Abandona os pontos flutuantes do PyTorch, assumindo base vetorizada e esparsa.
type TensorO1 struct {
	Shape       []int
	BaseWeights []uint32 // Modelo-Base estático em memória (Dicionário)
	XORDeltas   []uint32 // Máscaras Esparsas aplicadas sob demanda
}

// NewTensorO1 inicializa o "tensor" Go pré-alocando para garantir zero-alloc nos passes seguintes.
func NewTensorO1(shape []int) *TensorO1 {
	size := 1
	for _, dim := range shape {
		size *= dim
	}
	return &TensorO1{
		Shape:       shape,
		BaseWeights: make([]uint32, size),
		XORDeltas:   make([]uint32, size),
	}
}

// ForwardDiscreteUpdate computa o output em complexidade O(1) para uma matriz esparsa ativada.
// Utiliza um array "out" previamente alocado para atingir zero-allocation no loop nativo.
func (t *TensorO1) ForwardDiscreteUpdate(inputIDs []uint32, out []uint32) error {
	if len(inputIDs) > len(out) {
		return errors.New("tamanho de output alocado é menor do que os input IDs")
	}

	for i := 0; i < len(inputIDs); i++ {
		idx := inputIDs[i]
		if int(idx) < len(t.BaseWeights) {
			// Aplica a mutação matemática (XOR) sem custo de operação Float32
			out[i] = t.BaseWeights[idx] ^ t.XORDeltas[idx]
		}
	}
	return nil
}
