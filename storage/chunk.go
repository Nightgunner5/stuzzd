package storage

import (
	"fmt"
	"github.com/Nightgunner5/stuzzd/chunk"
	"runtime"
	"sync"
	"time"
)

func loadChunk(chunkX, chunkZ int32) *chunk.Chunk {
	chunk, err := ReadChunk(chunkX, chunkZ)
	if err != nil {
		return ChunkGen(chunkX, chunkZ)
	}
	return chunk
}

func InitSpawnArea() {
	for x := int32(-8); x < 8; x++ {
		for z := int32(-8); z < 8; z++ {
			runtime.Gosched() // We want to accept connections while we start up, even on GOMAXPROCS=1.
			chunk := GetChunk(x, z)
			chunk.Packet()
			// Keep the chunks (don't release them) as they are the spawn chunks and are used very frequently.
			WriteChunk(chunk)
		}
	}
}

var chunks = make(map[uint64]*chunk.Chunk)
var chunkLock sync.RWMutex
var users = make(map[uint64]uint64)
var userLock sync.Mutex

func GetChunkContaining(x, z int32) *chunk.Chunk {
	return GetChunk(x>>4, z>>4)
}

func GetChunk(x, z int32) *chunk.Chunk {
	id := uint64(uint32(x))<<32 | uint64(uint32(z))
	chunkLock.RLock()
	if chunk, ok := chunks[id]; ok {
		userLock.Lock()
		users[id]++
		userLock.Unlock()
		chunkLock.RUnlock()
		return chunk
	}
	chunkLock.RUnlock()

	chunkLock.Lock()
	defer chunkLock.Unlock()

	chunks[id] = loadChunk(x, z)
	userLock.Lock()
	users[id]++
	userLock.Unlock()
	return chunks[id]
}

func ReleaseChunk(x, z int32) {
	userLock.Lock()
	defer userLock.Unlock()
	id := uint64(uint32(x))<<32 | uint64(uint32(z))
	if users[id] == 0 {
		panic(fmt.Sprintf("User count for chunk %d, %d is less than zero!", x, z))
	}
	users[id]--
}

func ReleaseChunkContaining(x, z int32) {
	ReleaseChunk(x>>4, z>>4)
}

func init() {
	go chunkRecycler()
}

func chunkRecycler() {
	for {
		time.Sleep(2 * time.Minute)

		chunkLock.Lock()
		toSave := make([]*chunk.Chunk, 0, len(chunks))

		userLock.Lock()
		for id, chunk := range chunks {
			toSave = append(toSave, chunk)
			if users[id] == 0 {
				delete(chunks, id)
			}
		}
		userLock.Unlock()
		chunkLock.Unlock()

		for _, chunk := range toSave {
			go WriteChunk(chunk) // They will save in parallel as long as they are in different regions. If not, they will wait for their turn.
		}
	}
}

func SaveAndUnloadAllChunks() {
	chunkLock.Lock()
	defer chunkLock.Unlock()

	userLock.Lock()
	defer userLock.Unlock()

	users = make(map[uint64]uint64) // Almost definitely faster than looping through it.

	var wg sync.WaitGroup
	for _, chunk := range chunks {
		wg.Add(1)
		go func() {
			WriteChunk(chunk)
			wg.Done()
		}()
	}

	chunks = make(map[uint64]*chunk.Chunk)

	wg.Wait() // Don't return until all the chunks are written or the server could stop while writing a chunk or before all chunks are written.
}
