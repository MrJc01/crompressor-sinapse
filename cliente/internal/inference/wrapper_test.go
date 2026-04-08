package inference

import (
	"strings"
	"testing"

	"github.com/MrJc01/crompressor-sinapse/internal/cdc"
)

func TestWrapper_ProcessPrompt_DifferentialLatency(t *testing.T) {
	cache := NewActivationCache(500)
	wrapper := NewSimulatedWrapper(cache, 128)
	cdcOpts := cdc.DefaultOptions() // min: 32, max: 1024, avg: 128

	// Construindo um prompt grande (Muitos CDC Chunks)
	longPrompt := strings.Repeat("I am a simulated LLM prompt generating artificial semantic chunks repeatedly. ", 50)
	
	// Primeira passagem: TUDO DEVE SER MISS. 
	// Computamos todos os chunks sob o custo InjectedLatencyMs (15ms * X)
	// (Nota: o mock usou string repetition no prompt, o CDC nativamente acusará 
	// alguns bypasses intrinsecos pro próprio Turno 1!).
	outA := wrapper.ProcessPrompt(longPrompt, cdcOpts)
	
	if outA.Computed == 0 {
		t.Errorf("Nenhum chunk foi computado, isso invalida o LLM Mock")
	}

	t.Logf("Time To First Token [Turno 1]: %v | Computed: %d", outA.TimeToken, outA.Computed)

	// Segunda passagem diferencial: O prompt agora engloba uma inserção pequena no final, de resto a matriz CDC é a mesma.
	deltaPrompt := longPrompt + " Aqui está a diferença pequena e isolada de bytes adicionada."
	
	outB := wrapper.ProcessPrompt(deltaPrompt, cdcOpts)
	
	// Criterio de Sucesso: As porções pré processadas (hits do HASH CDC) devem escoar do Cache pro Layer Final.
	if outB.Bypassed == 0 {
		t.Errorf("Na 2ª passagem diferencial falhou no ByPass: Tivemos 0 Bypasses e o Cache não performou.")
	}
	
	if outB.Computed > 5 {
		t.Errorf("A pequena diferença acionou recalculo excessivo de rede Neural (Amnésia).")
	}

	t.Logf("Time To First Token [Turno 2 - CDC Cached]: %v | Bypassed: %d | Computed: %d", outB.TimeToken, outB.Bypassed, outB.Computed)

	// O Teste fundamental da tese Crompressor-Sinapse
	// (Limitado a analisar Bypass em vez de cronometrar CPU Time)
	if outB.Bypassed == 0 {
		t.Errorf("Turno Dinâmico não bypassou nada.")
	}
}
