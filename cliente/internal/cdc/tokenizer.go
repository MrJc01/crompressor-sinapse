package cdc

import (
	"os"
	"path/filepath"
)

// ── Tipos ──

// Chunk representa um fragmento de dados com fronteiras definidas pelo conteúdo.
type Chunk struct {
	Data   []byte // Conteúdo bruto do chunk
	Offset uint64 // Posição no arquivo original
	Size   uint32 // Tamanho em bytes
	Hash   uint64 // Hash do conteúdo (FNV-1a)
}

// Options configura o comportamento do tokenizador CDC.
type Options struct {
	MinSize int // Tamanho mínimo de chunk (default: 32)
	MaxSize int // Tamanho máximo de chunk (default: 1024)
	AvgSize int // Tamanho médio desejado (default: 128)
}

// DefaultOptions retorna opções padrão otimizadas para texto/código.
func DefaultOptions() Options {
	return Options{
		MinSize: 32,
		MaxSize: 1024,
		AvgSize: 128,
	}
}

// ── Tokenizador CDC ──

// Tokenize fragmenta dados em chunks usando Content-Defined Chunking (Rabin).
// As fronteiras são detectadas pelo rolling hash, garantindo que inserções/remoções
// afetem apenas 1-2 chunks adjacentes (resiliência a edições).
func Tokenize(data []byte, opts Options) []Chunk {
	if len(data) == 0 {
		return nil
	}

	if opts.MinSize <= 0 {
		opts.MinSize = 32
	}
	if opts.MaxSize <= 0 {
		opts.MaxSize = 1024
	}
	if opts.AvgSize <= 0 {
		opts.AvgSize = 128
	}

	mask := MaskForAvgSize(opts.AvgSize)
	hasher := NewRabinHasher()

	var chunks []Chunk
	start := 0

	for i := 0; i < len(data); i++ {
		h := hasher.Roll(data[i])
		size := i - start + 1

		// Condições de fronteira:
		// 1. Tamanho mínimo atingido E hash casa com máscara (fronteira natural)
		// 2. Tamanho máximo atingido (forçar fronteira)
		isBoundary := size >= opts.MinSize && IsBoundary(h, mask)
		isMaxSize := size >= opts.MaxSize

		if isBoundary || isMaxSize {
			chunkData := data[start : i+1]
			chunks = append(chunks, Chunk{
				Data:   chunkData,
				Offset: uint64(start),
				Size:   uint32(len(chunkData)),
				Hash:   HashBytes(chunkData),
			})
			start = i + 1
			hasher.Reset()
		}
	}

	// Último chunk (dados restantes)
	if start < len(data) {
		chunkData := data[start:]
		chunks = append(chunks, Chunk{
			Data:   chunkData,
			Offset: uint64(start),
			Size:   uint32(len(chunkData)),
			Hash:   HashBytes(chunkData),
		})
	}

	return chunks
}

// TokenizeFile lê um arquivo e retorna seus chunks CDC.
func TokenizeFile(path string, opts Options) ([]Chunk, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Tokenize(data, opts), nil
}

// TokenizeDir tokeniza todos os arquivos de um diretório (não recursivo).
func TokenizeDir(dir string, opts Options) (map[string][]Chunk, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	results := make(map[string][]Chunk)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		chunks, err := TokenizeFile(path, opts)
		if err != nil {
			continue // Pular arquivos não-legíveis
		}
		results[entry.Name()] = chunks
	}

	return results, nil
}

// ── Métricas ──

// ChunkStats calcula estatísticas de um conjunto de chunks.
type ChunkStats struct {
	TotalChunks  int
	TotalBytes   uint64
	UniqueChunks int
	DedupRate    float64 // Percentual de deduplicação (0.0 - 1.0)
	AvgChunkSize float64
	MinChunkSize uint32
	MaxChunkSize uint32
}

// StatsFor calcula estatísticas de um conjunto de chunks.
func StatsFor(chunks []Chunk) ChunkStats {
	if len(chunks) == 0 {
		return ChunkStats{}
	}

	seen := make(map[uint64]bool)
	var totalBytes uint64
	minSize := chunks[0].Size
	maxSize := chunks[0].Size

	for _, c := range chunks {
		seen[c.Hash] = true
		totalBytes += uint64(c.Size)
		if c.Size < minSize {
			minSize = c.Size
		}
		if c.Size > maxSize {
			maxSize = c.Size
		}
	}

	unique := len(seen)
	dedupRate := 0.0
	if len(chunks) > 0 {
		dedupRate = 1.0 - float64(unique)/float64(len(chunks))
	}

	return ChunkStats{
		TotalChunks:  len(chunks),
		TotalBytes:   totalBytes,
		UniqueChunks: unique,
		DedupRate:    dedupRate,
		AvgChunkSize: float64(totalBytes) / float64(len(chunks)),
		MinChunkSize: minSize,
		MaxChunkSize: maxSize,
	}
}
