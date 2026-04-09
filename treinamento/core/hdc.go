package core

import (
	"crypto/sha256"
	"encoding/binary"
	"math/bits"
	"os"
	"strings"
)

// HyperVector representa as 256 dimensões compactadas
type HyperVector [4]uint64

type HDCNode struct {
	Signature HyperVector
	Logic     string
}

type HDCGraph struct {
	Nodes []HDCNode
}

// ComputeSimHash com Força Sintática (Tri-Grams). Recalcula exatamente o sentido.
func ComputeSimHash(text string) HyperVector {
	text = strings.ToLower(text)
	words := strings.Fields(text)

	var v [256]int

	for i := 0; i < len(words); i++ {
		// Base (Unigram)
		addHashToVector(&v, sha256.Sum256([]byte(words[i])))
		
		// Ligação Curta (Bigrama) - Peso 2x para garantir Syntax
		if i < len(words)-1 {
			h := sha256.Sum256([]byte(words[i] + " " + words[i+1]))
			addHashToVector(&v, h)
			addHashToVector(&v, h)
		}
		
		// Sentido Geral e Concordância (Trigrama) - Peso 3x
		if i < len(words)-2 {
			h := sha256.Sum256([]byte(words[i] + " " + words[i+1] + " " + words[i+2]))
			for k := 0; k < 3; k++ { addHashToVector(&v, h) }
		}
	}

	var sig HyperVector
	for i := 0; i < 256; i++ {
		if v[i] > 0 {
			chunk := i / 64
			bitPos := uint(i % 64)
			sig[chunk] |= (1 << (63 - bitPos))
		}
	}
	return sig
}

func addHashToVector(v *[256]int, h [32]byte) {
	for byteIndex := 0; byteIndex < 32; byteIndex++ {
		b := h[byteIndex]
		for bitIndex := 0; bitIndex < 8; bitIndex++ {
			pos := byteIndex*8 + bitIndex
			if (b & (1 << (7 - bitIndex))) != 0 {
				v[pos]++
			} else {
				v[pos]--
			}
		}
	}
}

func HammingDistance(a, b HyperVector) int {
	return bits.OnesCount64(a[0]^b[0]) +
		bits.OnesCount64(a[1]^b[1]) +
		bits.OnesCount64(a[2]^b[2]) +
		bits.OnesCount64(a[3]^b[3])
}

// LoadHDCGraph Carrega o arquivo com Transições Palavra a Palavra.
func LoadHDCGraph(filepath string) (*HDCGraph, error) {
	file, err := os.Open(filepath)
	if err != nil { return nil, err }
	defer file.Close()

	var total uint64
	if err := binary.Read(file, binary.LittleEndian, &total); err != nil { return nil, err }

	g := &HDCGraph{
		Nodes: make([]HDCNode, 0, total),
	}
	
	for i := 0; uint64(i) < total; i++ {
		var sig HyperVector
		binary.Read(file, binary.LittleEndian, &sig)

		var l uint32
		binary.Read(file, binary.LittleEndian, &l)
		buffer := make([]byte, l)
		file.Read(buffer)

		g.Nodes = append(g.Nodes, HDCNode{
			Signature: sig,
			Logic:     string(buffer),
		})
	}
	return g, nil
}

// SearchAutoregressive acha no Dicionário O(1) qual Palavra segue o atual Contexto da Frase. (K-NN K=1 Distância Greedy)
func (g *HDCGraph) SearchAutoregressive(target HyperVector) string {
	bestDist := 999999
	bestStr := ""
	
	// Motor Causal Linear
	for _, n := range g.Nodes {
		d := HammingDistance(target, n.Signature)
		if d < bestDist {
			bestDist = d
			bestStr = n.Logic
		}
		// Saída precoce Suprema Hashing (0 colisão total)
		if d == 0 { break } 
	}
	return bestStr
}
