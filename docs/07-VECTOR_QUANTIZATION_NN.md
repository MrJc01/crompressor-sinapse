# 07 — Vector Quantization nos Neurônios (VQ-NN)

> *"A rede não precisa de precisão infinita. Ela precisa de um vocabulário interno."*

---

## O Paradigma: Contínuo vs. Discreto

Nos layers ocultos de uma rede neural tradicional, cada neurônio produz um **vetor de ponto flutuante contínuo**:

```
Saída do neurônio:  [0.84234, -0.21073, 0.99121, 0.00342, -0.77882]
                     ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
                     Precisão arbitrária. Maioria é ruído matemático.
```

A proposta do VQ-NN é forçar cada neurônio a **mapear sua saída para o vetor mais próximo de um Codebook aprendido** — criando um "alfabeto" interno discreto.

---

## Como Funciona

### Sem VQ (padrão)
```
Input → Layer → [0.84234, -0.21073, 0.99121] → Próximo Layer
                 ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
                 Vetor contínuo com precisão arbitrária
```

### Com VQ-NN (Sinapse)
```
Input → Layer → [0.84234, -0.21073, 0.99121]
                          │
                          ▼
                ┌─────────────────────┐
                │  CODEBOOK INTERNO    │
                │                      │
                │  [0] = [0.85, -0.21, 1.00]  ← Mais próximo! ✓
                │  [1] = [0.50, 0.30, -0.70]
                │  [2] = [-0.90, 0.80, 0.10]
                │  [3] = [0.20, -0.60, 0.40]
                │  ...                 │
                └─────────────────────┘
                          │
                          ▼
                Output: ID=0 → [0.85, -0.21, 1.00]
                → Vetor discreto. Limpo. Sem ruído.
```

---

## Analogia com o Crompressor

```
CROMPRESSOR:
  Chunk de bytes → Busca no Codebook → ID do padrão mais próximo
  → Armazena apenas o ID + delta residual

VQ-NN:
  Vetor de ativação → Busca no Codebook Neural → ID do conceito mais próximo
  → Propaga apenas o vetor quantizado (sem delta, pois é treinável)
```

A rede desenvolve um **dicionário de conceitos** interno. Assim como o Codebook do Crompressor contém padrões binários universais, o Codebook do VQ-NN contém **representações semânticas discretas** aprendidas.

---

## Implementação Conceitual

```go
// VQLayer implementa uma camada com Vector Quantization.
type VQLayer struct {
    Codebook  [][]float32  // K vetores de dimensão D
    K         int           // Tamanho do dicionário (ex: 512, 1024)
    D         int           // Dimensão dos vetores
    commitment float32     // Peso do commitment loss
}

// Forward mapeia a ativação contínua para o codeword mais próximo.
func (vq *VQLayer) Forward(z []float32) (quantized []float32, idx int) {
    // 1. Encontrar o codeword mais próximo (distância L2)
    minDist := float32(math.MaxFloat32)
    bestIdx := 0
    
    for i, cw := range vq.Codebook {
        dist := l2Distance(z, cw)
        if dist < minDist {
            minDist = dist
            bestIdx = i
        }
    }
    
    // 2. Retornar o codeword quantizado
    return vq.Codebook[bestIdx], bestIdx
}

// Loss combina reconstrução + commitment (força o encoder a se comprometer)
func (vq *VQLayer) Loss(z, quantized []float32) float32 {
    // VQ Loss: ||sg[z] - e||² + β||z - sg[e]||²
    // sg = stop gradient (não propaga gradiente para o outro termo)
    codebookLoss := l2Distance(stopGrad(z), quantized)
    commitLoss := l2Distance(z, stopGrad(quantized))
    return codebookLoss + vq.commitment*commitLoss
}
```

---

## Vantagens do VQ-NN

### 1. Eliminação de Ruído

```
Sem VQ:  [0.84234, -0.21073, 0.99121]  → 12 bytes (3 × float32)
Com VQ:  ID=42                          → 2 bytes (uint16)
         → Codebook[42] decodifica para [0.85, -0.21, 1.00]
         
→ 6x menos bandwidth entre camadas
→ Sinal semântico puro, sem ruído numérico
```

### 2. Interpretabilidade

Com VQ, cada ativação é um **ID discreto** que aponta para um conceito aprendido. Isso permite:

```
Camada 5, Neurônio 23 ativa Codeword #42 quando vê:
  - "declaração de função em Go"
  - "function keyword em JavaScript"  
  - "def statement em Python"

→ Codeword #42 REPRESENTA o conceito "declaração de função"!
→ A rede torna-se interpretável.
```

### 3. Compressão de Modelo

Se cada ativação é substituída por um ID de 16 bits:

```
Ativações float32:   768 dims × 4 bytes = 3072 bytes por token
Ativações VQ (ID):   1 ID × 2 bytes = 2 bytes por token

→ 1536x menos armazenamento de ativações
→ Viável para inferência em hardware ultra-limitado
```

---

## Estado da Arte: Onde o VQ Já Funciona

O VQ não é teoria — já demonstra sucesso em múltiplos domínios:

| Sistema | Domínio | Uso do VQ |
|---------|---------|-----------|
| **VQ-VAE** (DeepMind) | Geração de imagens | Codebook latente de patches visuais |
| **SoundStream** (Google) | Compressão de áudio | Múltiplos codebooks hierárquicos |
| **EnCodec** (Meta) | Codec neural de áudio | RVQ (Residual Vector Quantization) |
| **DALL-E 1** (OpenAI) | Geração text-to-image | dVAE com Codebook de 8192 entries |

O Sinapse propõe levar o VQ para as **camadas ocultas do Transformer** — não apenas no encoder/decoder de mídia, mas no processamento linguístico interno.

---

## Residual Vector Quantization (RVQ)

Para capturar nuances que um único Codebook não alcança, o RVQ aplica **múltiplas rodadas** de quantização:

```
Ativação z:         [0.84, -0.21, 0.99]

Rodada 1:  Codebook₁  → cw₁ = [0.85, -0.20, 1.00]
           Resíduo₁    = z - cw₁ = [-0.01, -0.01, -0.01]

Rodada 2:  Codebook₂  → cw₂ = [-0.01, -0.01, -0.01]  (match perfeito!)
           Resíduo₂    = 0

Reconstrução: cw₁ + cw₂ = [0.84, -0.21, 0.99] = z  (lossless!)

Armazenamento: [ID₁=42, ID₂=7]  → 4 bytes total
```

Isso é **análogo** ao Delta do Crompressor: o primeiro Codebook captura o padrão base, e o segundo captura o resíduo.

---

## Impacto no Ecossistema

```
VQ-NN + CDC Tokenização + Activation Cache + XOR Delta Training
    = Modelo neural que opera inteiramente em domínio discreto
    = Hardware-friendly (inteiros, lookup tables, sem float)
    = Comprimível pelo próprio Crompressor!
```

A tese completa do Sinapse: se toda a cadeia neural opera com **IDs discretos** e **deltas esparsos**, o modelo inteiro se torna um **artefato comprimível** — e o Crompressor pode servir como o formato nativo de armazenamento e transmissão de modelos de IA.

---

> **Próximo:** [08 — Descoberta de Rotas](08-DESCOBERTA_ROTAS.md)
