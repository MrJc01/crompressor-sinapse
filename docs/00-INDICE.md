# 📖 Crompressor-Sinapse — Índice da Documentação

> Mapa de navegação para toda a documentação técnica do projeto.

---

## Fundação

| # | Documento | Descrição |
|---|-----------|-----------|
| 01 | [Manifesto](01-MANIFESTO.md) | Tese central: compressão = inteligência. A visão do projeto. |
| 02 | [Arquitetura Crompressor](02-ARQUITETURA_CROMPRESSOR.md) | Resumo da engine original: CDC, Codebook, HNSW, XOR Delta. |
| 03 | [O Problema das Redes Neurais](03-PROBLEMA_REDES_NEURAIS.md) | Gargalos: BPE, forward pass redundante, custo de VRAM. |

## Pesquisa

| # | Documento | Descrição |
|---|-----------|-----------|
| 04 | [Tokenização CDC](04-TOKENIZACAO_CDC.md) | Rolling Hash como substituto do BPE — chunks semânticos. |
| 05 | [Forward Pass Diferencial](05-FORWARD_PASS_DIFERENCIAL.md) | Cache de ativações indexado por hash — computar só o delta. |
| 06 | [Treinamento XOR Delta](06-TREINAMENTO_XOR_DELTA.md) | Pesos como deltas sobre Codebook base — LoRA discreto. |
| 07 | [Vector Quantization Neural](07-VECTOR_QUANTIZATION_NN.md) | Codebook discreto dentro dos neurônios. |
| 08 | [Descoberta de Rotas](08-DESCOBERTA_ROTAS.md) | Flooding vs. eficiência em grafos distribuídos. |

## Planejamento

| # | Documento | Descrição |
|---|-----------|-----------|
| 09 | [Roadmap](09-ROADMAP.md) | Fases 1–4 do planejamento arquitetural. |
| 10 | [Glossário](10-GLOSSARIO.md) | Termos técnicos unificados do projeto. |

---

## Como Ler

**Primeira vez?** Comece pelo [Manifesto](01-MANIFESTO.md) → [Arquitetura](02-ARQUITETURA_CROMPRESSOR.md) → [Problema](03-PROBLEMA_REDES_NEURAIS.md).

**Pesquisador?** Vá direto para os docs 04–08, cada um é autocontido.

**Implementador?** O [Roadmap](09-ROADMAP.md) detalha as 4 fases de execução.
