package neural

import (
	"testing"
)

func TestDeltaSparsityAndReconstruction(t *testing.T) {
	// Apenas 1 Bloco (4096 valores)
	codebook := NewWeightCodebook(1)
	if codebook.TotalSize() != BlockSize {
		t.Fatalf("Codebook size errático: %d vs esperado %d", codebook.TotalSize(), BlockSize)
	}

	// Treinador agressivo que altera em torno de 15% apenas.
	trainer := NewTrainer(codebook, 0.15)
	
	delta, metrics := trainer.SimulateEpoch()
	
	if metrics.SparsityRate < 70 {
		t.Errorf("Compressão de matriz ineficiente: %.1f%% de Sparsity", metrics.SparsityRate)
	}

	t.Logf("Base Train footprint: %d Bytes", metrics.BaseCodebookSize)
	t.Logf("Delta footprint     : %d Bytes", metrics.ComputedDeltaSize)
	t.Logf("Compressão (Sparsity): %.2f%% menos uso de RAM/Disk para a 'Skill'", metrics.SparsityRate)

	// Validação Lógica Reconstruct
	// Sem delta acoplado (delta puramente vazio), o reconstructed deve ser 100% == ao Pretrained
	emptyDelta := NewDeltaMap()
	reconstEmpty := Reconstruct(codebook, emptyDelta)
	
	for i := 0; i < BlockSize; i++ {
		if reconstEmpty[i] != codebook.Blocks[0].Data[i] {
			t.Fatalf("Reconstrutor de Base Falsa. As matrizes cruas sem Delta divergiram.")
		}
	}

	// Com delta
	reconstTrained := Reconstruct(codebook, delta)
	
	diffCount := 0
	for i := 0; i < BlockSize; i++ {
		if reconstTrained[i] != codebook.Blocks[0].Data[i] {
			diffCount++
		}
	}

	// A contagem de diffs deve convergir próximo ao numero de Posições no mapa (apenas 1 entry[0] pois é um bloco)
	if _, ok := delta.Entries[0]; ok {
		posicoesReais := len(delta.Entries[0].Positions)
		if diffCount != posicoesReais {
			t.Errorf("DiffCount Reconstruction (%d) divergente das Posições Alvos O(1) do DeltaMap (%d)", diffCount, posicoesReais)
		}
	}
}
