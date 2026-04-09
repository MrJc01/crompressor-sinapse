package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/MrJc01/crompressor-sinapse/treinamento/core"
)

type Record struct {
	Instruction string `json:"instruction"`
	Input       string `json:"input"`
	Output      string `json:"output"`
}

func buildVocab(dataset []Record, maxSize int) (map[string]int, []string) {
	counts := make(map[string]int)
	for _, r := range dataset {
		words := strings.Fields(strings.ToLower(r.Instruction + " " + r.Input + " " + r.Output))
		for _, w := range words {
			clean := strings.Trim(w, "?!.,;()[]\"'")
			if len(clean) > 0 { counts[clean]++ }
		}
	}
	
	type kv struct { K string; V int }
	var arr []kv
	for k, v := range counts { arr = append(arr, kv{k, v}) }
	sort.Slice(arr, func(i, j int) bool { return arr[i].V > arr[j].V })

	vocab := make(map[string]int)
	var reverse []string
	
	// Palavras Reservadas no ID 0
	vocab["<UNK>"] = 0
	reverse = append(reverse, "<UNK>")

	for i := 0; i < maxSize-1 && i < len(arr); i++ {
		vocab[arr[i].K] = i + 1
		reverse = append(reverse, arr[i].K)
	}

	// Complete until max
	for len(reverse) < maxSize {
		reverse = append(reverse, "<PAD>")
	}
	return vocab, reverse
}

func getWordIDs(text string, vocab map[string]int) []int {
	words := strings.Fields(strings.ToLower(text))
	var ids []int
	for _, w := range words {
		clean := strings.Trim(w, "?!.,;()[]\"'")
		if id, exists := vocab[clean]; exists {
			ids = append(ids, id)
		} else {
			ids = append(ids, 0) // UNK
		}
	}
	return ids
}

func main() {
	fmt.Println("\033[1;36m===================================================\033[0m")
	fmt.Println("\033[1;36m| CROM-LLM: TREINAMENTO MOE (BACKPROPAGATION NATIVO) |\033[0m")
	fmt.Println("\033[1;36m===================================================\033[0m")

	file, _ := os.Open("cabrita.json")
	var fullDataset []Record
	json.NewDecoder(file).Decode(&fullDataset)
	file.Close()

	var dataset []Record
	// Filtro de Qualidade: Apenas resumos e coisas curtas e puras em portugues
	for _, rec := range fullDataset {
		if strings.Contains(rec.Output, "gravidade") || strings.Contains(rec.Output, "O ") || strings.Contains(rec.Output, "A ") {
			if len(rec.Output) > 10 && len(rec.Output) < 200 {
				dataset = append(dataset, rec)
			}
		}
		if len(dataset) >= 500 { break } // Overfit Extremo de 500 records pra forçar a inteligência em 15 seg
	}

	fmt.Println("[*] Vocabulário e Mapeamento Tensorial...")
	vocabSize := 2048
	numExperts := 32 // LSH Routing Experts
	dim := 64

	vocab, revVocab := buildVocab(dataset, vocabSize)
	model := core.NewCROMModel(vocabSize, dim, numExperts)

	fmt.Printf("[*] Iniando Deep Learning NATIVO! Parâmetros: %d\n", len(model.Emb) + len(model.Experts)*vocabSize*dim)

	// LSH Hash Routing Função. Onde os 256 Bits decidem qual Rede Neural Mini atua.
	routeExpert := func(ctxIds []int) int {
		sig := core.ComputeSimHash(fmt.Sprintf("%v", ctxIds))
		return int(sig[0] % uint64(numExperts))
	}

	epochs := 40
	for epoch := 1; epoch <= epochs; epoch++ {
		var avgLoss float32
		var steps int
		rand.Seed(time.Now().UnixNano())

		for i, rec := range dataset {
			tokens := getWordIDs(rec.Instruction+" "+rec.Input+" <think> "+rec.Output, vocab)
			
			// Auto Regressive Context Window (4 tokens)
			for j := 4; j < len(tokens); j++ {
				ctx := tokens[j-4 : j]
				target := tokens[j]

				expID := routeExpert(ctx)
				loss := model.TrainStep(ctx, target, expID, 0.05)
				
				avgLoss += loss
				steps++
			}
			if i%500 == 0 && i > 0 {
				fmt.Printf("Epoch %d: [%d/%d] Loss Distilada: %.4f\n", epoch, i, len(dataset), avgLoss/float32(steps))
				avgLoss = 0; steps = 0
			}
		}
	}

	fmt.Println("[*] Quantizando os Neurônios (Exportando para Int8/BitNet)...")
	qModel := model.QuantizeModel()

	// Salvando os Modelos no Formato Proprietário CROM
	f, _ := os.Create("brain.crom")
	defer f.Close()

	// Grava Vocabulário
	binary.Write(f, binary.LittleEndian, uint32(vocabSize))
	for _, w := range revVocab {
		length := uint16(len(w))
		binary.Write(f, binary.LittleEndian, length)
		f.Write([]byte(w))
	}

	// Grava a Rede Neural Int8
	binary.Write(f, binary.LittleEndian, uint32(dim))
	binary.Write(f, binary.LittleEndian, uint32(numExperts))
	binary.Write(f, binary.LittleEndian, qModel.Scale)
	binary.Write(f, binary.LittleEndian, qModel.Emb)

	for _, exp := range qModel.Experts {
		binary.Write(f, binary.LittleEndian, exp.Weights)
		binary.Write(f, binary.LittleEndian, exp.Bias)
	}

	fmt.Println("[!] Destilação LSH-MoE GO completa! O Modelo é independente de LLMs PyTorch agora.")
}
