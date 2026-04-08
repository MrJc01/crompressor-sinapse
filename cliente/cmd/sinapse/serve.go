package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/MrJc01/crompressor-sinapse/internal/daemon"
)

func serveCmd() *cobra.Command {
	var port string
	var llamaURL string
	var nodeID string
	var peersRaw string

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Inicializa o Daemon MMap Crompressor acoplado em background a um servidor Open-Source API",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("🌀 Inicializando Crompressor-Sinapse Daemon Core (Node: %s)...\n\n", nodeID)
			
			var peers []string
			if peersRaw != "" {
				peers = strings.Split(peersRaw, ",")
			}

			server := daemon.NewDaemonServer(port, llamaURL, nodeID, peers)

			if err := server.Start(); err != nil {
				return fmt.Errorf("daemon server crasheou: %v", err)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&port, "port", "p", "8888", "Porta local para escutar requisições Front-End")
	cmd.Flags().StringVarP(&llamaURL, "llama-url", "l", "http://127.0.0.1:8080", "Endpoint do Llama.cpp backend real")
	cmd.Flags().StringVarP(&nodeID, "node-id", "n", "node-master", "ID único para roteamento Gossip DHT P2P")
	cmd.Flags().StringVar(&peersRaw, "peers", "", "Lista de endereços http dos peers vizinhos separados por virgula")

	return cmd
}
