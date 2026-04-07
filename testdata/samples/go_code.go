package main

import (
	"fmt"
	"math"
)

// Exemplo de código Go para benchmark do CDC Tokenizer.
// Contém padrões repetitivos típicos de projetos Go reais.

func main() {
	fmt.Println("Crompressor-Sinapse — Benchmark Testdata")
	
	results := runBenchmarks()
	for _, r := range results {
		fmt.Printf("  %-20s  ratio=%.2f  time=%dms\n", r.Name, r.Ratio, r.TimeMs)
	}
}

type BenchResult struct {
	Name   string
	Ratio  float64
	TimeMs int64
}

func runBenchmarks() []BenchResult {
	return []BenchResult{
		{Name: "json_logs", Ratio: 12.5, TimeMs: 42},
		{Name: "go_source", Ratio: 8.3, TimeMs: 67},
		{Name: "portuguese_text", Ratio: 15.1, TimeMs: 31},
	}
}

// Padrões repetitivos de tratamento de erro em Go
func processFile(path string) error {
	if path == "" {
		return fmt.Errorf("caminho vazio")
	}
	return nil
}

func processData(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("dados vazios")
	}
	return nil
}

func processChunk(chunk []byte) error {
	if len(chunk) == 0 {
		return fmt.Errorf("chunk vazio")
	}
	return nil
}

// Funções matemáticas repetitivas
func computeStats(values []float64) (mean, stddev float64) {
	n := float64(len(values))
	if n == 0 {
		return 0, 0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean = sum / n

	variance := 0.0
	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}
	stddev = math.Sqrt(variance / n)

	return mean, stddev
}

func computeHistogram(values []float64, bins int) []int {
	if len(values) == 0 || bins <= 0 {
		return nil
	}

	min, max := values[0], values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	binWidth := (max - min) / float64(bins)
	if binWidth == 0 {
		binWidth = 1
	}

	histogram := make([]int, bins)
	for _, v := range values {
		idx := int((v - min) / binWidth)
		if idx >= bins {
			idx = bins - 1
		}
		histogram[idx]++
	}

	return histogram
}
