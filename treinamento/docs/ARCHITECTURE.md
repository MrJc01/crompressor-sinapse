# Arquitetura Crompressor-Sinapse: Treinamento O(1)

Bem-vindo ao portal lógico do modelo cognitivo construído puramente em Go.

## 1. Visão Geral (Vector Quantization Euclidiana)
Esta seção do Monorepo hospeda o pipeline de Treinamento de LLM Nativo. Aqui banimos completamente abordagens como Numpy e PyTorch. Nosso treinamento é projetado para rodar em DHT P2P Escalar.

Para evitar os gargalos O(N²) de uma rede Transformer Clássica, transformamos a rede num roteador Hash.

Em termos práticos:
- Não possuímos `float32` densos iteráveis no caminho quente real de teste.
- O treinamento transforma o gradiente clássico em Distância Euclidiana e a salva como `XORDeltas`.

## 2. A Camada de Tensão (TensorO1)
O pacote `core` detém o `TensorO1`. Essa struct armazena chaves codificadas e é inteiramente voltada para processamento `Zero-Allocation`. Isso significa que nenhuma fatia (`slice`) é constantemente alocada e destruída na Heap do Go, aliviando o Garbage Collector.
