package inference

import (
	"crypto/rand"
	"math/big"
	"time"

	"github.com/MrJc01/crompressor-sinapse/internal/cdc"
)

// InjectedLatencyMs define o delay induzido para cada chunk não-cacheado.
const InjectedLatencyMs = 15

// SimulatedWrapper compõe um Mock de Interceptador LLM GGUF.
type SimulatedWrapper struct {
	cache      *ActivationCache
	dim        int // O tamanho final do Array de Float (embedding layer dimension, ex: 768)
}

// NewSimulatedWrapper instacia o mock interceptado que se conecta ao LRU Cache criado.
func NewSimulatedWrapper(cache *ActivationCache, dimension int) *SimulatedWrapper {
	if dimension <= 0 {
		dimension = 768
	}

	return &SimulatedWrapper{
		cache:      cache,
		dim:        dimension,
	}
}

// InferenceOutput reflete a saída final e a métrica de uma geração isolada de Forward Differential Pass.
type InferenceOutput struct {
	Vectors   [][]float32
	Bypassed  int
	Computed  int
	TimeToken time.Duration
}

// ProcessPrompt simula o processo central Sinapse GGUF (Interceptador). 
// 1. O prompt do user vira CDC. 
// 2. Os chunks geram um Hash e checamos o LRU Cache para recuperar a Ativação Neural
// 3. O que der `hit` passamos reto (Bypass O(1)).
// 4. O que der `miss` punimos com Thread.Sleep(InjectedLatencyMs) gerando "vetores estocásticos".
func (s *SimulatedWrapper) ProcessPrompt(prompt string, opts cdc.Options) InferenceOutput {
	start := time.Now()

	// 1. CDC Tokenization Semântica
	chunks := cdc.Tokenize([]byte(prompt), opts)
	
	out := InferenceOutput{
		Vectors: make([][]float32, 0, len(chunks)),
	}

	for _, chunk := range chunks {
		// 2. Consultando o Cache Forward Neural
		if latent, hit := s.cache.Get(chunk.Hash); hit {
			// Bypass
			out.Vectors = append(out.Vectors, latent)
			out.Bypassed++
			continue
		}

		// 3. Fallback ao Cômputo (ForwardPass Tradicional Pesado Simulado por Delay Lock)
		time.Sleep(time.Millisecond * InjectedLatencyMs)
		
		simVector := s.fakeDenseLayerGeneration()
		
		out.Vectors = append(out.Vectors, simVector)
		out.Computed++

		// Cacheamento obrigatório pra futuras inferências semelhantes colherem os frutos do Bypass
		s.cache.Put(chunk.Hash, simVector)
	}

	out.TimeToken = time.Since(start)

	return out
}

// fakeDenseLayerGeneration devolve floats aleatórios representando pesos treinados ou ativações da camada Causal.
func (s *SimulatedWrapper) fakeDenseLayerGeneration() []float32 {
	v := make([]float32, s.dim)
	for i := 0; i < s.dim; i++ {
		// math/rand para float artificial (simulando entropia LLM)
		val, _ := rand.Int(rand.Reader, big.NewInt(100))
		v[i] = float32(val.Int64()) / 100.0
	}
	return v
}
