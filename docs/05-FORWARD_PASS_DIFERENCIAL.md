# 05 — Forward Pass Diferencial

> *"Se o Crompressor não reescreve bytes que já conhece, por que a rede neural recomputa ativações que já calculou?"*

---

## O Desperdício: Recomputação Integral

No pipeline de inferência padrão, quando o modelo recebe um prompt, ele executa o **forward pass completo** — toda a multiplicação de matrizes, para todos os tokens, do zero.

```
Prompt A: "Explique a arquitetura do Crompressor em Go"
  → Forward pass: 100% de computação → Resultado A

Prompt B: "Explique a arquitetura do Crompressor em Rust"  
  → Forward pass: 100% de computação → Resultado B
  
  → 80% do contexto era IDÊNTICO!
  → 80% da computação foi DESPERDIÇADA!
```

---

## A Proposta: Cache de Ativações Indexado por Hash CDC

O estado de ativação da rede pode possuir um **cache indexado pelos hashes dos chunks CDC** da entrada. Se um chunk já passou pela rede e está no Codebook de ativações, a inferência **pula as camadas densas** e simplesmente recupera o vetor latente da memória.

### Arquitetura

```
┌──────────────────────────────────────────────────────────┐
│              FORWARD PASS DIFERENCIAL                     │
│                                                           │
│  Prompt → CDC Tokenizer → [Chunk₁, Chunk₂, ..., Chunkₙ]  │
│                              │                            │
│                    Para cada Chunkᵢ:                      │
│                              │                            │
│              ┌───────────────┼───────────────┐            │
│              ▼                               ▼            │
│     Hash no Cache?                   Hash no Cache?       │
│     ┌────┐  ┌─────┐                ┌────┐  ┌─────┐       │
│     │SIM │  │ NÃO │                │SIM │  │ NÃO │       │
│     └──┬─┘  └──┬──┘                └──┬─┘  └──┬──┘       │
│        │       │                      │       │           │
│        ▼       ▼                      ▼       ▼           │
│   Recuperar  Forward                Recuperar Forward     │
│   do Cache   Pass Real              do Cache  Pass Real   │
│   (O(1))     (O(n×m))              (O(1))    (O(n×m))    │
│        │       │                      │       │           │
│        └───┬───┘                      └───┬───┘           │
│            ▼                              ▼               │
│     Vetor Latente₁                 Vetor Latente₂         │
│            │                              │               │
│            └──────────┬───────────────────┘               │
│                       ▼                                   │
│              Combinar + Delta Layer                        │
│                       ▼                                   │
│                   Resposta                                │
└──────────────────────────────────────────────────────────┘
```

### Implementação Conceitual

```go
// ActivationCache armazena vetores latentes indexados por hash CDC.
type ActivationCache struct {
    mu    sync.RWMutex
    store map[uint64][]float32  // hash → vetor latente
    lru   *LRUPolicy            // eviction policy
}

// ForwardPassDifferential executa apenas os chunks novos.
func (engine *Engine) ForwardPassDifferential(
    chunks []CDCChunk,
    cache  *ActivationCache,
) [][]float32 {
    
    activations := make([][]float32, len(chunks))
    var newChunks []int  // índices dos chunks que precisam de computação
    
    // Fase 1: Recuperar do cache o que já existe
    for i, chunk := range chunks {
        if cached, ok := cache.Get(chunk.Hash); ok {
            activations[i] = cached  // O(1) — sem computação!
        } else {
            newChunks = append(newChunks, i)
        }
    }
    
    // Fase 2: Computar APENAS os chunks novos
    for _, idx := range newChunks {
        chunk := chunks[idx]
        
        // Forward pass real — só executa para chunks inéditos
        latent := engine.model.Forward(chunk.Embedding)
        
        activations[idx] = latent
        cache.Put(chunk.Hash, latent)  // Salvar para futuro
    }
    
    return activations
}
```

---

## Cenários de Economia

### Cenário 1: Chat Conversacional
```
Turno 1: "O que é o Crompressor?"
  → 100% chunks novos → 100% forward pass real
  → Cache: 0% hit

Turno 2: "E como funciona o CDC do Crompressor?"
  → 60% chunks idênticos ao turno 1 → cache hit
  → 40% chunks novos → forward pass real
  → Economia: 60%

Turno 3: "Me mostre um exemplo de CDC do Crompressor em Go"
  → 70% chunks idênticos aos turnos anteriores → cache hit
  → 30% chunks novos → forward pass real
  → Economia: 70%

Média da conversa: ~50-70% de computação eliminada
```

### Cenário 2: Processamento de Repositório
```
Arquivo 1: main.go → cria 100 chunks → 100 forward passes
Arquivo 2: handler.go → 40% dos chunks são imports/patterns comuns → cache hit
Arquivo 3: utils.go → 60% dos chunks são patterns Go comuns → cache hit

Para um repo com 500 arquivos Go:
  → Primeiros 50 arquivos: cache ~20% hit
  → Últimos 50 arquivos: cache ~80% hit
  → O modelo "acelera" conforme vê mais código
```

---

## O Delta de Contexto

Quando um prompt é **similar** a um anterior, mas com mutações pontuais, a rede não precisa reprocessar tudo. Ela pode:

1. **Recuperar** o estado base do cache (chunks que já viu)
2. **Computar** apenas os chunks novos/alterados
3. **Combinar** via um "Delta Layer" leve

```
Estado Base (cache):   [v₁, v₂, v₃, v₄, v₅]  (5 vetores latentes)
Prompt Novo:           [v₁, v₂, v₃', v₄, v₆] (2 chunks mudaram)

Delta:                 [_, _, Δ₃, _, v₆]       (computar apenas 2 forward passes)

Combinação:            [v₁, v₂, v₃⊕Δ₃, v₄, v₆]  → Resposta
```

Isso é análogo ao `XOR Delta` do Crompressor: o padrão base vem do Codebook, e apenas a diferença é computada.

---

## Requisitos de Implementação

| Componente | Descrição | Complexidade |
|------------|-----------|-------------|
| **CDC Tokenizer** | Gerar chunks com hash determinístico | Média |
| **Activation Cache** | HashMap thread-safe com LRU eviction | Baixa |
| **Cache Key** | Hash do chunk (xxhash64) | Trivial |
| **Cache Value** | Vetor latente (float32[dim]) | ~dim×4 bytes |
| **Delta Layer** | Camada leve de combinação cache+novo | Alta |
| **Invalidação** | Quando o modelo muda, flush do cache | Baixa |

### Estimativa de Memória

```
Cache para 100K chunks:
  - Key: 8 bytes × 100K = 800KB
  - Value: 768 floats × 4 bytes × 100K = 307MB
  - Total: ~308MB de RAM para economizar ~60% de computação

Trade-off extremamente favorável.
```

---

## Diferença do KV Cache Existente

Os Transformers já usam **KV Cache** (cache de chaves/valores no attention). A diferença:

| Aspecto | KV Cache (padrão) | Activation Cache (Sinapse) |
|---------|-------------------|---------------------------|
| **Granularidade** | Por token individual | Por chunk semântico |
| **Persistência** | Vive durante uma sessão | Persistente entre sessões |
| **Indexação** | Posição sequencial | Hash de conteúdo |
| **Cross-prompt** | Não reutiliza entre prompts | Reutiliza entre quaisquer prompts |
| **Armazenamento** | RAM volátil | Pode ir para disco (mmap) |

O Activation Cache do Sinapse é **superset** do KV Cache: ele funciona *entre prompts*, *entre sessões*, e é indexado por *conteúdo*, não por posição.

---

> **Próximo:** [06 — Treinamento XOR Delta](06-TREINAMENTO_XOR_DELTA.md)
