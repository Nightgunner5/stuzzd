package networking

import (
	"github.com/Nightgunner5/stuzzd/protocol"
	"runtime"
	"sync"
)

func GetBlockAt(x, y, z int32) protocol.BlockType {
	return GetChunkAt(x, z).GetBlock(uint8(x&0xF), uint8(y), uint8(z&0xF))
}

func SetBlockAt(x, y, z int32, block protocol.BlockType, data uint8) {
	chunk := GetChunkAt(x, z)
	chunk.SetBlock(uint8(x&0xF), uint8(y), uint8(z&0xF), block)
	chunk.SetBlockData(uint8(x&0xF), uint8(y), uint8(z&0xF), data)
	SendToAll(protocol.BlockChange{X: x, Y: uint8(y), Z: z, Block: block, Data: data})
}

func loadChunk(chunkX, chunkZ int32) *protocol.Chunk {
	return ChunkGen(chunkX, chunkZ)
}

func InitSpawnArea() {
	for x := int32(-8); x < 8; x++ {
		for z := int32(-8); z < 8; z++ {
			runtime.Gosched() // We want to accept connections while we start up, even on GOMAXPROCS=1.
			chunk := GetChunkMark(x, z)
			chunk.Compressed()
			// Don't release the chunks so connecting the game will be faster for new players.
		}
	}
}

var chunks = make(map[uint64]*protocol.Chunk)
var chunkLock sync.RWMutex

func GetChunkAt(x, z int32) *protocol.Chunk {
	return GetChunk(x>>4, z>>4)
}

func GetChunk(x, z int32) *protocol.Chunk {
	id := uint64(uint32(x))<<32 | uint64(uint32(z)) // Yes, this is required.
	chunkLock.RLock()
	if chunk, ok := chunks[id]; ok {
		chunkLock.RUnlock()
		return chunk
	}
	chunkLock.RUnlock()

	chunkLock.Lock()
	defer chunkLock.Unlock()

	chunks[id] = loadChunk(x, z)
	return chunks[id]
}

func GetChunkMark(x, z int32) *protocol.Chunk {
	id := uint64(uint32(x))<<32 | uint64(uint32(z))
	chunkLock.RLock()
	if chunk, ok := chunks[id]; ok {
		chunk.MarkUsed()
		chunkLock.RUnlock()
		return chunk
	}
	chunkLock.RUnlock()

	chunkLock.Lock()
	defer chunkLock.Unlock()

	chunks[id] = loadChunk(x, z)
	chunks[id].MarkUsed()
	return chunks[id]
}

func init() {
	protocol.RecycleChunk = func(c *protocol.Chunk) {
		chunkLock.Lock()
		defer chunkLock.Unlock()

		for id, chunk := range chunks {
			if chunk == c {
				c.Save()
				delete(chunks, id)
				return
			}
		}
	}
}
