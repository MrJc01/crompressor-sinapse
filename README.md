<p align="center">
  <strong>🧠 Crompressor-Sinapse</strong>
</p>

<p align="center">
  <em>"Quando a compressão se torna cognição."</em>
</p>

<p align="center">
  <a href="https://github.com/MrJc01/crompressor">Motor Crompressor</a> ·
  <a href="docs/00-INDICE.md">Documentação</a> ·
  <a href="docs/09-ROADMAP.md">Roadmap</a>
</p>

---

## O Que É Isto?

**Crompressor-Sinapse** é um projeto de pesquisa que investiga a fusão entre o motor de compressão semântica [Crompressor](https://github.com/MrJc01/crompressor) e o ciclo de treinamento/inferência de redes neurais.

A tese central é que **compressão e inteligência são o mesmo processo matemático.** Se o Crompressor ataca redundância estrutural no nível do byte (CDC, Codebook, XOR Delta), essas mesmas primitivas podem ser aplicadas dentro do processamento neural para eliminar computação desperdiçada.

## Frentes de Pesquisa

| # | Frente | Descrição | Status |
|---|--------|-----------|--------|
| 1 | [Tokenização CDC](docs/04-TOKENIZACAO_CDC.md) | Substituir BPE por Content-Defined Chunking como porta de entrada da rede | 🔬 Pesquisa |
| 2 | [Forward Pass Diferencial](docs/05-FORWARD_PASS_DIFERENCIAL.md) | Cache de ativações indexado por hash CDC — computar apenas o delta | 🔬 Pesquisa |
| 3 | [Treinamento XOR Delta](docs/06-TREINAMENTO_XOR_DELTA.md) | Pesos como deltas sobre Codebook base — LoRA discreto | 🔬 Pesquisa |
| 4 | [Vector Quantization Neural](docs/07-VECTOR_QUANTIZATION_NN.md) | Codebook discreto nos neurônios, eliminando ruído contínuo | 🔬 Pesquisa |
| 5 | [Descoberta de Rotas](docs/08-DESCOBERTA_ROTAS.md) | Flooding vs. eficiência em grafos distribuídos | 🔬 Pesquisa |

## Fundação

O Sinapse herda a arquitetura do Crompressor original:

```
Dados → CDC Chunking → Busca no Codebook (HNSW) → XOR Delta → .crom
                              ↑
                  Codebook Universal (50GB+)
                  Nunca descartado. Sempre disponível.
```

A diferença é o **domínio de aplicação**: em vez de comprimir arquivos, comprimimos o *processamento* neural.

## Documentação Completa

📖 Consulte o [Índice da Documentação](docs/00-INDICE.md) para o mapa completo.

## Relação com o Ecossistema

```
crompressor (engine Go)
    └── crompressor-sinapse (pesquisa: compressão × NN)
```

## Licença

[MIT](LICENSE) © 2026 MrJc01

---

> *"Nós não comprimimos dados. Nós indexamos o universo."*
