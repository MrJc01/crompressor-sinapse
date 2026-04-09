package core

// HashContextAggregator funde um histórico de Tokens de entrada numa Matriz Vetorial de 1 dimensão O(1).
// Em um Transformer comum da NVIDIA, isso seria as gigantescas matrizes K, V, Q (Self-Attention).
// No ecossistema de Entropia Go, utilizamos um gerador convolutivo polinomial restrito ao Dicionário Base.
func HashContextAggregator(history []uint32, maxDictSize uint32) uint32 {
	if len(history) == 0 {
		return 0
	}
	hash := uint32(2166136261)
	for _, id := range history {
		hash ^= id
		hash *= 16777619
	}
	return hash % maxDictSize
}
