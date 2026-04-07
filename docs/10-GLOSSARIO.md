# 10 — Glossário Técnico

> Termos técnicos unificados do ecossistema Crompressor-Sinapse.

---

## A

### Activation Cache
Cache de vetores latentes (ativações) indexado por hash CDC. Permite reutilizar computações anteriores do forward pass sem recalcular. Ver [doc 05](05-FORWARD_PASS_DIFERENCIAL.md).

---

## B

### Backpropagation
Algoritmo de treinamento que calcula gradientes da função de perda em relação a cada peso da rede, propagando o erro "de trás para frente". No Sinapse, o backprop é aplicado apenas sobre o Delta esparso, não sobre os pesos completos.

### BPE (Byte-Pair Encoding)
Tokenizador estatístico padrão de LLMs. Constrói vocabulário mesclando pares de bytes mais frequentes. Cego à estrutura semântica. O Sinapse propõe substituí-lo por CDC. Ver [doc 04](04-TOKENIZACAO_CDC.md).

### Broadcast Storm
Tempestade de transmissão. Ocorre quando todos os nós de uma rede replicam mensagens sem restrição, gerando tráfego exponencial. Ver [doc 08](08-DESCOBERTA_ROTAS.md).

---

## C

### CDC (Content-Defined Chunking)
Técnica de fragmentação que usa rolling hash para detectar fronteiras naturais no fluxo de dados. Gera chunks de tamanho variável baseados no conteúdo, não em posições fixas. Fundamental tanto para o Crompressor quanto para o Sinapse.

### Codebook
Banco de dados binário estático contendo padrões indexados (codewords). No Crompressor: padrões de bytes (`.cromdb`). No Sinapse: pode ser padrões de tokens, ativações neurais, ou pesos base.

### Codeword
Entrada individual no Codebook. No Crompressor: fragmento de 64–512 bytes extraído de datasets massivos. No VQ-NN: vetor representativo de um conceito semântico discreto.

### Commitment Loss
Componente da função de perda do VQ que força o encoder a "se comprometer" com codewords específicas, evitando colapso do Codebook. Ver [doc 07](07-VECTOR_QUANTIZATION_NN.md).

### .crom
Formato binário compacto do Crompressor. Contém header, chunk table (mapa de IDs) e delta pool. É "inútil" sem o Codebook correspondente. Ver [doc 02](02-ARQUITETURA_CROMPRESSOR.md).

### .cromdb
Formato do Codebook Universal do Crompressor. Contém header, índice HNSW, codewords e metadados. Acessado via mmap. Ver [doc 02](02-ARQUITETURA_CROMPRESSOR.md).

---

## D

### Delta (XOR Delta)
Diferença exata (XOR byte a byte) entre um chunk e o codeword mais próximo no Codebook. Propriedade fundamental: `A ⊕ B = D → B ⊕ D = A`. No Sinapse: aplicado a pesos de treinamento. Ver [doc 06](06-TREINAMENTO_XOR_DELTA.md).

### Delta Layer
Camada neural leve proposta pelo Sinapse para combinar vetores cacheados com novos vetores computados, integrando o contexto diferencial.

### Delta Stacking
Composição de múltiplos deltas via XOR para combinar skills independentes sem retreino. Possível porque `A ⊕ B ⊕ C` é associativo e comutativo.

### DHT (Distributed Hash Table)
Estrutura de dados distribuída que mapeia chaves a valores sem servidor central. Kademlia é a implementação mais usada (BitTorrent, IPFS). Usa distância XOR para roteamento. Ver [doc 08](08-DESCOBERTA_ROTAS.md).

### Dijkstra
Algoritmo que encontra o caminho mais curto em um grafo com pesos. Cada nó mantém um mapa da rede e calcula rotas localmente. Ver [doc 08](08-DESCOBERTA_ROTAS.md).

---

## F

### Flooding (Inundação)
Protocolo de descoberta de rotas onde cada nó replica a mensagem para todos os vizinhos. Infalível mas ineficiente — gera overhead de O(N²). Ver [doc 08](08-DESCOBERTA_ROTAS.md).

### Forward Pass
Propagação dos dados de entrada através de todas as camadas da rede neural para produzir saída. No Sinapse: pode ser parcial (diferencial) se chunks foram cacheados.

### Forward Pass Diferencial
Variante do forward pass onde apenas chunks novos/alterados são computados, enquanto chunks conhecidos são recuperados do Activation Cache. Ver [doc 05](05-FORWARD_PASS_DIFERENCIAL.md).

---

## G

### GGUF
Formato binário para modelos quantizados, usado pelo llama.cpp. O Sinapse intercepta a inferência GGUF antes da tokenização para injetar CDC.

### GossipSub
Protocolo de propagação de mensagens onde cada nó "fofoca" para um subconjunto fixo de vizinhos, controlando redundância. Usado em libp2p. Ver [doc 08](08-DESCOBERTA_ROTAS.md).

---

## H

### HNSW (Hierarchical Navigable Small World)
Algoritmo de busca aproximada de vizinhos mais próximos com complexidade O(log N). Usado no Crompressor para buscar codewords similares no Codebook.

---

## K

### Kademlia
Protocolo DHT que usa distância XOR entre IDs de nós para roteamento eficiente. Encontra qualquer nó em O(log N) hops. Ver [doc 08](08-DESCOBERTA_ROTAS.md).

### KV Cache
Cache de chaves/valores no mecanismo de atenção do Transformer. Diferente do Activation Cache do Sinapse: opera por posição (não por conteúdo) e não persiste entre sessões.

---

## L

### LoRA (Low-Rank Adaptation)
Técnica de fine-tuning que insere matrizes de baixo rank (A×B) para adaptar modelos sem alterar pesos originais. O Sinapse propõe XOR Delta como alternativa discreta.

### LSH (Locality-Sensitive Hashing)
Família de funções hash que preservam similaridade: dados similares geram hashes próximos. Usado no Crompressor para embeddings e no Sinapse para indexação de chunks.

### LRU (Least Recently Used)
Política de eviction para caches: quando cheio, o item menos recentemente usado é removido. Usado no Activation Cache do Sinapse.

---

## M

### MCP (Model Context Protocol)
Protocolo padronizado para agentes locais consumirem modelos de inferência. O Sinapse expõe um conector MCP para integração com agentes.

### mmap (Memory-Mapped File)
Técnica onde o SO mapeia um arquivo no espaço de endereçamento virtual, carregando páginas sob demanda. Permite acessar Codebooks de 50GB+ com ~200MB de RAM ativa.

---

## P

### PMR (Perfect Match Rate)
Percentual de chunks que correspondem exatamente a um codeword do Codebook (Delta = 0). Meta para o Crompressor: > 75%.

---

## R

### Rabin Fingerprint
Rolling hash usado para Content-Defined Chunking. Calcula um fingerprint que desliza pela janela de dados, detectando fronteiras quando o hash "casa" com uma máscara.

### Rolling Hash
Função hash que pode ser atualizada incrementalmente à medida que a janela de dados se move. Essencial para CDC eficiente.

### RVQ (Residual Vector Quantization)
Extensão do VQ onde múltiplas rodadas de quantização capturam resíduos progressivamente menores. Análogo ao Delta do Crompressor. Ver [doc 07](07-VECTOR_QUANTIZATION_NN.md).

---

## S

### SimHash
Tipo de Locality-Sensitive Hash que projeta dados em vetores binários preservando similaridade de cosseno. Usado para gerar embeddings de chunks.

### SLM (Small Language Model)
Modelos de linguagem de menor porte (1B–7B parâmetros) otimizados para inferência local. Alvo principal do Sinapse.

---

## T

### TTFT (Time To First Token)
Tempo entre o envio do prompt e a geração do primeiro token da resposta. Métrica crítica para latência de inferência. O Sinapse visa reduzir > 40%.

---

## V

### VQ (Vector Quantization)
Técnica que mapeia vetores contínuos para o vetor mais próximo em um dicionário (Codebook) finito e discreto. Ver [doc 07](07-VECTOR_QUANTIZATION_NN.md).

### VQ-NN (Vector Quantization Neural Network)
Proposta do Sinapse: aplicar VQ nas camadas ocultas do Transformer para criar representações internas discretas e interpretáveis.

### VQ-VAE (Vector Quantized Variational Autoencoder)
Arquitetura da DeepMind que introduziu VQ no espaço latente de autoencoders. Referência principal para o VQ-NN. Ver [doc 07](07-VECTOR_QUANTIZATION_NN.md).

---

## X

### XOR (Exclusive OR)
Operação lógica bit a bit: `1 ⊕ 1 = 0`, `0 ⊕ 0 = 0`, `1 ⊕ 0 = 1`. Propriedade fundamental: `A ⊕ B ⊕ B = A` (auto-inversa). Base do cálculo de Delta no Crompressor e do treinamento diferencial no Sinapse.

---

> **Início:** [00 — Índice](00-INDICE.md) | **Repositório:** [crompressor-sinapse](https://github.com/MrJc01/crompressor-sinapse)
