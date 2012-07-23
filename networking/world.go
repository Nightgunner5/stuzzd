package networking

import (
	"github.com/Nightgunner5/stuzzd/protocol"
	"github.com/Nightgunner5/stuzzd/util"
	"sync"
)

func GetBlockAt(x, y, z int32) protocol.BlockType {
	return GetChunk(x>>4, z>>4).GetBlock(uint8(x&0xF), uint8(y), uint8(z&0xF))
}

func SetBlockAt(x, y, z int32, block protocol.BlockType, data uint8) {
	chunk := GetChunk(x>>4, z>>4)
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
			change2 := uint8(60 + 2 * util.Noise2((float64(x) / 16 + float64(chunkX)) / 10 + 153420321.4644, (float64(z) / 16 + float64(chunkZ)) / 10 - 54413135.0542))
			for y := uint8(1); y < change1; y++ {
				chunk.SetBlock(x, y, z, protocol.Stone)
			}
			for y := change1; y < change2; y++ {
				chunk.SetBlock(x, y, z, protocol.Dirt)
			}
			chunk.SetBlock(x, change2, z, protocol.Grass)
			for y := uint16(0); y < 256; y++ {
				chunk.LightSky.Set(x, uint8(y), z, 15)
			}
			chunk.Biomes[x][z] = protocol.Plains
		}
	}

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

func GetChunk(x, z int32) *protocol.Chunk {
	id := uint64(x)<<32 | uint64(z)
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
