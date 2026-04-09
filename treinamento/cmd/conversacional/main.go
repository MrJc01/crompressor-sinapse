package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/MrJc01/crompressor-sinapse/treinamento/core"
)

func LoadNeuralEngine(filepath string) (*core.Int8Model, []string, map[string]int, error) {
	f, err := os.Open(filepath)
	if err != nil { return nil, nil, nil, err }
	defer f.Close()

	var vocabSize, seqLen uint32
	binary.Read(f, binary.LittleEndian, &vocabSize)
	binary.Read(f, binary.LittleEndian, &seqLen)

	revVocab := make([]string, vocabSize)
	vocab := make(map[string]int)

	for i := 0; i < int(vocabSize); i++ {
		var length uint16
		binary.Read(f, binary.LittleEndian, &length)
		buf := make([]byte, length)
		f.Read(buf)
		word := string(buf)
		revVocab[i] = word
		vocab[word] = i
	}

	var dim, numExperts, numLayers uint32
	var scale float32
	binary.Read(f, binary.LittleEndian, &dim)
	binary.Read(f, binary.LittleEndian, &numExperts)
	binary.Read(f, binary.LittleEndian, &numLayers)
	binary.Read(f, binary.LittleEndian, &scale)

	q := &core.Int8Model{
		VocabSize: int(vocabSize),
		SeqLen:    int(seqLen),
		Dim:       int(dim),
		NumLayers: int(numLayers),
		Scale:     scale,
		Emb:       make([]int8, vocabSize*dim),
		Pos:       make([]int8, seqLen*dim),
		Layers:    make([]core.Int8Layer, numLayers),
		LMHeadW:   make([]int8, dim*vocabSize),
		LMHeadB:   make([]int32, vocabSize),
	}
	binary.Read(f, binary.LittleEndian, &q.Emb)
	binary.Read(f, binary.LittleEndian, &q.Pos)

	for l := 0; l < int(numLayers); l++ {
		q.Layers[l].WQ = make([]int8, dim*dim)
		q.Layers[l].WK = make([]int8, dim*dim)
		q.Layers[l].WV = make([]int8, dim*dim)
		binary.Read(f, binary.LittleEndian, &q.Layers[l].WQ)
		binary.Read(f, binary.LittleEndian, &q.Layers[l].WK)
		binary.Read(f, binary.LittleEndian, &q.Layers[l].WV)
		
		// Fase 29: Parametros Float32 de Normalização para estancar crescimento de Desvios O(1)
		q.Layers[l].NormW = make([]float32, dim)
		q.Layers[l].NormB = make([]float32, dim)
		binary.Read(f, binary.LittleEndian, &q.Layers[l].NormW)
		binary.Read(f, binary.LittleEndian, &q.Layers[l].NormB)
		
		q.Layers[l].Experts = make([]core.Int8Expert, numExperts)
		for e := 0; e < int(numExperts); e++ {
			q.Layers[l].Experts[e].Weights = make([]int8, dim*dim) // Dimensionality O(1) Preservada!
			q.Layers[l].Experts[e].Bias = make([]int32, dim)
			binary.Read(f, binary.LittleEndian, &q.Layers[l].Experts[e].Weights)
			binary.Read(f, binary.LittleEndian, &q.Layers[l].Experts[e].Bias)
		}
	}

	binary.Read(f, binary.LittleEndian, &q.LMHeadW)
	binary.Read(f, binary.LittleEndian, &q.LMHeadB)

	return q, revVocab, vocab, nil
}

func main() {
	fmt.Println("\033[1;36m===================================================\033[0m")
	fmt.Println("\033[1;36m| CROM-LLM MoE (FASE 24 - INFERÊNCIA INT8 NATIVA)  |\033[0m")
	fmt.Println("\033[1;36m| REDE NEURAL BITNET EMBUTIDA (100% CROM MATEMÁTICO)|\033[0m")
	fmt.Println("\033[1;36m===================================================\033[0m")

	fmt.Println("[*] Acoplando Modelo LSH-MoE Diretamente na RAM...")
	model, revVocab, vocab, err := LoadNeuralEngine("brain.crom")
	if err != nil {
		fmt.Printf("\033[1;31m[ERRO]\033[0m Treine o modelo localmente com 'go run ./cmd/compiler'.\n")
		return
	}

	fmt.Printf("\033[1;32m[*] Cérebro Neural Carregado! %d Experts Quantizados em %d Camadas Profundas. Scale: %.4f\033[0m\n", len(model.Layers[0].Experts), model.NumLayers, model.Scale)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\n\033[34m[Usuário]>\033[0m ")
		if !scanner.Scan() { break }
		
		raw := strings.TrimSpace(scanner.Text())
		if raw == "!exit" || raw == "exit" { break }
		if raw == "" { continue }

		// Encode (Forçando Contexto Padrao = 16)
		words := strings.Fields(strings.ToLower(raw))
		var buffer []int = make([]int, model.SeqLen) 
		for idx := range buffer { buffer[idx] = 1 } // ID 1 é o <PAD>
		
		for _, w := range words {
			clean := strings.Trim(w, "?!.,;()[]\"'")
			if id, ok := vocab[clean]; ok {
				buffer = append(buffer, id)
			} else {
				buffer = append(buffer, 0) // <UNK>
			}
		}

		// O Gatilho Cognitivo: Obriga a rede a parar de achar que estamos escrevendo o Instruction
		// Injetamos um Ponto (.) artificial antes do <think> porque o Python foi treinado 
		// sempre com pontuação no fim da frase (Alpaca Data). Se mandarmos sem ponto, a 
		// matemática não encaixa com a Memória Posicional do Pytorch.
		trigger := []string{".", "think", "raciocínio", ".", ".", ".", "think"}
		for _, tw := range trigger {
			if id, ok := vocab[tw]; ok {
				buffer = append(buffer, id)
			}
		}

		fmt.Printf("\033[1;35m[Crompressor Neuromórfico]\033[0m -> \033[1;37m")
		
		for i := 0; i < 60; i++ {
			start := len(buffer) - model.SeqLen
			if start < 0 { start = 0 }
			ctx := buffer[start:]

			var ctxSum int
			for _, v := range ctx { ctxSum += v }
			expID := ctxSum % len(model.Layers[0].Experts)

            penalty := make(map[int]bool)
            for j := len(buffer) - 15; j < len(buffer); j++ {
                if j >= 0 { penalty[buffer[j]] = true }
            }

			nextID := model.ForwardInt8(ctx, expID, penalty)

			if nextID == 0 {
				continue
			}

			fmt.Printf("%s ", revVocab[nextID])
			time.Sleep(20 * time.Millisecond)

			buffer = append(buffer, nextID)
		}
		fmt.Println("\033[0m")
	}
}
