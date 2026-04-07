package neural

// ScaleFactor simula o desvio paramétrico usado em Quantização 4-Bit e Int8 Delta.
// Nosso mock aqui aplica valores com isso pra simplificar as flutuações nativamente em Go.
const ScaleFactor float32 = 0.05

// SparseDelta guarda as unhas diferenciais exclusivas de uma Skil neural aprendida.
type SparseDelta struct {
	// Posições com base no Bloco de tamanho 4096 (O(1) look-ups)
	Positions []int
	// Os Deltas isolados já submetidos a simulação de quantização em Int8 pseudo-esparsa
	Values []float32 
}

// NonZeroEntries retorna um interador virtual por conveniência associando index a magnitude de delta.
func (sd *SparseDelta) NonZeroEntries() map[int]float32 {
	mapping := make(map[int]float32, len(sd.Positions))
	for i, pos := range sd.Positions {
		mapping[pos] = sd.Values[i]
	}
	return mapping
}

// DeltaMap concentra exclusivamente a "Especialização" (ex: LoRA Dinâmico Inteligente).
type DeltaMap struct {
	// key = BlockIndex (relativo ao matriz base principal de Mmap)
	Entries map[int]SparseDelta
}

// NewDeltaMap inicializa um mapa vazio preparatório.
func NewDeltaMap() *DeltaMap {
	return &DeltaMap{
		Entries: make(map[int]SparseDelta),
	}
}

// Reconstruct computa o Forward Efetivo mesclando os Pesos Básicos "Frios" e as pequenas amarras de XOR/Delta Ativas.
// Matematicamente: W_efetivo = W_base ⊕ Δ
func Reconstruct(codebook *WeightCodebook, delta *DeltaMap) []float32 {
	// A RAM massiva para suportar o estado efetivado da Inferência neural contígua.
	effWeights := make([]float32, codebook.TotalSize())

	// Fusão Linear - (Codebook Imutável vs Delta)
	for blockIdx, block := range codebook.Blocks {
		start := blockIdx * BlockSize
		end := start + BlockSize
		
		// 1. Array Copy Cru (W_base)
		copy(effWeights[start:end], block.Data)

		// 2. Mesclagem em O(1) de Delta, apenas se houver Skil Delta em cima da área!
		if d, ok := delta.Entries[blockIdx]; ok {
			for pos, val := range d.NonZeroEntries() {
				// Simulação de Descompressão de Quantização: 
				// (O modelo matemático puro do Sinapse aplicaria XOR bit-a-bit aqui, 
				// nós adicionaremos o valor delta flutuando pelo Fator Escalar).
				effWeights[start+pos] += val * ScaleFactor
			}
		}
	}

	return effWeights
}

// MemoryAnalysis mede os dados armazenados nas duas arquiteturas base para relatório O(1) de ganho.
func (dm *DeltaMap) MemoryAnalysis() (activeEntries, sparsityPercentage float64) {
	totalActive := 0
	for _, sd := range dm.Entries {
		totalActive += len(sd.Positions)
	}

	return float64(totalActive), float64(totalActive)
}
