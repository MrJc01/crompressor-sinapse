// Package codebook implementa um dicionário dinâmico de chunks CDC.
// Registra chunks inéditos e rastreia frequência de acesso para métricas de deduplicação.
package codebook

import (
	"sync"

	"github.com/MrJc01/crompressor-sinapse/internal/cdc"
)

// Entry representa uma entrada no Codebook.
type Entry struct {
	Hash     uint64
	Size     uint32
	HitCount uint64
}

// Stats contém estatísticas do Codebook.
type Stats struct {
	TotalEntries uint64
	TotalHits    uint64
	TotalMisses  uint64
	HitRate      float64
}

// Codebook é um dicionário dinâmico thread-safe de chunks CDC.
type Codebook struct {
	mu      sync.RWMutex
	entries map[uint64]*Entry
	hits    uint64
	misses  uint64
}

// New cria um Codebook vazio.
func New() *Codebook {
	return &Codebook{
		entries: make(map[uint64]*Entry),
	}
}

// Insert registra um chunk no Codebook. Retorna true se é novo (inédito).
func (cb *Codebook) Insert(chunk cdc.Chunk) bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if entry, exists := cb.entries[chunk.Hash]; exists {
		entry.HitCount++
		cb.hits++
		return false
	}

	cb.entries[chunk.Hash] = &Entry{
		Hash:     chunk.Hash,
		Size:     chunk.Size,
		HitCount: 1,
	}
	cb.misses++
	return true
}

// Lookup verifica se um hash existe no Codebook.
func (cb *Codebook) Lookup(hash uint64) (*Entry, bool) {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	entry, exists := cb.entries[hash]
	if exists {
		cb.mu.RUnlock()
		cb.mu.Lock()
		entry.HitCount++
		cb.hits++
		cb.mu.Unlock()
		cb.mu.RLock()
		return entry, true
	}
	return nil, false
}

// InsertAll registra todos os chunks de uma vez. Retorna quantos são novos.
func (cb *Codebook) InsertAll(chunks []cdc.Chunk) int {
	newCount := 0
	for _, c := range chunks {
		if cb.Insert(c) {
			newCount++
		}
	}
	return newCount
}

// Stats retorna estatísticas de uso do Codebook.
func (cb *Codebook) Stats() Stats {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	total := cb.hits + cb.misses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(cb.hits) / float64(total)
	}

	return Stats{
		TotalEntries: uint64(len(cb.entries)),
		TotalHits:    cb.hits,
		TotalMisses:  cb.misses,
		HitRate:      hitRate,
	}
}

// Len retorna o número de entradas únicas.
func (cb *Codebook) Len() int {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return len(cb.entries)
}

// Reset limpa o Codebook.
func (cb *Codebook) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.entries = make(map[uint64]*Entry)
	cb.hits = 0
	cb.misses = 0
}
