# 06 — Treinamento por XOR Delta

> *"Não treine o modelo inteiro. Treine apenas a diferença."*

---

## O Problema: Backpropagation É Caro

O gargalo do treinamento neural é o custo em VRAM para calcular e aplicar gradientes em matrizes de ponto flutuante imensas:

```
Modelo 7B parâmetros:
  Pesos:       28 GB (float32)
  Gradientes:  28 GB
  Otimizador:  56 GB (Adam: 2 estados por peso)
  ─────────────────────────────
  Total VRAM: 112 GB  ← Inacessível para hardware consumer
```

### Soluções Atuais e Suas Limitações

| Técnica | O Que Faz | Limitação |
|---------|-----------|-----------|
| **LoRA** | Treina matrizes de baixo rank (A×B) | Ainda opera em float contínuo |
| **QLoRA** | LoRA + quantização 4-bit | Perda de precisão significativa |
| **Pruning** | Remove pesos "inúteis" | Irreversível, pode degradar |

---

## A Proposta: Pesos como Deltas sobre Codebook Base

Imagine inicializar um modelo com um **Codebook imutável de matrizes de pesos universais** — como um dicionário base. Durante o treinamento de uma nova habilidade, em vez de atualizar essas matrizes gigantes, os neurônios aprendem a gerar apenas um **mapa de XOR deltas** ou **máscaras de bits esparsas**.

### Mecânica

```
┌──────────────────────────────────────────────────────┐
│           TREINAMENTO XOR DELTA                       │
│                                                       │
│  1. ESTADO INICIAL                                    │
│     Codebook de Pesos Base: W_base (imutável, mmap'd) │
│     Delta: Δ = zeros (esparso)                        │
│                                                       │
│  2. FORWARD PASS                                      │
│     W_efetivo = W_base ⊕ Δ                            │
│     output = input × W_efetivo                        │
│                                                       │
│  3. BACKWARD PASS                                     │
│     grad = ∂L/∂W_efetivo                              │
│     Δ_novo = quantize(Δ + lr × grad)                  │
│     → Atualiza APENAS o Δ, NUNCA o W_base             │
│                                                       │
│  4. ARMAZENAMENTO                                     │
│     Salvar: Δ (esparso, poucos KB-MB)                 │
│     NÃO salvar: W_base (50GB, já está no Codebook)    │
└──────────────────────────────────────────────────────┘
```

### Analogia Direta com o Crompressor

```
CROMPRESSOR:
  Dado Original  ⊕  Codebook Pattern  =  Delta (poucos bytes)
  Codebook Pattern  ⊕  Delta  =  Dado Original Restaurado

SINAPSE (TREINAMENTO):
  Pesos Treinados  ⊕  Pesos Base (Codebook)  =  Delta (esparso)
  Pesos Base  ⊕  Delta  =  Pesos Treinados Restaurados
```

---

## Implementação Conceitual

```go
// WeightCodebook é o dicionário imutável de matrizes de pesos.
type WeightCodebook struct {
    blocks []WeightBlock  // Cada bloco: [64×64] float16
    index  map[uint64]int // Hash → índice do bloco
}

// DeltaMap armazena APENAS as diferenças do treinamento.
type DeltaMap struct {
    entries map[int]SparseDelta  // Índice do bloco → delta esparso
}

// SparseDelta representa uma máscara de bits com as mudanças.
type SparseDelta struct {
    Mask    []uint64   // Bitmap: quais posições mudaram
    Values  []int8     // Valores quantizados das mudanças
}

// Reconstruct restaura os pesos efetivos para inferência.
func Reconstruct(codebook *WeightCodebook, delta *DeltaMap) []float16 {
    weights := make([]float16, codebook.TotalSize())
    
    for blockIdx, block := range codebook.blocks {
        copy(weights[blockIdx*64*64:], block.Data)
        
        if d, ok := delta.entries[blockIdx]; ok {
            // Aplicar delta esparso — apenas nas posições que mudaram
            for pos, val := range d.NonZeroEntries() {
                weights[blockIdx*64*64+pos] += float16(val) * scaleFactor
            }
        }
    }
    
    return weights
}
```

---

## Vantagens sobre LoRA

| Aspecto | LoRA Tradicional | XOR Delta (Sinapse) |
|---------|------------------|---------------------|
| **Representação** | Float32 contínuo (A×B) | Bits discretos (máscara + int8) |
| **Armazenamento** | ~50-200MB por adapter | ~1-10MB por delta |
| **Composição** | Complexa (merge de matrizes) | Trivial (XOR de deltas) |
| **Reversibilidade** | Destrutiva (merge permanente) | Reversível (remove o delta) |
| **Múltiplas skills** | Múltiplos adapters separados | Delta stacking (OR de máscaras) |

### Composição de Skills via Delta Stacking

```
Codebook Base: W                (modelo generalista)
Delta Skill A: Δ_go             (especialização Go)
Delta Skill B: Δ_security       (especialização segurança)

Modelo Go + Security:
  W_final = W ⊕ Δ_go ⊕ Δ_security
  → Duas skills compostas com XOR, sem retreino!

Remover skill Go:
  W_sem_go = W_final ⊕ Δ_go
  → Skill removida instantaneamente (XOR é auto-inverso)
```

---

## Redução de VRAM

```
TREINAMENTO TRADICIONAL (7B):
  Pesos:       28 GB
  Gradientes:  28 GB
  Otimizador:  56 GB
  Total:       112 GB

TREINAMENTO XOR DELTA:
  Codebook:    mmap (200MB ativos de 28GB mapeados)
  Delta:       10 MB (esparso, int8)
  Gradientes:  10 MB (apenas sobre o Delta)
  Total:       ~220 MB  ← Cabe em qualquer GPU consumer!
```

---

## Processo de Treinamento

```
┌──────────────────────────────────────────────────────┐
│               LOOP DE TREINAMENTO                     │
│                                                       │
│  for each batch:                                      │
│    1. W = codebook.Lookup(block_ids) ⊕ delta          │
│    2. output = forward(input, W)                      │
│    3. loss = criterion(output, target)                 │
│    4. grad = backward(loss)                           │
│                                                       │
│    5. delta_grad = quantize_to_int8(grad)             │
│    6. delta = delta ⊕ delta_grad                      │
│       └─ APENAS as posições com gradiente > threshold │
│          são atualizadas na máscara esparsa           │
│                                                       │
│  Salvar:                                              │
│    - delta.bin (~5MB)                                  │
│    - Referência ao Codebook: hash do bloco base       │
└──────────────────────────────────────────────────────┘
```

---

## Riscos e Mitigações

| Risco | Impacto | Mitigação |
|-------|---------|-----------|
| Quantização destrutiva | Perda de informação no delta | Escalar os int8 com fator adaptativo |
| Codebook desalinhado | Deltas imensos (longe do base) | Atualizar Codebook periodicamente |
| Esparsidade insuficiente | Delta não comprime bem | Threshold adaptativo de gradiente |
| Composição de deltas | Acúmulo de erros | Validação SHA-256 a cada merge |

---

> **Próximo:** [07 — Vector Quantization Neural](07-VECTOR_QUANTIZATION_NN.md)
