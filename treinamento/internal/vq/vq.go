package vq

import (
	"fmt"
	"math"
)

// O módulo VQ-NN implementa Vector Quantization Neural Networks puro em O(1) Golang.
// Banindo Tensores densos, nós converteremos ativações de Inteligências Artificiais em "IDs Discretos".

type VQLayer struct {
	Codebook [][]float32 // K Vetores Latentes Universais (Pinos do Dicionário)
	K        int         // Tamanho do vocabulário/dicionário K
	D        int         // Dimensão original geométrica da Inteligência
}

// NewVQLayer instancia a camada esparsa treinável no Codebook O(1).
func NewVQLayer(k int, d int) *VQLayer {
	// Exemplo: 512 IDs de 768 dimensões = Codebook de ~1.5 MB em RAM
	codebook := make([][]float32, k)
	for i := range codebook {
		codebook[i] = make([]float32, d)
		// Pesos são inicializados perto do zero na auscultação de Tensores Reais (Simulado)
		for j := 0; j < d; j++ {
			codebook[i][j] = float32(i+j) * 0.001 
		}
	}
	return &VQLayer{
		Codebook: codebook,
		K:        k,
		D:        d,
	}
}

// Forward varre o input de Ponto Flutuante denso gerado pela "Llama",
// e obriga o dado a colar no Codeword de ID (Int) mais idêntico. Compressão Numérica Total.
func (vq *VQLayer) Forward(z []float32) (quantized []float32, bestIdx int, err error) {
	if len(z) != vq.D {
		return nil, -1, fmt.Errorf("dimensão inválida: experada %d, recebida %d", vq.D, len(z))
	}

	minDist := float32(math.MaxFloat32)
	bestIdx = 0

	// Abordagem Sequencial Linear para fins educacionais puristas do Treino (L2 Distance)
	for i, cw := range vq.Codebook {
		dist := l2Distance(z, cw)
		if dist < minDist {
			minDist = dist
			bestIdx = i
		}
	}

	return vq.Codebook[bestIdx], bestIdx, nil
}

// UpdateCodebook simula matematicamente a backpropagation para atualizar o pino.
// "Puxamos" o Codebook ligeiramente à direção da ativação real usando Exponential Moving Average.
func (vq *VQLayer) UpdateCodebook(idx int, z []float32, learningRate float32) {
	cw := vq.Codebook[idx]
	for i := 0; i < vq.D; i++ {
		// Equação de Delta Simples em CPU
		cw[i] = cw[i] + learningRate*(z[i]-cw[i])
	}
}

// l2Distance calcula o delta euclidiano L^2
func l2Distance(a, b []float32) float32 {
	var sum float32
	for i := 0; i < len(a); i++ {
		diff := a[i] - b[i]
		sum += diff * diff
	}
	return sum
}
