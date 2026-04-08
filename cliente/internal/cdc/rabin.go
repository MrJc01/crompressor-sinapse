// Package cdc implementa Content-Defined Chunking usando Rabin fingerprint.
// É o coração do Crompressor-Sinapse: detecta fronteiras naturais em texto e código.
package cdc

import (
	"hash/fnv"
	"math/bits"
)

// ── Constantes do Rabin Rolling Hash ──

const (
	// rabinPoly é o polinômio irredutível usado pelo Rabin fingerprint.
	// Escolhido para boa distribuição em texto natural e código-fonte.
	rabinPoly uint64 = 0x3DA3358B4DC173

	// windowSize é o tamanho da janela deslizante (bytes).
	windowSize = 48
)

// RabinHasher implementa um rolling hash baseado em Rabin fingerprint.
type RabinHasher struct {
	hash   uint64
	window [windowSize]byte
	wpos   int
	count  int

	// Tabelas pré-computadas para remoção eficiente do byte mais antigo.
	popTable [256]uint64
}

// NewRabinHasher cria um novo hasher Rabin.
func NewRabinHasher() *RabinHasher {
	r := &RabinHasher{}
	r.precompute()
	return r
}

// precompute gera a tabela de remoção para o polinômio.
func (r *RabinHasher) precompute() {
	for i := 0; i < 256; i++ {
		h := uint64(i)
		for j := 0; j < windowSize; j++ {
			h = (h << 1) ^ (rabinPoly & (0 - (h >> 63)))
		}
		r.popTable[i] = h
	}
}

// Roll adiciona um byte ao hash e remove o byte mais antigo da janela.
// Retorna o hash atualizado.
func (r *RabinHasher) Roll(b byte) uint64 {
	// Remover byte antigo da janela
	if r.count >= windowSize {
		old := r.window[r.wpos]
		r.hash ^= r.popTable[old]
	}

	// Adicionar novo byte
	r.hash = (r.hash << 1) ^ (rabinPoly & (0 - (r.hash >> 63)))
	r.hash ^= uint64(b)

	// Atualizar janela circular
	r.window[r.wpos] = b
	r.wpos = (r.wpos + 1) % windowSize
	r.count++

	return r.hash
}

// Reset limpa o estado do hasher para reutilização.
func (r *RabinHasher) Reset() {
	r.hash = 0
	r.wpos = 0
	r.count = 0
	for i := range r.window {
		r.window[i] = 0
	}
}

// Hash retorna o hash atual.
func (r *RabinHasher) Hash() uint64 {
	return r.hash
}

// ── Hash rápido para chunks (não-rolling) ──

// HashBytes calcula um hash rápido de um bloco de bytes (FNV-1a).
// Usado para indexar chunks no Codebook.
func HashBytes(data []byte) uint64 {
	h := fnv.New64a()
	h.Write(data)
	return h.Sum64()
}

// ── Utilitários ──

// IsBoundary verifica se o hash indica uma fronteira de chunk.
// Usa a máscara para controlar o tamanho médio dos chunks:
//   - mask = (1 << 10) - 1 → avg ~1024 bytes
//   - mask = (1 << 7) - 1  → avg ~128 bytes
func IsBoundary(hash uint64, mask uint64) bool {
	return hash&mask == 0
}

// MaskForAvgSize calcula a máscara de bits para um tamanho médio desejado.
// avgSize deve ser potência de 2 para resultados exatos.
func MaskForAvgSize(avgSize int) uint64 {
	if avgSize <= 0 {
		avgSize = 128
	}
	// Encontrar o bit mais significativo
	b := bits.Len(uint(avgSize)) - 1
	if b < 1 {
		b = 1
	}
	return (1 << b) - 1
}
