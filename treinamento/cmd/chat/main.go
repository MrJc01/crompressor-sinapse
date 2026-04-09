package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/MrJc01/crompressor-sinapse/treinamento/core"
	"github.com/MrJc01/crompressor-sinapse/treinamento/tokenizador"
)

func main() {
	fmt.Println("\033[36m[Crompressor-IA] Sequenciando Boot do Nano-Brain O(1)...\033[0m")

	file, err := os.Open("dataset_simples.txt")
	if err != nil {
		fmt.Printf("\033[31m[CRITICAL] Falha letal na leitura do Dataset: %v\033[0m\n", err)
		return
	}
	defer file.Close()

	content, _ := io.ReadAll(file)
	text := string(content)
	lines := strings.Split(text, "\n")
	
	dictSize := uint32(16384)
	model := core.NewTensorO1([]int{int(dictSize)})
    
    // Treinamento Super-Sônico de 3 Epochs (Nano-segundos em CPU)
	for epoch := 1; epoch <= 3; epoch++ {
		for _, line := range lines {
			if len(strings.TrimSpace(line)) == 0 { continue }
			tokensIDs := tokenizador.TokenizeToIDs(line, dictSize)
			if len(tokensIDs) < 2 { continue }
			
			for i := 0; i < len(tokensIDs)-1; i++ {
				ctx, tgt := tokensIDs[i], tokensIDs[i+1]
				inMap := []uint32{ctx}
				outMap := []uint32{0}
				model.ForwardDiscreteUpdate(inMap, outMap)
				if core.CalculateEuclideanDelta(outMap[0], tgt) != 0 {
					model.ApplyBackprop(inMap, outMap, []uint32{tgt})
				}
			}
		}
	}

    // Criando um Mapeamento Reverso de Hardware p/ o Humano UI ler
    reverseDictionary := make(map[uint32]string)
    for _, line := range lines {
        words := strings.Fields(line)
        for _, w := range words {
            wClean := strings.TrimSpace(w)
            tids := tokenizador.TokenizeToIDs(wClean, dictSize)
            if len(tids) > 0 {
                reverseDictionary[tids[0]] = wClean
            }
        }
    }

	fmt.Println("\033[32m[Crompressor-IA] Aprendizado Concluído T=0s. Rede Euclidiana carregada na RAM.\033[0m")
	fmt.Println("\033[33m(Digite palavras do Dataset Ex: 'swarm', 'nucleo', 'backprop' ... para ver ele adivinhar o target ou '!exit' para pular)\033[0m")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\n\033[34m[O(1) User]>\033[0m ")
		if !scanner.Scan() { break }
		
		input := strings.TrimSpace(scanner.Text())
		if input == "!exit" || input == "exit" { break }
		if input == "" { continue }

		inputIDs := tokenizador.TokenizeToIDs(input, dictSize)
		if len(inputIDs) == 0 { continue }

		// Usa apenas o Hash da ultima palavra escrita (Context Limit = 1) no modelo O(1) de Laboratorio
		lastContextHash := inputIDs[len(inputIDs)-1]
		outMap := []uint32{0}
		
		model.ForwardDiscreteUpdate([]uint32{lastContextHash}, outMap)

		predictionHash := outMap[0]
		word, exists := reverseDictionary[predictionHash]
		
		if exists {
		    fmt.Printf("\033[1;35m[O(1) Model]\033[0m -> \033[1;37m%s\033[0m \033[90m(Hashed Target: %X)\033[0m\n", word, predictionHash)
		} else {
		    fmt.Printf("\033[1;35m[O(1) Model]\033[0m -> \033[1;31m[Out-of-Vocabulary ID ou Ruído Zero: %X]\033[0m\n", predictionHash)
		}
	}
}
