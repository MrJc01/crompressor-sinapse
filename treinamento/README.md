# Laboratório: Treinamento Neural O(1) e Modificação de Pesos

Bem-vindo à bifurcação de pesquisa `treinamento/`.

Enquanto a pasta `cliente/` lida primariamente com a interceptação e dedupicação "Forward-Pass" de inferências e malhas DHT (Fases 1 a 6), este diretório nasce para explorar as Fases de Convergência da **Vector Quantization (VQ-NN)** e **Sparse Deltas**.

Aqui, não atuaremos como Proxy de uma IA. Nós assumiremos a missão de modificar Ponderações e Pesos Internos de modelagens (Fine-Tuning e Training Loops), empregando as teses de XOR Bitmask e Hashes para treinar modelos usando Frações Epistêmicas de Custo O(1), contornando as densas Backpropagations de dezenas de Gigabytes.

### Roadmap do Módulo Treinamento
- [ ] Definir o Framework Lógico (Python/PyTorch vs. Go Nativo vs. C/GGML).
- [ ] Construir o Vector Quantization Layer (Codebooks Dinâmicos).
- [ ] Aplicar o Sparse Mask em Tensores Reais (Extração de LoRA/Differs).
