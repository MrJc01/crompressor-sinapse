package inference

import (
	"reflect"
	"testing"
)

func TestActivationCache_PutGet(t *testing.T) {
	cache := NewActivationCache(5)
	
	vector1 := []float32{0.45, -0.12, 0.99}
	cache.Put(101, vector1)

	got, found := cache.Get(101)
	if !found {
		t.Fatalf("Esperava encontrar a chave 101 inserida no cache LRU")
	}

	if !reflect.DeepEqual(got, vector1) {
		t.Errorf("Vetor recuperado é inconsistente: Obtido %v Esperado %v", got, vector1)
	}

	// Miss Condition
	_, foundMiss := cache.Get(404)
	if foundMiss {
		t.Error("Não esperava recuperar chave inexistente")
	}
}

func TestActivationCache_LRUEviction(t *testing.T) {
	// Cache suporta apenas 3 itens
	cache := NewActivationCache(3)
	
	cache.Put(1, []float32{1.0})
	cache.Put(2, []float32{2.0})
	cache.Put(3, []float32{3.0}) // Layout Interno: [3, 2, 1]
	
	// Um acesso promove ao topo -> Layout Interno [1, 3, 2]
	cache.Get(1)
	
	// Adiciona item adicional causando eviction ao "mais velho dos novos" (exclui o 2)
	cache.Put(4, []float32{4.0}) // Layout Interno: [4, 1, 3]

	if _, found := cache.Get(4); !found {
		t.Errorf("Chave mais nova 4 foi rejeitada de inserção")
	}

	if _, found := cache.Get(1); !found {
		t.Errorf("Chave 1 que foi promovida falsamente evictada")
	}

	if _, found := cache.Get(2); found {
		t.Errorf("Chave 2 não foi evictada como prescreve a política LRU")
	}

	hits, misses, length := cache.Stats()
	if length != 3 {
		t.Errorf("Length inconsistente: %d (capacidade 3)", length)
	}
	t.Logf("LRU Stats - Hits: %d, Misses: %d, Len: %d", hits, misses, length)
}
