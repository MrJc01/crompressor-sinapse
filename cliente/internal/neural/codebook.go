package neural

import (
	"crypto/rand"
	"math/big"
)

// BlockSize define o tamanho de cada fragmento alocado contiguamente do Codebook de Pesos.
// Na matemática do projeto um bloco típico seria 64x64 matriz (4096 valores).
const BlockSize = 4096

// WeightBlock simboliza um bloco de pesos de um Modelo Neural denso em float32.
type WeightBlock struct {
	Data []float32
}

// WeightCodebook é a biblioteca "Base Imutável" construída teoricamente como Mmap File.
// Ele centraliza os pesos gerais de inteligência e NUNCA é re-treinado num processo clássico.
type WeightCodebook struct {
	Blocks []WeightBlock
	Size   int
}

// NewWeightCodebook aloca randomicamente as matrizes de pseudo-memória (Pesos do Pre-Treinamento Prévio) 
func NewWeightCodebook(numBlocks int) *WeightCodebook {
	blocks := make([]WeightBlock, numBlocks)

	for i := 0; i < numBlocks; i++ {
		blocks[i] = WeightBlock{
			Data: generateRandomWeights(BlockSize),
		}
	}

	return &WeightCodebook{
		Blocks: blocks,
		Size:   numBlocks * BlockSize,
	}
}

// TotalSize calcula o foot-print métrico Float32 de todos os blocos na arquitetura estática.
func (c *WeightCodebook) TotalSize() int {
	return c.Size
}

// generateRandomWeights cria pesos densos na inicialização simulando LLM GGUF nativo.
func generateRandomWeights(size int) []float32 {
	w := make([]float32, size)
	for i := 0; i < size; i++ {
		// math/rand para float32 num intervalo realístico via crypto/rand
		val, _ := rand.Int(rand.Reader, big.NewInt(2000))
		w[i] = (float32(val.Int64()) - 1000.0) / 1000.0 // entre -1.0 e 1.0
	}
	return w
}
