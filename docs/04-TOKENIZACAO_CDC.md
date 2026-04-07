# 04 — Tokenização Semântica via CDC

> *"O BPE corta palavras. O CDC corta conceitos."*

---

## O Fim do Byte-Pair Encoding

O BPE (Byte-Pair Encoding) é o tokenizador padrão de LLMs. Ele constrói um vocabulário estatístico mesclando pares de bytes mais frequentes, iterativamente, até atingir um tamanho de vocabulário alvo (~32K–128K tokens).

O problema é que o BPE é **cego à estrutura**. Ele não sabe que `func main()` é uma unidade lógica. Para o BPE, é uma sequência estatística como qualquer outra.

---

## A Proposta: Rolling Hash para Fronteiras Naturais

O Crompressor usa **Content-Defined Chunking (CDC)** com rolling hash (Rabin, Buzhash) para identificar fronteiras naturais no fluxo de dados binários. A mesma técnica pode ser aplicada a **texto natural e código-fonte**.

### Mecânica

```
Texto de entrada:
    "O Crompressor usa CDC para encontrar fronteiras naturais no texto."

BPE (fixo, estatístico):
    ["O", " Cromp", "ressor", " usa", " CDC", " para", " encontrar",
     " front", "eiras", " naturais", " no", " texto", "."]
    → 13 tokens. "Crompressor" foi quebrado arbitrariamente.

CDC (dinâmico, estrutural):
    [CHUNK: "O Crompressor", CHUNK: "usa CDC", 
     CHUNK: "para encontrar fronteiras naturais", CHUNK: "no texto."]
    → 4 chunks semânticos. Cada um é uma unidade de significado.
```

### Implementação Conceitual

```go
// CDCTokenizer aplica rolling hash para encontrar fronteiras em texto.
func CDCTokenizer(text []byte, avgSize int) []Token {
    const (
        minSize = 8    // Mínimo: evitar chunks degenerados
        maxSize = 256  // Máximo: forçar fronteira
        mask    = (1 << 11) - 1  // Avg ~2048 chars
    )
    
    var (
        tokens []Token
        start  int
        fp     uint64  // Rolling hash (Rabin fingerprint)
    )
    
    for i := 0; i < len(text); i++ {
        fp = rabinUpdate(fp, text[i])
        size := i - start
        
        // Fronteira: quando o hash "casa" com a máscara
        // OU espaço/pontuação natural OU tamanho máximo
        if (size >= minSize && fp&mask == 0) || 
           isNaturalBoundary(text[i]) || 
           size >= maxSize {
            
            chunk := text[start : i+1]
            hash := xxhash.Sum64(chunk)
            
            // Buscar no Codebook de tokens
            tokenID, found := codebook.Lookup(hash)
            if !found {
                tokenID = codebook.Insert(chunk, hash)
            }
            
            tokens = append(tokens, Token{
                ID:   tokenID,
                Data: chunk,
                Hash: hash,
            })
            start = i + 1
        }
    }
    
    return tokens
}
```

---

## Vantagens do CDC sobre BPE

| Aspecto | BPE | CDC |
|---------|-----|-----|
| **Fronteiras** | Estatísticas (frequência de pares) | Estruturais (hash + semântica) |
| **Tamanho dos tokens** | Fixo/Semi-fixo (~3-4 chars médios) | Variável (8–256 chars) |
| **Sensibilidade a inserção** | 1 byte inserido → todos os tokens mudam | 1 byte inserido → 1-2 chunks mudam |
| **Vocabulário** | 32K–128K entradas fixas | Dinâmico, cresce com o corpus |
| **Semântica** | Nenhuma (fragmentos de palavras) | Alta (unidades lógicas) |
| **Cache-friendly** | Não (fronteiras instáveis) | Sim (fronteiras determinísticas) |

---

## Impacto na Rede Neural

### Redução de Sequência

Se o CDC gera **3-10x menos tokens** do que o BPE para o mesmo texto, o contexto efetivo da rede se expande proporcionalmente:

```
Contexto do modelo: 4096 tokens

Com BPE:   4096 tokens ≈ 12KB de texto
Com CDC:   4096 chunks  ≈ 40-120KB de texto  (3-10x mais contexto!)
```

### Atenção Mais Eficiente

O mecanismo de atenção do Transformer escala com O(n²) onde n = número de tokens. Com CDC gerando menos tokens:

```
BPE:  n = 4096 → n² = 16,777,216 operações de atenção
CDC:  n = 1000 → n² = 1,000,000 operações de atenção
                      → 16x menos computação no attention layer
```

---

## Codebook de Tokens

Ao contrário do vocabulário fixo do BPE, o CDC constrói um **Codebook dinâmico** de tokens:

```
┌────────────────────────────────────────────┐
│         CODEBOOK DE TOKENS CDC              │
│                                             │
│  Hash → Token:                              │
│  0xA3C2... → "func main() {"               │
│  0xF891... → "fmt.Println("                 │
│  0x12B4... → "if err != nil {"              │
│  0x9D7E... → "return nil, fmt.Errorf("      │
│  ...                                        │
│                                             │
│  Cada chunk comum do corpus → 1 ID único    │
│  A rede recebe IDs, não bytes               │
└────────────────────────────────────────────┘
```

Para código Go, o Codebook capturaria automaticamente padrões como:
- Declarações de função
- Tratamento de erro (`if err != nil`)
- Imports comuns
- Loops e condicionais idiomáticos

A rede não precisaria mais "aprender" essas estruturas — elas já estariam codificadas como unidades atômicas.

---

## Validação Proposta

### Experimento 1: Densidade de Tokens
```
Dataset: 1GB de repositórios Go (~10M linhas de código)

Medir:
  - Tokens BPE gerados (tiktoken, cl100k_base)
  - Chunks CDC gerados (Rabin hash, avg=128)
  - Ratio: BPE/CDC

Hipótese: CDC gera 3-8x menos unidades
```

### Experimento 2: Estabilidade de Cache
```
Arquivo v1: handler.go original
Arquivo v2: handler.go com 5 linhas alteradas

Medir:
  - % de tokens BPE que mudam entre v1 e v2
  - % de chunks CDC que mudam entre v1 e v2

Hipótese: CDC tem >90% de estabilidade; BPE <60%
```

---

> **Próximo:** [05 — Forward Pass Diferencial](05-FORWARD_PASS_DIFERENCIAL.md)
