package tokenizador

// TokenizeToIDs varre a string em blocos, agindo de fato como um CDC adaptado a texto.
// Em O(N) de tempo linear do tamanho do texto, produz instantaneamente tokens de dicionário O(1).
// Abandona os pesos BPE e inferência de Vocabulário da arquitetura Llama/Qwen.
func TokenizeToIDs(text string, dicionarioSize uint32) []uint32 {
	if len(text) == 0 {
		return nil
	}
	
	// Pré-alocando tamanho conservador na Heap para amassar o tempo de GC Time
	output := make([]uint32, 0, len(text)/4+1)
	
	start := 0
	for i := 0; i <= len(text); i++ {
		// Heurística tática: CDC lexical no formato FNV-1a.
		// Em bytes criptografados usaríamos: (h & targetMask == 0) como limite de Janela.
		if i == len(text) || text[i] == ' ' || text[i] == '\n' || text[i] == '.' || text[i] == ',' {
			if i > start {
				// FNV-1a inline desestruturado (Zero-Overhead Call)
				h := uint32(2166136261)
				for j := start; j < i; j++ {
					h ^= uint32(text[j])
					h *= 16777619
				}
				
				// Assentamento do contexto no módulo base TensorO1
				tokenID := h % dicionarioSize
				output = append(output, tokenID)
			}
			start = i + 1
		}
	}
	return output
}
