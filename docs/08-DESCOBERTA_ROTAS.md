# 08 — Descoberta de Rotas: Flooding vs. Eficiência

> *"Se todos falarem, a melhor rota será inevitavelmente trilhada no meio do caos. Mas o caos não escala."*

---

## A Intuição Original

Se o nó A0 precisa falar com o nó A100 em uma rede de 100 nós, e A0 manda a mensagem para **todos** os seus vizinhos, que mandam para todos os vizinhos deles, e assim por diante — no final, **todas as rotas possíveis terão sido percorridas**.

A primeira vez que a mensagem chegar no A100, ela obrigatoriamente terá viajado pela **rota mais rápida** (em termos de latência naquele momento).

Essa é a forma mais primitiva e infalível de descoberta de caminhos em redes: o **Flooding** (Inundação).

---

## O Problema: Broadcast Storm

Embora a melhor rota tenha sido descoberta na prática, o custo é absurdo:

```
Rede de 100 nós, cada nó com 10 vizinhos:

Mensagem original:                    1
Depois do hop 1:                     10
Depois do hop 2:                    100
Depois do hop 3:                  1,000
Depois do hop 4:                 10,000
Depois do hop 5:                100,000  ← Mensagens redundantes!

Total de mensagens para descobrir 1 rota: ~111,111
Mensagens ÚTEIS: 100 (1 por nó)
Overhead: 99.9%
```

A rede **engasga e cai**. O custo computacional e o consumo de banda destroem o ecossistema.

---

## As Soluções Engenhadas

Redes descentralizadas não deixam todos os nós falarem ao mesmo tempo. Elas aplicam **protocolos** para encontrar a melhor rota sem explodir a rede:

### 1. Gossip Protocols (GossipSub)

```
Filosofia: "Fofoca seletiva"

Em vez de gritar para TODOS:
  A0 → conta para 3 vizinhos aleatórios
  Cada vizinho → conta para 3 vizinhos aleatórios
  
  Propagação: O(log N) hops para atingir toda a rede
  Redundância: controlada (~3x em vez de ~1000x)
  
┌───┐     ┌───┐     ┌───┐
│ A0│────▶│ A3│────▶│A12│────▶ ...
│   │────▶│ A7│────▶│A28│────▶ ...
│   │────▶│A15│────▶│A44│────▶ ...
└───┘     └───┘     └───┘
  3         9        27       81 ... N
```

**Usado em:** libp2p (IPFS), Ethereum 2.0 (beacon chain)

### 2. DHT — Tabelas de Hash Distribuídas (Kademlia)

```
Filosofia: "Endereçamento matemático"

Cada nó tem um ID binário (ex: 160 bits).
O A0 não manda a mensagem em todas as direções
— ele calcula matematicamente quais vizinhos têm
um ID mais "próximo" (distância XOR) do A100.

A0 (ID: 0000) quer A100 (ID: 1100):
  Distância XOR: 0000 ⊕ 1100 = 1100
  → Buscar vizinho com prefixo 1... (metade mais próxima)
  → A50 (ID: 1000) é o intermediário
  → A50 busca vizinho com prefixo 11... 
  → A80 (ID: 1100) = A100 encontrado!

Hops: O(log N) = ~7 hops para 100 nós
```

**Nota:** A distância Kademlia usa **XOR** — a mesma operação que o Crompressor usa para calcular deltas.

**Usado em:** BitTorrent, IPFS, Ethereum (node discovery)

### 3. Algoritmos de Estado de Link (Dijkstra)

```
Filosofia: "Mapa compartilhado"

Em vez de descobrir rotas enviando mensagens completas,
os nós trocam constantemente pacotes MINÚSCULOS (metadados)
informando seus vizinhos e custos de link.

Cada nó constrói um MAPA INTERNO de toda a rede.
Quando A0 precisa falar com A100:
  → Calcula a melhor rota LOCALMENTE (Dijkstra)
  → Dispara a mensagem EXCLUSIVAMENTE por aquele caminho

                    ┌────────────────────┐
                    │   MAPA LOCAL do A0  │
                    │                    │
                    │  A0──A3──A12──A100  │ ← Melhor rota
                    │  A0──A7──A28──A100  │    (custo: 4)
                    │  A0──A15─A44──A100  │
                    └────────────────────┘
```

**Usado em:** OSPF (Internet backbone), IS-IS

---

## Comparação

| Aspecto | Flooding | GossipSub | DHT (Kademlia) | Dijkstra |
|---------|----------|-----------|----------------|----------|
| **Mensagens** | O(N²) | O(N log N) | O(log N) | O(1) |
| **Latência** | Mínima (brute force) | Baixa | Média | Ótima |
| **Estado necessário** | Nenhum | Nenhum | Tabela de rotas | Mapa completo |
| **Tolerância a falhas** | Máxima | Alta | Alta | Média |
| **Escalabilidade** | Péssima | Boa | Excelente | Boa |

---

## Relevância para o Sinapse

Os mesmos trade-offs de roteamento de mensagens em redes distribuídas aparecem no roteamento de **informação dentro de redes neurais**:

### Analogia: Atenção = Roteamento

```
REDE DISTRIBUÍDA:
  A0 precisa encontrar A100 entre 100 nós
  → Flooding: pergunta para todos → encontra, mas custa O(N²)
  → Kademlia: calcula direção XOR → encontra, custa O(log N)

REDE NEURAL (Transformer):
  Token[0] precisa "atender" a Token[100] entre 100 tokens
  → Full Attention: compara com todos → encontra, custa O(N²)
  → Sinapse Routing: hash CDC determina quais tokens são "vizinhos semânticos"
    → Sparse Attention guiada por hash → encontra, custa O(N log N)
```

O mecanismo de **atenção total** (full attention) do Transformer é, essencialmente, um **flooding** no espaço de tokens. Cada token "fala" com todos os outros tokens. O Sinapse propõe usar **roteamento baseado em hash** (análogo ao Kademlia) para criar atenção esparsa e eficiente.

### Codebook como Tabela de Rotas

```
DHT:       Hash(nó_destino) → próximo vizinho na rota
Sinapse:   Hash(chunk_CDC) → ativação cacheada mais relevante
```

O Codebook de ativações funciona como uma **tabela de roteamento**: dado o hash de um chunk, ele indica diretamente onde está a informação relevante, sem precisar "perguntar" a todas as camadas da rede.

---

## Implicações Arquiteturais

Se o Sinapse for distribuído (ex: múltiplos nós rodando inferência local em rede mesh), os protocolos de roteamento de grafos tornam-se diretamente aplicáveis:

```
Nó A: tem Codebook especializado em Go
Nó B: tem Codebook especializado em Segurança
Nó C: tem Codebook generalista

Query: "Auditoria de segurança em Go"
  → GossipSub propaga a query
  → Kademlia roteia para Nó A + Nó B (XOR distance)
  → Nós A e B colaboram na resposta
  → O custo de roteamento é O(log N), não O(N)
```

---

> **Próximo:** [09 — Roadmap](09-ROADMAP.md)
