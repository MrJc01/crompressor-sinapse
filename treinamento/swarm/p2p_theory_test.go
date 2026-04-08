package swarm

import (
	"testing"
	"time"

	"github.com/MrJc01/crompressor-sinapse/treinamento/core"
)

// TestGossipConvergenceAndStudy prova analiticamente o ganho colossal em Banda
// e Velocidade. Simula 3 nós descentralizados, avaliando o comportamento emergente 
// do LLM assim que um Delta de 8 Bytes flui na malha P2P.
func TestGossipConvergenceAndStudy(t *testing.T) {
	shape := []int{2048}

	t1 := core.NewTensorO1(shape)
	t2 := core.NewTensorO1(shape)
	t3 := core.NewTensorO1(shape)

	// Instanciando o Swarm em sub-processos dinêmicos
	alpha := NewSwarmNode("Alpha", t1)
	addrA, _ := alpha.StartGossipListenerUDP()
	defer alpha.Close()
	// Supress warnings (Alpha acts mainly as emitter here)
	_ = addrA 

	beta := NewSwarmNode("Beta", t2)
	addrB, _ := beta.StartGossipListenerUDP()
	defer beta.Close()

	gamma := NewSwarmNode("Gamma", t3)
	addrC, _ := gamma.StartGossipListenerUDP()
	defer gamma.Close()

	// 1. Alfa encontra uma "Surpresa" na inferência local e faz um Backprop Euclidiano O(1)
	inputIDs := []uint32{1024}
	alphaOutput := []uint32{0x00}
	alphaTarget := []uint32{0xFF} // Precisa aprender 0xFF

	startBackprop := time.Now()
	alpha.Tensor.ApplyBackprop(inputIDs, alphaOutput, alphaTarget)
	backpropLatency := time.Since(startBackprop)

	learntMask := alpha.Tensor.XORDeltas[1024]

	// 2. Transmissão em Efeito Cascata (Swarm Mesh Simulate)
	startTx := time.Now()
	_ = alpha.BroadcastDelta(addrB, 1024, learntMask)
	_ = alpha.BroadcastDelta(addrC, 1024, learntMask)
	txLatency := time.Since(startTx)

	// Aguarda estabilização do hardware de rede virtual (Loopback OS Overhead)
	time.Sleep(15 * time.Millisecond)

	// 3. O Estudo da Mecânica Forense e Integridade do Modelo Remoto
	core.Logger.Info("Metrics Laboratorio P2P", "Backpropagation_CPU_Time", backpropLatency.String(), "Broadcast_Tx_Time", txLatency.String())
	
	gamma.mu.Lock()
	p := gamma.RXPackets
	bytesRead := gamma.RXBytes
	gamma.mu.Unlock()

	core.Logger.Info("Auditoria Network Swarm - Nó Gamma", "RX_Packets", p, "RX_Bytes_Consumed", bytesRead)

	// Affirmações estritas da Ciência Aplicada
	if bytesRead != 8 {
		t.Fatalf("[Falha Critica] Overhead insustentável P2P esperado: 8 bytes. Obtido: %d bytes", bytesRead)
	}

	if beta.Tensor.XORDeltas[1024] != 0xFF {
		t.Errorf("Beta não atingiu convergência neural local. Estado atual: %X", beta.Tensor.XORDeltas[1024])
	}

	if gamma.Tensor.XORDeltas[1024] != 0xFF {
		t.Errorf("Gamma não atingiu convergência neural local. Estado atual: %X", gamma.Tensor.XORDeltas[1024])
	}
}
