package codebook

import (
	"testing"

	"github.com/MrJc01/crompressor-sinapse/internal/cdc"
)

func makeChunk(data string) cdc.Chunk {
	b := []byte(data)
	return cdc.Chunk{
		Data: b,
		Size: uint32(len(b)),
		Hash: cdc.HashBytes(b),
	}
}

func TestCodebook_InsertNew(t *testing.T) {
	cb := New()

	c := makeChunk("func main() {}")
	isNew := cb.Insert(c)

	if !isNew {
		t.Error("Primeira inserção deveria retornar true (novo)")
	}
	if cb.Len() != 1 {
		t.Errorf("Codebook deveria ter 1 entrada, tem %d", cb.Len())
	}
}

func TestCodebook_InsertDuplicate(t *testing.T) {
	cb := New()

	c := makeChunk("if err != nil {")
	cb.Insert(c)
	isNew := cb.Insert(c)

	if isNew {
		t.Error("Segunda inserção do mesmo chunk deveria retornar false")
	}
	if cb.Len() != 1 {
		t.Errorf("Codebook deveria ter 1 entrada (dedup), tem %d", cb.Len())
	}
}

func TestCodebook_InsertAll(t *testing.T) {
	cb := New()

	chunks := []cdc.Chunk{
		makeChunk("func main() {}"),
		makeChunk("if err != nil {"),
		makeChunk("func main() {}"), // Duplicado
		makeChunk("return nil"),
	}

	newCount := cb.InsertAll(chunks)

	if newCount != 3 {
		t.Errorf("Esperado 3 novos, obtido %d", newCount)
	}
	if cb.Len() != 3 {
		t.Errorf("Codebook deveria ter 3 entradas, tem %d", cb.Len())
	}
}

func TestCodebook_Stats(t *testing.T) {
	cb := New()

	chunks := []cdc.Chunk{
		makeChunk("padrão A"),
		makeChunk("padrão B"),
		makeChunk("padrão A"), // Hit
		makeChunk("padrão A"), // Hit
		makeChunk("padrão C"),
	}

	cb.InsertAll(chunks)
	stats := cb.Stats()

	t.Logf("Stats: entries=%d, hits=%d, misses=%d, hitRate=%.1f%%",
		stats.TotalEntries, stats.TotalHits, stats.TotalMisses, stats.HitRate*100)

	if stats.TotalEntries != 3 {
		t.Errorf("Entries: esperado 3, obtido %d", stats.TotalEntries)
	}
	if stats.TotalHits != 2 {
		t.Errorf("Hits: esperado 2, obtido %d", stats.TotalHits)
	}
}

func TestCodebook_Reset(t *testing.T) {
	cb := New()
	cb.Insert(makeChunk("dados"))
	cb.Reset()

	if cb.Len() != 0 {
		t.Error("Reset não limpou o Codebook")
	}

	stats := cb.Stats()
	if stats.TotalHits != 0 || stats.TotalMisses != 0 {
		t.Error("Reset não zerou as métricas")
	}
}

func BenchmarkCodebook_Insert(b *testing.B) {
	cb := New()
	chunk := makeChunk("benchmark data for codebook insertion")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb.Insert(chunk)
	}
}
