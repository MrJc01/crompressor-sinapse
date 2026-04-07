# 02 — Arquitetura do Crompressor (Resumo)

> *"A compressão tradicional esquece tudo ao terminar. O CROM lembra de tudo antes de começar."*

---

Este documento resume a arquitetura da engine [Crompressor](https://github.com/MrJc01/crompressor) original — a fundação sobre a qual o Sinapse constrói.

## O Paradigma: Compressão Baseada em Conhecimento

Compressores tradicionais (Gzip, Zstd, LZ4) operam com **amnésia total**. A cada arquivo, começam do zero.

```
PARADIGMA TRADICIONAL (LZ77/Gzip):
Arquivo → Análise de Frequência → Dicionário Local (32KB) → Codificação → .gz
                                          ↑
                                   (Descartado após uso)

PARADIGMA CROM:
Arquivo → CDC Chunking → Busca no Codebook Universal → Mapa de IDs + Delta → .crom
                                 ↑
                    50GB+ de Padrões Permanentes
                    Nunca descartado. Sempre disponível.
```

O CROM inverte a lógica:
1. **O conhecimento vem primeiro.** O Codebook já existe antes do arquivo.
2. **A compressão é uma busca, não uma construção.**
3. **O resíduo (Delta) garante fidelidade lossless.**

---

## Pipeline: Pack → Transfer → Unpack

```
┌──────────────────────────────────────────────────────────┐
│                     CROM RUNTIME                          │
│                                                           │
│  ┌─────────────┐    ┌──────────────┐    ┌──────────────┐ │
│  │  crom-pack   │◄──▶│   Codebook    │◄──▶│  crom-unpack  │ │
│  │ (Compilador) │    │  (.cromdb)   │    │(Decompilador)│ │
│  └──────┬──────┘    │   50GB+      │    └──────┬───────┘ │
│         │           └──────────────┘           │          │
│         ▼                                      ▼          │
│  ┌──────────────┐                      ┌──────────────┐  │
│  │ Arquivo .crom│ ──────────────────▶  │   Original   │  │
│  │  (Compacto)  │    Transferência     │ (Restaurado) │  │
│  └──────────────┘                      └──────────────┘  │
└──────────────────────────────────────────────────────────┘
```

---

## Etapa 1: Chunking (CDC)

O arquivo é fragmentado em **chunks** com fronteiras naturais detectadas via rolling hash (Rabin fingerprint):

```
Arquivo Original (N bytes)
        │
        ▼
┌──────────────────────────────┐
│      CHUNKING ENGINE         │
│                              │
│  ├─ Tamanho fixo (128 bytes) │
│  ├─ Content-Defined (Rabin)  │
│  └─ Tipo-específico          │
│                              │
│  Output: [C₁, C₂, ..., Cₙ]  │
└──────────────────────────────┘
```

O CDC é **fundamental** para o Sinapse: a mesma ideia será aplicada a texto e código.

---

## Etapa 2: Busca no Codebook (HNSW)

Para cada chunk, o motor busca o **codeword mais similar** no Codebook:

```
Chunk Cᵢ → SimHash Embedding → HNSW Search (O(log N))
                                     │
                                     ▼
                              { codebook_id, similarity, pattern }
```

O Codebook de 50GB é acessado via **mmap** — o OS carrega apenas as páginas necessárias (~200MB ativos de 50GB mapeados).

---

## Etapa 3: Delta (XOR Lossless)

A diferença exata entre o chunk e o codeword é calculada com **XOR**:

```
Chunk Original:    [0xFA, 0x3C, 0x92, 0x10, 0xFF, 0x88, 0x44]
Codeword:          [0xFA, 0x3C, 0x92, 0x10, 0xFF, 0x8A, 0x44]
Delta (XOR):       [0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0x00]
                    → Maioria zeros → comprime extraordinariamente com Zstd
```

**Propriedade fundamental:** `A ⊕ B = D` → `B ⊕ D = A` — reversível e determinístico.

---

## Formato `.crom`

```
┌──────────────────────────────────────┐
│          FORMATO .crom                │
│                                       │
│  Header:                              │
│  ├─ Magic: "CROM" (4 bytes)          │
│  ├─ Codebook Hash: SHA-256           │
│  ├─ Original Hash: SHA-256           │
│  ├─ Original Size: uint64            │
│  └─ Chunk Count: uint32              │
│                                       │
│  Chunk Table:                         │
│  ├─ [0] { codebook_id, delta, size } │
│  ├─ [1] { ... }                      │
│  └─ [N] { ... }                      │
│                                       │
│  Delta Pool: zstd(deltas[0..N])       │
│                                       │
│  Footer: CRC32                        │
└──────────────────────────────────────┘
```

---

## Performance Assimétrica

| Operação | Complexidade | Velocidade |
|----------|-------------|------------|
| **Pack** (Compilação) | O(n × log M) | ~25 MB/s |
| **Unpack** (Decompilação) | O(n) | ~200 MB/s |

A compilação é deliberadamente mais lenta (busca HNSW). A decompilação é **quase instantânea** (lookups diretos + XOR).

---

## Relevância para o Sinapse

Os componentes que o Sinapse reutiliza:

| Componente Crompressor | Aplicação no Sinapse |
|------------------------|----------------------|
| CDC Chunking | Tokenização de texto/código |
| Codebook | Memória persistente de ativações neurais |
| HNSW Search | Recuperação rápida de contexto cacheado |
| XOR Delta | Computação diferencial no forward pass |
| mmap | Acesso a pesos/Codebook sem carregar na RAM |

---

> **Próximo:** [03 — O Problema das Redes Neurais](03-PROBLEMA_REDES_NEURAIS.md)
