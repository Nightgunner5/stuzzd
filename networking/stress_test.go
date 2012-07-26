package networking

import (
	"math/rand"
	"testing"
)

// Generate tons of chunks. BECAUSE WE CAN.
func BenchmarkChunkGen(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ChunkGen(0, 0)
	}
}

// If we run the lighting function ten billion times, that'll make the results more realistic, riiiiiight?
func BenchmarkChunkLighting(b *testing.B) {
	b.StopTimer()
	chunk := ChunkGen(0, 0)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		chunk.InitLighting()
	}
}

// Because compressing the same data over and over is useful for... things.
func BenchmarkCompressChunk(b *testing.B) {
	b.StopTimer()
	chunk := ChunkGen(0, 0)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		chunk.MarkDirtyForTesting()
		chunk.Compressed()
	}
}

// Make sure the algorithm doesn't do something sneaky.
func BenchmarkRandomChunks(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ChunkGen(int32(rand.Int63n(1<<32)), int32(rand.Int63n(1<<32)))
	}
}
