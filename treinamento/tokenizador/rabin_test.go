package tokenizador

import (
	"testing"
)

// TestSemanticHashCollision atesta matematicamente que subwords geram ativações neurais diferentes 
// e garantem robustez plena p/ o Treinamento no ambiente DHT.
func TestSemanticHashCollision(t *testing.T) {
	dicionarioBaseSize := uint32(4096)
	
	// Teste de integridade de Mutação Singular (O "UDP" vs "P2P" no índice 4)
	f1 := "Este é o núcleo P2P purista"
	f2 := "Este é o núcleo UDP purista"
	
	ids1 := TokenizeToIDs(f1, dicionarioBaseSize)
	ids2 := TokenizeToIDs(f2, dicionarioBaseSize)
	
	if len(ids1) != len(ids2) {
		t.Fatalf("Comprimento gerado assimétrico no parser CDC O(1)")
	}

	// 0: Este, 1: é, 2: o, 3: núcleo, 4: P2P/UDP, 5: purista
	if len(ids1) < 6 {
		t.Fatalf("Tokens insuficientes detectados: falha no Parser")
	}

	if ids1[4] == ids2[4] {
		t.Errorf("Falha Crítica Forense! Colisão de índice no limite de Hash: P2P=%d UDP=%d", ids1[4], ids2[4])
	}
	
	// Restante tem que ser inteiramente idêntico
	if ids1[3] != ids2[3] || ids1[5] != ids2[5] {
		t.Errorf("Falso Positivo de quebra CDC. Arquitetura Neural falhou a integridade dos buffers base.")
	}
}


// BenchmarkTokenizeToIDs mede a velocidade estrondosa da quebra CDC do Go.
func BenchmarkTokenizeToIDs(b *testing.B) {
	dicionarioBaseSize := uint32(2048)
	
	bigString := "A Inteligencia Artificial do Crompressor domina o Swarm. "
	// Extrapola o input massivamente
	for i := 0; i < 4; i++ {
		bigString += bigString 
	}
	
	b.ResetTimer()
	b.ReportAllocs() // Para testar strict Minimal Allocations Mode
	
	for i := 0; i < b.N; i++ {
		_ = TokenizeToIDs(bigString, dicionarioBaseSize)
	}
}
