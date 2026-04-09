package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/MrJc01/crompressor-sinapse/treinamento/core"
	"github.com/MrJc01/crompressor-sinapse/treinamento/tokenizador"
)

func main() {
	core.LogForensic("Boot", "Inicializando Motor LLM Monolítico O(1)...", "Mode", "CLI")

	file, err := os.Open("../../dataset_simples.txt")
	if err != nil {
		core.Logger.Error("Falha crítica localizando banco de dados", "err", err)
		return
	}
	defer file.Close()

	content, _ := io.ReadAll(file)
	text := string(content)

	lines := strings.Split(text, "\n")
	dictSize := uint32(16384)

	// Instanciando Motor Zero-Alloc 
	model := core.NewTensorO1([]int{int(dictSize)})
	core.LogForensic("Engine", "Tensor Auto-Regressivo acoplado", "Parametros_Cap", dictSize)

	// Auto-Regressive Learning Loop
	for epoch := 1; epoch <= 3; epoch++ {
		core.LogForensic("Train", fmt.Sprintf("==== Start Epoch %d ====", epoch), "", "")
		errosCorrigidos := 0
		
		for _, line := range lines {
			if len(strings.TrimSpace(line)) == 0 {
				continue
			}

			// 1. O Tokenizador consome a string crua e devolve a lista CDC perfeita O(1)
			tokensIDs := tokenizador.TokenizeToIDs(line, dictSize)
			if len(tokensIDs) < 2 { continue }

			// 2. Sliding Window (Avaliação Casual N -> N+1)
			for i := 0; i < len(tokensIDs)-1; i++ {
				contextoBaseLayer := tokensIDs[i]
				targetPrevistoLayer := tokensIDs[i+1]

				inMap := []uint32{contextoBaseLayer}
				outMap := []uint32{0}

				// Forward Neural
				model.ForwardDiscreteUpdate(inMap, outMap)
				prediction := outMap[0]

				// Calculo Linear Euclidiano
				distanciaEuclidiana := core.CalculateEuclideanDelta(prediction, targetPrevistoLayer)
                
				// Correção Estrutural (Learning Factor)
				if distanciaEuclidiana != 0 {
					errosCorrigidos++
					model.ApplyBackprop(inMap, outMap, []uint32{targetPrevistoLayer})
					core.LogForensic("Aprendizado", fmt.Sprintf("Mutação Neural: %X apontará para %X", contextoBaseLayer, targetPrevistoLayer), "Loss_Ajustada", distanciaEuclidiana)
				}
			}
		}
		core.LogForensic("Metrics", fmt.Sprintf("Fim da Epoch %d. Erros (Loss): %d", epoch, errosCorrigidos), "", "")
	}

	// === TESTE DE MESA PÓS-TREINO =============================
	core.LogForensic("Inferência", "Buscando capacidade generativa via contexto hash...", "", "")
	
	// Como a linha 3 reza: "o treinamento de backprop euclidiano puro finalizado"
	// Nós entramos com "puro", ele deduz semanticamente o resto: ... "finalizado"
	prompt := tokenizador.TokenizeToIDs("puro ", dictSize) 
	
	if len(prompt) == 0 { return }
	hashPrompt := prompt[0]
	
	outTeste := []uint32{0}
	model.ForwardDiscreteUpdate([]uint32{hashPrompt}, outTeste)
	
	hashAlvoCorreto := tokenizador.TokenizeToIDs("finalizado ", dictSize)[0]

	if outTeste[0] == hashAlvoCorreto {
	    core.LogForensic("Atestado Analítico", fmt.Sprintf("SUCESSO CRÍTICO O(1): Comando contextual 'puro' (%X) gerou a intuição da palavra 'finalizado' (%X)", hashPrompt, outTeste[0]), "", "")
	} else {
	    core.LogForensic("Atestado Analítico", fmt.Sprintf("FALHA: Predição incoerente. Obteve %X, Target %X", outTeste[0], hashAlvoCorreto), "", "")
	}
}
