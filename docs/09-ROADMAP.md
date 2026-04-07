# 09 — Roadmap: Planejamento Arquitetural

> *"4 fases. Do rolling hash ao daemon de inferência local."*

---

## Visão Geral

```
┌──────────┐    ┌──────────────┐    ┌──────────────┐    ┌──────────────────┐
│  FASE 1  │───▶│   FASE 2     │───▶│   FASE 3     │───▶│     FASE 4       │
│ Fundação │    │ Interceptador│    │   Cálculo    │    │  Integração      │
│   CDC    │    │  de Inferência│    │ Diferencial  │    │  Local-First     │
│          │    │              │    │              │    │                  │
│ Go CLI   │    │ GGUF Wrapper │    │ XOR Delta    │    │ Daemon + API     │
│ Rolling  │    │ Activation   │    │ Benchmarks   │    │ MCP + P2P        │
│ Hash     │    │ Cache        │    │ TTFT         │    │ Codebook Dist.   │
└──────────┘    └──────────────┘    └──────────────┘    └──────────────────┘
   Semanas          Semanas             Semanas              Semanas
    1-4               5-8                9-12                13-16
```

---

## Fase 1: Fundação & Tokenização Semântica (CDC Layer)

**Objetivo:** Validar que o CDC gera menos unidades lógicas que o BPE para texto e código, usando Go puro.

### Tarefas

| # | Tarefa | Critério de Sucesso |
|---|--------|---------------------|
| 1.1 | Implementar Rolling Hash (Rabin/Buzhash) otimizado para texto natural e código-fonte | Fronteiras detectadas em < 1ms por KB |
| 1.2 | Geração do Codebook Dinâmico | Tabela de hash em memória registra chunks inéditos |
| 1.3 | CLI `sinapse-tokenize` | `sinapse-tokenize --input ./repo/ --output tokens.json` |
| 1.4 | Benchmark BPE vs CDC | Rodar em 1GB de código Go, medir ratio tokens/chunks |
| 1.5 | Validação de Densidade | CDC gera 3-8x menos unidades que BPE cl100k_base |
| 1.6 | Validação de Estabilidade | Edit de 5 linhas → >90% chunks inalterados |

### Stack

```
crompressor-sinapse/
├── cmd/
│   └── sinapse-tokenize/
│       └── main.go          # CLI de tokenização CDC
├── internal/
│   ├── cdc/
│   │   ├── rabin.go         # Rolling hash (Rabin fingerprint)
│   │   ├── buzhash.go       # Alternativa: Buzhash
│   │   ├── tokenizer.go     # Interface + lógica CDC para texto
│   │   └── cdc_test.go
│   └── codebook/
│       ├── dynamic.go       # Codebook dinâmico em memória
│       └── codebook_test.go
├── go.mod
└── Makefile
```

### Validação

```bash
# Rodar contra corpus denso em português
sinapse-tokenize --input ./corpus-pt/ --strategy rabin --avg-size 128

# Esperado:
# ╔═══════════════════════════════════════════╗
# ║     CDC TOKENIZATION REPORT              ║
# ╠═══════════════════════════════════════════╣
# ║  Input:         ./corpus-pt/ (1.2GB)     ║
# ║  BPE tokens:    41,203,847               ║
# ║  CDC chunks:    8,240,769                ║
# ║  Ratio:         5.0x menos unidades      ║
# ║  Unique chunks: 1,203,445 (14.6%)        ║
# ║  Dedup rate:    85.4%                    ║
# ╚═══════════════════════════════════════════╝
```

---

## Fase 2: O Interceptador de Inferência (Forward Cache)

**Objetivo:** Plugar o motor de compressão na mecânica de forward pass de um SLM (Small Language Model).

### Tarefas

| # | Tarefa | Critério de Sucesso |
|---|--------|---------------------|
| 2.1 | Wrapper de Inferência GGUF | Camada que recebe prompt → CDC → modelo |
| 2.2 | Activation Cache (HashMap) | Cache thread-safe com LRU, indexado por hash CDC |
| 2.3 | Mapeamento de Ativação | Salvar vetor latente associado ao hash do chunk |
| 2.4 | Bypass de Computação | Se hash no cache → pular forward, injetar vetor |
| 2.5 | Benchmark cache hit rate | Medir hit rate em conversa de 10 turnos |
| 2.6 | Integração com llama.cpp | Interceptar prompt antes de tokenização nativa |

### Arquitetura

```
┌──────────────────────────────────────────────────┐
│              INTERCEPTADOR SINAPSE                │
│                                                   │
│  Prompt do Usuário                                │
│       │                                           │
│       ▼                                           │
│  CDC Tokenizer (Fase 1)                           │
│       │                                           │
│       ├── Chunk já no cache? ──▶ SIM: retornar    │
│       │                         vetor latente     │
│       │                                           │
│       └── Chunk novo? ──▶ NÃO: enviar para        │
│           llama.cpp/GGUF ──▶ computar ──▶ cachear │
│                                                   │
│  Montar resposta com vetores (cache + novos)      │
│       │                                           │
│       ▼                                           │
│  Resposta ao Usuário                              │
└──────────────────────────────────────────────────┘
```

### Formato de Cache

```go
type CacheEntry struct {
    Hash       uint64      // xxhash do chunk CDC
    Latent     []float32   // Vetor latente (dim=768 ou similar)
    Layer      int         // Camada da rede onde foi cacheado
    Timestamp  int64       // Para LRU eviction
    HitCount   uint32      // Frequência de uso
}
```

---

## Fase 3: Cálculo Diferencial (XOR Delta Neural)

**Objetivo:** Transição da deduplicação pura para a inteligência de adaptação.

### Tarefas

| # | Tarefa | Critério de Sucesso |
|---|--------|---------------------|
| 3.1 | Isolamento do Delta | Quando prompt é similar a contexto em cache, computar apenas a diferença |
| 3.2 | Delta Layer | Camada leve que combina vetores cacheados + novos |
| 3.3 | Benchmark TTFT | Medir Time To First Token com/sem cache |
| 3.4 | Benchmark latência | Em CPU sem GPU, medir speedup do cache vs. recomputação |
| 3.5 | Validação de qualidade | Respostas com cache ≈ respostas sem cache (BLEU score) |
| 3.6 | Análise de trade-off | Gráfico: RAM do cache × speedup × qualidade |

### Hipótese a Provar

```
O custo do cache em RAM/SSD para o Codebook de ativações é
ORDENS DE GRANDEZA menor que o custo de recomputar a mesma
lógica no processador.

Cache de 300MB → economia de 60% de computação no forward pass
→ TTFT reduzido de 2s para 0.8s em CPU consumer
```

---

## Fase 4: Integração com o Ecossistema Local-First

**Objetivo:** Transformar a prova de conceito em ferramenta de infraestrutura soberana.

### Tarefas

| # | Tarefa | Critério de Sucesso |
|---|--------|---------------------|
| 4.1 | CLI & Daemon | `sinapse-daemon` rodando em background, baixo consumo |
| 4.2 | API HTTP local | Endpoint `POST /v1/completions` compatível com OpenAI API |
| 4.3 | Conector MCP | Model Context Protocol para agentes locais |
| 4.4 | Persistência de Codebook | Salvar/carregar Codebook de ativações do disco |
| 4.5 | Codebook distribuído (P2P) | Codebooks portáteis via rede de grafos |
| 4.6 | Documentação de deployment | README final com instruções completas |

### Topologia Final

```
┌─────────────────────────────────────────────────────┐
│                 MÁQUINA DO USUÁRIO                   │
│                                                      │
│  ┌──────────────┐    ┌──────────────────────────┐   │
│  │  sinapse-     │    │   Activation Codebook    │   │
│  │  daemon       │◄──▶│   (.sinapse.db)          │   │
│  │  (Go binary)  │    │   mmap'd, persistente    │   │
│  └──────┬───────┘    └──────────────────────────┘   │
│         │                                            │
│         ├── HTTP API (localhost:8080)                 │
│         ├── MCP Connector (stdio)                    │
│         └── P2P Mesh (libp2p, opcional)              │
│                                                      │
│  Agentes Locais ──▶ sinapse-daemon ──▶ GGUF Model   │
│  IDE / Terminal ──▶ sinapse-daemon ──▶ CDC Cache     │
│  Outros nós P2P ──▶ sinapse-daemon ──▶ Codebook Sync│
└─────────────────────────────────────────────────────┘
```

---

## Priorização

### 🔴 P0 — Crítico (Fase 1)
- [ ] Rolling Hash (Rabin) para texto
- [ ] Codebook dinâmico em memória
- [ ] CLI `sinapse-tokenize`
- [ ] Benchmark BPE vs CDC
- [ ] Testes unitários

### 🟡 P1 — Importante (Fase 2)
- [ ] Wrapper de inferência GGUF
- [ ] Activation Cache (HashMap + LRU)
- [ ] Bypass de computação
- [ ] Integração llama.cpp

### 🟢 P2 — Desejável (Fase 3)
- [ ] Delta Layer
- [ ] Benchmarks TTFT
- [ ] Validação de qualidade (BLEU)
- [ ] Análise de trade-off

### 🔵 P3 — Futuro (Fase 4)
- [ ] Daemon background
- [ ] API HTTP compatível OpenAI
- [ ] Conector MCP
- [ ] P2P Codebook sync
- [ ] Documentação de deployment

---

## Métricas de Sucesso por Fase

| Fase | Métrica | Meta |
|------|---------|------|
| 1 | CDC/BPE ratio | > 3x menos unidades |
| 1 | Estabilidade CDC | > 90% chunks inalterados após edit |
| 2 | Cache hit rate (10 turnos) | > 50% |
| 2 | Overhead do cache | < 500MB RAM |
| 3 | Redução TTFT | > 40% |
| 3 | Qualidade (BLEU vs baseline) | > 0.95 |
| 4 | Uptime do daemon | > 99.9% |
| 4 | Latência API local | < 100ms overhead |

---

> **Próximo:** [10 — Glossário](10-GLOSSARIO.md)
