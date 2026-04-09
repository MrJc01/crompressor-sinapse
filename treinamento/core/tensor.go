package core

import (
	"math"
	"math/rand"
	"sort"
)

// ========= TREINAMENTO (Float32) =========

type CROMExpert struct {
	Weights []float32 // Tamanho: Dim * VocabSize
	Bias    []float32 // Tamanho: VocabSize
}

type CROMModel struct {
	VocabSize int
	Dim       int
	Emb       []float32
	Experts   []CROMExpert
}

func NewCROMModel(vocabSize, dim, numExperts int) *CROMModel {
	m := &CROMModel{
		VocabSize: vocabSize,
		Dim:       dim,
		Emb:       make([]float32, vocabSize*dim),
		Experts:   make([]CROMExpert, numExperts),
	}
	// Init weights (Xavier mini)
	scale := float32(math.Sqrt(2.0 / float64(dim)))
	for i := range m.Emb {
		m.Emb[i] = (float32(i%3) - 1.0) * scale * 0.1 // pseudo random deterministico
	}
	for i := range m.Experts {
		m.Experts[i].Weights = make([]float32, dim*vocabSize)
		m.Experts[i].Bias = make([]float32, vocabSize)
		for j := range m.Experts[i].Weights {
			m.Experts[i].Weights[j] = (float32((i+j)%3) - 1.0) * scale * 0.1
		}
	}
	return m
}

// Forward e Backward Pass p/ Treinamento
func (m *CROMModel) TrainStep(ctxWords []int, targetWord int, expertID int, lr float32) float32 {
	// 1. Embedding Fetch (Averaging / CBOW)
	ctxVec := make([]float32, m.Dim)
	for _, wID := range ctxWords {
		offset := wID * m.Dim
		for d := 0; d < m.Dim; d++ {
			ctxVec[d] += m.Emb[offset+d]
		}
	}
	for d := 0; d < m.Dim; d++ {
		ctxVec[d] /= float32(len(ctxWords))
	}

	// 2. Linear Layer do Expert Roteado
	exp := m.Experts[expertID]
	logits := make([]float32, m.VocabSize)
	for i := 0; i < m.VocabSize; i++ {
		sum := exp.Bias[i]
		wOffset := i * m.Dim
		for d := 0; d < m.Dim; d++ {
			sum += exp.Weights[wOffset+d] * ctxVec[d]
		}
		logits[i] = sum
	}

	// 3. Softmax
	maxL := logits[0]
	for i := 1; i < m.VocabSize; i++ {
		if logits[i] > maxL {
			maxL = logits[i]
		}
	}
	sumExp := float32(0.0)
	probs := make([]float32, m.VocabSize)
	for i := 0; i < m.VocabSize; i++ {
		p := float32(math.Exp(float64(logits[i] - maxL)))
		probs[i] = p
		sumExp += p
	}
	for i := 0; i < m.VocabSize; i++ {
		probs[i] /= sumExp
	}

	loss := -float32(math.Log(float64(probs[targetWord] + 1e-9)))

	// 4. Backpropagation (Gradientes Cross-Entropy)
	err := make([]float32, m.VocabSize)
	for i := 0; i < m.VocabSize; i++ {
		err[i] = probs[i]
	}
	err[targetWord] -= 1.0

	// Gradiente p/ Contexto (Emb)
	gradCtx := make([]float32, m.Dim)
	for i := 0; i < m.VocabSize; i++ {
		e := err[i]
		wOffset := i * m.Dim
		// Atualiza o Expert Weight e Bias
		exp.Bias[i] -= lr * e
		for d := 0; d < m.Dim; d++ {
			gradCtx[d] += e * exp.Weights[wOffset+d]
			exp.Weights[wOffset+d] -= lr * e * ctxVec[d]
		}
	}

	// Gradiente passa pelo Emb
	for _, wID := range ctxWords {
		offset := wID * m.Dim
		for d := 0; d < m.Dim; d++ {
			m.Emb[offset+d] -= lr * (gradCtx[d] / float32(len(ctxWords)))
		}
	}

	return loss
}

// ========= INFERÊNCIA INT8 QUANTIZADA CROM =========

type Int8Expert struct {
	Weights []int8
	Bias    []int32
}

type Int8Layer struct {
	WQ      []int8
	WK      []int8
	WV      []int8
	NormW   []float32
	NormB   []float32
	Experts []Int8Expert
}

type Int8Model struct {
	VocabSize int
	SeqLen    int
	Dim       int
	NumLayers int
	Scale     float32
	Emb       []int8
	Pos       []int8
	Layers    []Int8Layer
	LMHeadW   []int8
	LMHeadB   []int32
}

// QuantizeModel exporta o Float32 treinado para o Formato BitNet/Int8 (8x mais leve)
// QuantizeModel (Legado Go-Native CROM)
func (m *CROMModel) QuantizeModel() *Int8Model {
    // Agora o Treinamento roda em GPU PyTorch (colab_hijacker.py) e exporta o binário.
    // Esta função foi suprimida para dar lugar ao parser deep-stack C.
	return nil
}

// O Próprio Motor de Atenção escrito em Matemática da CPU em Go
// Calcula Foco Sintático sem depender de Arrays Pytorch
func (q *Int8Model) ForwardInt8(ctxWords []int, expertID int, penalty map[int]bool) int {
	seqLen := len(ctxWords)
	if seqLen == 0 { return 0 }

	// Array Inicial (Estado Zero)
	X := make([][]int32, seqLen)
	for i, wID := range ctxWords {
		X[i] = make([]int32, q.Dim)
		offset := wID * q.Dim
		posOffset := i * q.Dim
		for d := 0; d < q.Dim; d++ {
			X[i][d] = int32(q.Emb[offset+d]) + int32(q.Pos[posOffset+d])
		}
	}

	// Sincronia Matemática Pytorch: Acumulador constante para não estourar
	scaleToPyTorch := float32(math.Pow(float64(q.Scale), 4)) / float32(math.Sqrt(float64(q.Dim)))

	// Loop Vertical PROFUNDO (DeepStacking)
	for lIdx := 0; lIdx < q.NumLayers; lIdx++ {
		layer := q.Layers[lIdx]
		
		Q := make([]int32, q.Dim)
		for out := 0; out < q.Dim; out++ {
			var sum int32
			wOffset := out * q.Dim
			for in := 0; in < q.Dim; in++ {
				sum += X[seqLen-1][in] * int32(layer.WQ[wOffset+in])
			}
			Q[out] = sum
		}

		K := make([][]int32, seqLen)
		V := make([][]int32, seqLen)

		for i := 0; i < seqLen; i++ {
			K[i] = make([]int32, q.Dim)
			V[i] = make([]int32, q.Dim)
			for out := 0; out < q.Dim; out++ {
				var sumK, sumV int32
				wOffset := out * q.Dim
				for in := 0; in < q.Dim; in++ {
					e_val := X[i][in]
					sumK += e_val * int32(layer.WK[wOffset+in])
					sumV += e_val * int32(layer.WV[wOffset+in])
				}
				K[i][out] = sumK
				V[i][out] = sumV
			}
		}

		scores := make([]float32, seqLen)
		var maxScore float32 = -9999999.0
		for i := 0; i < seqLen; i++ {
			var dot int64
			for d := 0; d < q.Dim; d++ {
				dot += int64(Q[d]) * int64(K[i][d])
			}
			sc := float32(dot) * scaleToPyTorch
			scores[i] = sc
			if sc > maxScore { maxScore = sc }
		}

		var sumExp float32
		for i := 0; i < seqLen; i++ {
			p := float32(math.Exp(float64(scores[i] - maxScore)))
			scores[i] = p
			sumExp += p
		}
		
		ctxVecFloat := make([]float32, q.Dim)
		for i := 0; i < seqLen; i++ {
			prob := scores[i] / sumExp
			for d := 0; d < q.Dim; d++ {
				ctxVecFloat[d] += float32(V[i][d]) * prob
			}
		}

		ctxVec := make([]int32, q.Dim)
		for d := 0; d < q.Dim; d++ {
			ctxVec[d] = int32(ctxVecFloat[d] * q.Scale)
		}

		exp := layer.Experts[expertID]
		// Passa pelo Expert Intermediário (DIM -> DIM)
		expertOut := make([]int32, q.Dim)
		for out := 0; out < q.Dim; out++ {
			sum := exp.Bias[out]
			wOffset := out * q.Dim
			for in := 0; in < q.Dim; in++ {
				sum += int32(exp.Weights[wOffset+in]) * ctxVec[in]
			}
			expertOut[out] = sum
		}

		// Fase 29 SRE FIX: Layer Normalization O(1)
		// Restaura as matrizes para o domínio Causal calculando a Variância real e aplicando Gamma/Beta.
		var sumMean float32
		fltX := make([]float32, q.Dim)
		for d := 0; d < q.Dim; d++ {
			// X no Go está escalado (int ints), então o expertOut converte para a mesma base.
			val := float32(X[seqLen-1][d]) + (float32(expertOut[d]) * q.Scale)
			fltX[d] = val
			sumMean += val
		}
		mean := sumMean / float32(q.Dim)

		var sumVar float32
		for d := 0; d < q.Dim; d++ {
			diff := fltX[d] - mean
			sumVar += diff * diff
		}
		variance := sumVar / float32(q.Dim)
		stdDev := float32(math.Sqrt(float64(variance + 1e-5)))

		// Aplicar Peso (Gamma), Viés (Beta) da LayerNorm originais do Pytorch, 
		// e em seguida **Requantizar** para a escala Inteira (dividir por Scale) para o próximo loop!
		invScale := 1.0 / q.Scale
		for d := 0; d < q.Dim; d++ {
			ptNorm := ((fltX[d] - mean) / stdDev) * layer.NormW[d] + layer.NormB[d]
			X[seqLen-1][d] = int32(ptNorm * invScale)
		}
	}

	// Final = LM Head (DIM -> VocabSize)
	finalCtx := X[seqLen-1]
	
	type LogitPair struct {
		Word  int
		Score float32
	}
	var pairs []LogitPair
		
	for i := 1; i < q.VocabSize; i++ {
		sum := q.LMHeadB[i]
		wOffset := i * q.Dim
		for d := 0; d < q.Dim; d++ {
			sum += int32(q.LMHeadW[wOffset+d]) * finalCtx[d]
		}
		
		if penalty[i] {
			continue // Ignora e impede o looping
		}

		// Reverter o Score Logit Int32 para Float32 Pytorch Absoluto = sum * Scale^2
		fScore := float32(sum) * q.Scale * q.Scale
		pairs = append(pairs, LogitPair{Word: i, Score: fScore})
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Score > pairs[j].Score // Descending order
	})

	// TOP-K Sampling Lógico: Foca nas N respostas mais próximas da intenção
	topK := 8
	if topK > len(pairs) { topK = len(pairs) }
	if topK == 0 { return 0 }
	
	temperature := float32(0.35) // Resgatando de 0.01 para permitir a ramificação do Contexto Injetado
	var sumProbs float32
	probs := make([]float32, topK)
	
	for i := 0; i < topK; i++ {
		probs[i] = float32(math.Exp(float64(pairs[i].Score / temperature)))
		sumProbs += probs[i]
	}

	r := rand.Float32() * sumProbs
	var accum float32
	for i := 0; i < topK; i++ {
		accum += probs[i]
		if r <= accum {
			return pairs[i].Word
		}
	}

	return pairs[0].Word
}
