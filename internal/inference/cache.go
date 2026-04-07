package inference

import (
	"container/list"
	"sync"
)

// CacheEntry define o registro mapeado de uma ativação de rede neural no Cache LRU.
type CacheEntry struct {
	Hash      uint64      // xxhash64 CDC Chunk do prompt fatiado.
	Latent    []float32   // O vetor de ativação serializado.
	Timestamp int64       // Marca de tempo ou ordem.
}

// ActivationCache é a implementação de um buffer thread-safe com limite de despejo (LRU policy).
type ActivationCache struct {
	mu       sync.RWMutex
	capacity int
	
	// Map para acesso restrito O(1)
	items map[uint64]*list.Element
	
	// Lista duplamente encadeada para orientar a remoção de itens Antigos
	order *list.List
	
	hits   int
	misses int
}

// NewActivationCache inicializa o buffer neural diferencial com a capacidade em quantidade de vetores (não megabytes).
func NewActivationCache(capacity int) *ActivationCache {
	if capacity <= 0 {
		capacity = 1000 // default capacity
	}
	return &ActivationCache{
		capacity: capacity,
		items:    make(map[uint64]*list.Element),
		order:    list.New(),
	}
}

// Put armazena ou insere o Vetor da ativação simulada por sua Hash originária CDC.
func (c *ActivationCache) Put(hash uint64, latent []float32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Se o Hash já estiver cacheados, nós somente o atualizamos e o marcamos como mais recente.
	if elem, ok := c.items[hash]; ok {
		c.order.MoveToFront(elem)
		elem.Value.(*CacheEntry).Latent = latent
		return
	}

	// Caso Cache Limite, retira o mais velho da fila.
	if c.order.Len() >= c.capacity {
		c.evictOldest()
	}

	entry := &CacheEntry{
		Hash:   hash,
		Latent: latent,
	}
	
	elem := c.order.PushFront(entry)
	c.items[hash] = elem
}

// Get recupera o Array Float (vetor neural) partindo do Hash, para burlar Forwarding Pases em LLM.
func (c *ActivationCache) Get(hash uint64) ([]float32, bool) {
	c.mu.RLock()
	elem, ok := c.items[hash]
	if !ok {
		c.misses++
		c.mu.RUnlock()
		return nil, false
	}
	c.mu.RUnlock()

	// Operação de Write - Precisamos marcar Hit e realocá-lo ao topo como LRU manda.
	c.mu.Lock()
	c.hits++
	c.order.MoveToFront(elem)
	c.mu.Unlock()

	return elem.Value.(*CacheEntry).Latent, true
}

func (c *ActivationCache) evictOldest() {
	back := c.order.Back()
	if back != nil {
		c.order.Remove(back)
		delete(c.items, back.Value.(*CacheEntry).Hash)
	}
}

// Stats retorna os hits e misses cumulativos para monitorar a eficácia do Forwarding Diferencial (Bypass).
func (c *ActivationCache) Stats() (hits, misses, len int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.hits, c.misses, c.order.Len()
}

// Clear reseta hard limit
func (c *ActivationCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[uint64]*list.Element)
	c.order.Init()
	c.hits = 0
	c.misses = 0
}
