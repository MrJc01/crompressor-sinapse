package swarm

import (
	"bytes"
	"encoding/binary"
	"net"
	"sync"

	"github.com/MrJc01/crompressor-sinapse/treinamento/core"
)

// UpdatePayload é a estrutura mínima de Backprop que flui na malha O(1).
// Em um float32 normal (NVIDIA/PyTorch), transmitiríamos matrizes gigantescas (Centenas de Megabytes a Gigabytes).
// Aqui na topologia CDC Euclidiana transmitimos rigorosos 8 Bytes da Mutação (Index e Máscara XOR).
type UpdatePayload struct {
	Index uint32
	Mask  uint32
}

// SwarmNode detém o ciclo de vida do "Aprendizado Assíncrono UDP".
type SwarmNode struct {
	ID        string
	Tensor    *core.TensorO1
	conn      *net.UDPConn
	Wg        sync.WaitGroup
	mu        sync.Mutex
	RXPackets int
	RXBytes   int
}

// NewSwarmNode anexa o motor matemático a uma interface de rede.
func NewSwarmNode(id string, t *core.TensorO1) *SwarmNode {
	return &SwarmNode{
		ID:     id,
		Tensor: t,
	}
}

// StartGossipListenerUDP sobe um socket dinâmico não bloqueante alinhado à goroutine.
func (n *SwarmNode) StartGossipListenerUDP() (*net.UDPAddr, error) {
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}
	n.conn = conn

	n.Wg.Add(1)
	go func() {
		defer n.Wg.Done()
		buf := make([]byte, 8) // Janela ultra fixada para Zero Overhead de GC
		for {
			bytesRead, _, err := n.conn.ReadFromUDP(buf)
			if err != nil {
				return // Socket Fechado
			}
			if bytesRead == 8 {
				n.mu.Lock()
				n.RXPackets++
				n.RXBytes += bytesRead
				n.mu.Unlock()

				var payload UpdatePayload
				reader := bytes.NewReader(buf)
				_ = binary.Read(reader, binary.LittleEndian, &payload)

				core.LogForensic("Camada Rede/Aplicação (Swarm)", "Delta Euclidiano recebido via Gossip P2P. Atualizando O(1).", "IdLocal", n.ID, "TargetIndex", payload.Index)
				
				// Convergência Direta de Estado na Memória Base (Desvio total do Backprop de CPU)
				n.Tensor.XORDeltas[payload.Index] ^= payload.Mask
			}
		}
	}()
	return n.conn.LocalAddr().(*net.UDPAddr), nil
}

// Close encerra a malha.
func (n *SwarmNode) Close() {
	if n.conn != nil {
		n.conn.Close()
	}
}

// BroadcastDelta atua como o emissor Gossip, desabafando o aprendizado local para a rede distribuída.
func (n *SwarmNode) BroadcastDelta(targetAddr *net.UDPAddr, index, mask uint32) error {
	conn, err := net.DialUDP("udp", nil, targetAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	payload := UpdatePayload{Index: index, Mask: mask}
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, payload)

	_, err = conn.Write(buf.Bytes())
	return err
}
