package neural

import (
	"crypto/rand"
	"math/big"
	"time"
)

// TrainMetrics reporta estatísticas comparativas entre pesos clássicos flutuados e a revolução Compressão O(1).
type TrainMetrics struct {
	Duration         time.Duration
	BaseCodebookSize int     // Bytes (float32 == 4 bytes/valor)
	ComputedDeltaSize int    // Bytes (Posição em 'int' e float delta)
	SparsityRate     float64 // 0 a 100% (Grau de compactação)
}

// Trainer gerencia um preudo loop Backward pass aplicando esparsidade induzida do projeto crompressor na gravação da Matrix.
type Trainer struct {
	codebook     *WeightCodebook
	diffRatio    float64 // Limiar que filtra atualizações pequenas (simulando Dropout Mask + Delta).
}

// NewTrainer gera um Loop Trainer Mock de Gradientes Neurais para provas físicas.
func NewTrainer(base *WeightCodebook, sparsityRatio float64) *Trainer {
	return &Trainer{
		codebook:  base,
		diffRatio: sparsityRatio,
	}
}

// SimulateEpoch "treina" injetando aleatoriedades direcionais como um Gradiente massivo mas guarda exclusivamente 1 Delta Map.
// Gradientes abaixo do DiffRatio não reescrevem sobre a base, consolidando "Sparsity".
func (t *Trainer) SimulateEpoch() (*DeltaMap, TrainMetrics) {
	start := time.Now()
	delta := NewDeltaMap()

	totalBaseElements := t.codebook.Size
	totalAllocatedDeltas := 0

	for i := 0; i < len(t.codebook.Blocks); i++ {
		var changedPositions []int
		var changedValues []float32
		
		seen := make(map[int]bool)

		for pos := 0; pos < BlockSize; pos++ {
			// Sorteamos a chance deste pseudo-peso sofrer gradiente baseando-se no diffRatio
			val, _ := rand.Int(rand.Reader, big.NewInt(1000))
			prob := float64(val.Int64()) / 1000.0 // Maior resolução

			if prob < t.diffRatio && !seen[pos] {
				seen[pos] = true
				changedPositions = append(changedPositions, pos)
				
				// O "Caminho do aprendizado" no XOR
				deltaIntensity, _ := rand.Int(rand.Reader, big.NewInt(55))
				learningDelta := float32(deltaIntensity.Int64()-25) / 10.0 // -2.5 a 2.5
				if learningDelta == 0 {
					learningDelta = 0.5
				}
				changedValues = append(changedValues, learningDelta)

				totalAllocatedDeltas++
			}
		}

		if len(changedPositions) > 0 {
			delta.Entries[i] = SparseDelta{
				Positions: changedPositions,
				Values:    changedValues,
			}
		}
	}

	metrics := TrainMetrics{
		Duration:         time.Since(start),
		BaseCodebookSize: totalBaseElements * 4, // 4 bytes pr Float32
		ComputedDeltaSize: totalAllocatedDeltas * 8, // Int(4b) + Float32(4b) = 8 Bytes salvos isoladamente.
	}

	metrics.SparsityRate = 100.0 - (float64(totalAllocatedDeltas) / float64(totalBaseElements) * 100.0)

	return delta, metrics
}
