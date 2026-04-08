package p2p

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/MrJc01/crompressor-sinapse/internal/inference"
)

// GossipMessage carrega vetores O(1) de atalhos e difusões P2P (Bypasses Neurais)
type GossipMessage struct {
	SenderID string      `json:"sender_id"`
	Hash     uint64      `json:"hash"`
	Vector   []float32   `json:"vector"`
}

// Node manipula a conectividade horizontal
type Node struct {
	ID         string
	Peers      []string // Lista de IPs/Endereços. Ex: "http://127.0.0.1:8081"
	Cache      *inference.ActivationCache
	client     *http.Client
	peerLock   sync.RWMutex
}

// NewNode instancia um vizinho da rede Swarm
func NewNode(id string, bootstrapPeers []string, cache *inference.ActivationCache) *Node {
	return &Node{
		ID:    id,
		Peers: bootstrapPeers,
		Cache: cache,
		client: &http.Client{
			Timeout: 2 * time.Second, // Gossips são rápidos ou ignorados se o nó caiu
		},
	}
}

// BroadcastHashedVector anuncia um bypass recém-calculado a todos os conhecidos vizinhos.
func (n *Node) BroadcastHashedVector(hash uint64, vector []float32) {
	msg := GossipMessage{
		SenderID: n.ID,
		Hash:     hash,
		Vector:   vector,
	}
	
	payload, _ := json.Marshal(msg)

	n.peerLock.RLock()
	defer n.peerLock.RUnlock()

	for _, peerAddr := range n.Peers {
		go func(addr string) {
			targetURL := addr + "/p2p/gossip"
			req, err := http.NewRequest("POST", targetURL, bytes.NewBuffer(payload))
			if err == nil {
				req.Header.Set("Content-Type", "application/json")
				n.client.Do(req)
				// Na política Gossip, falhas (Node offline) são silenciosamente descartadas (No-Op)
			}
		}(peerAddr)
	}
}

// HandleGossip processa injeções de vizinhos e armazena bypass semântico remotamente calculado.
func (n *Node) HandleGossip(w http.ResponseWriter, r *http.Request) {
	var msg GossipMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Insere ou ignora de forma safe dentro do core ActivationCache
	// Se já existir, evitamos loop infinito de broadcasts em anel
	if _, hit := n.Cache.Get(msg.Hash); !hit {
		n.Cache.Put(msg.Hash, msg.Vector)
		log.Printf("[P2P Mesh] BINGO! Recebeu vetor latente O(1) resolvido pelo nó: %s", msg.SenderID)
		
		// Propagação Epidêmica (Encaminha pros amiguinhos dele se for a 1º vez)
		n.BroadcastHashedVector(msg.Hash, msg.Vector)
	}

	w.WriteHeader(http.StatusOK)
}
