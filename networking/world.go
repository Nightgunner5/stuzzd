package networking

import (
	"github.com/Nightgunner5/stuzzd/protocol"
	"github.com/Nightgunner5/stuzzd/util"
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
	chunk := new(protocol.Chunk)

	for x := uint8(0); x < 16; x++ {
		for z := uint8(0); z < 16; z++ {
			chunk.SetBlock(x, 0, z, protocol.Bedrock)
			change1 := uint8(40 + 4 * util.Noise2(float64(x) / 16 + float64(chunkX), float64(z) / 16 + float64(chunkZ)))
			change2 := uint8(58 + 8 * util.Noise2((float64(x) / 16 + float64(chunkX)) / 10, (float64(z) / 16 + float64(chunkZ)) / 10))
			for y := uint8(1); y < change1; y++ {
				chunk.SetBlock(x, y, z, protocol.Stone)
			}
			for y := change1; y < change2; y++ {
				chunk.SetBlock(x, y, z, protocol.Dirt)
			}
			chunk.SetBlock(x, change2, z, protocol.Grass)
			chunk.SetBiome(x, z, protocol.Plains)
		}
	}
	chunk.InitLighting()

	return chunk
}

func InitSpawnArea() {
	for x := int32(-8); x < 8; x++ {
		for z := int32(-8); z < 8; z++ {
			GetChunk(x, z).Compressed()
		}
	}
}

var chunks = make(map[uint64]*protocol.Chunk)
var chunkLock sync.RWMutex

func GetChunkAt(x, z int32) *protocol.Chunk {
	if x < 0 {
		x -= 15
	}
	if z < 0 {
		z -= 15
	}
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
