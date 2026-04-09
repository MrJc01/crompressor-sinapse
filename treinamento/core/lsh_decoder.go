package core

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

// CausalGraph encapsula o mapa O(1) das Transições Extraídas da Nuvem
type CausalGraph struct {
	EdgesCount  uint64
	NGramSize   uint64
	StateMap    map[[3]uint32]uint32 // NATIVE GO HASHING O(1) (sem FNV manual)
	ValidStates [][3]uint32          // Vetor de resgate pra Trigram Smoothing
}

// LoadCromGraph substitui o Decoder LSH Randômico pelo Decoder de Estados N-Gram O(1)
func LoadCromGraph(filepath string) (*CausalGraph, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("impossível acoplar grafo: %v", err)
	}
	defer file.Close()

	graph := &CausalGraph{
		StateMap: make(map[[3]uint32]uint32),
	}

	err = binary.Read(file, binary.LittleEndian, &graph.EdgesCount)
	if err != nil { return nil, err }

	err = binary.Read(file, binary.LittleEndian, &graph.NGramSize)
	if err != nil { return nil, err }

	// Lê as conexões de inteligência estritas.
	for i := uint64(0); i < graph.EdgesCount; i++ {
		var edge [4]uint32
		err = binary.Read(file, binary.LittleEndian, &edge)
		if err != nil { return nil, err }
		
		state := [3]uint32{edge[0], edge[1], edge[2]}
		graph.StateMap[state] = edge[3]
		graph.ValidStates = append(graph.ValidStates, state)
	}

	rand.Seed(time.Now().UnixNano())
	return graph, nil
}

// NextToken LSH Smoothing. Se cair fora do grafo exato, ele se recupera achando vizinho no Grafo (Cache Miss handler).
func (g *CausalGraph) NextToken(a, b, c uint32) (uint32, [3]uint32) {
	state := [3]uint32{a, b, c}
	target, ok := g.StateMap[state]
	if !ok {
		// FALLBACK CAUSAL O(1): Resgate Aleatorio Controlado
		// "Pensamento Livre" (Quando n sabe o grama exato, puxa tracao de memoria proxima)
		recup := g.ValidStates[rand.Intn(len(g.ValidStates))]
		return g.StateMap[recup], recup
	}
	return target, state
}

// LoadVocab importa as Strings do modelo da nuvem com limpeza BPE
func LoadVocab(filepath string) (map[uint32]string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("vocab.json not found: %v", err)
	}
	defer file.Close()

	var strMap map[string]string
	if err := json.NewDecoder(file).Decode(&strMap); err != nil {
		return nil, err
	}

	vocab := make(map[uint32]string)
	for k, v := range strMap {
		var id uint32
		fmt.Sscanf(k, "%d", &id)
		
		// Remove artefatos de Tokenizador BPE padrão de Nuvem (ex: Ġ, <unk>)
		clean := strings.ReplaceAll(v, "Ġ", " ")
		clean = strings.ReplaceAll(clean, "Ċ", "\n")
		vocab[id] = clean
	}
	return vocab, nil
}
