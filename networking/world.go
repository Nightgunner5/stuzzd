package networking

import (
	"github.com/Nightgunner5/stuzzd/protocol"
)

var debugChunk protocol.Chunk

func init() {
	for x := uint8(0); x < 16; x++ {
		for z := uint8(0); z < 16; z++ {
			debugChunk.SetBlock(x, 0, z, protocol.Bedrock)
			for y := uint8(1); y < 40; y++ {
				debugChunk.SetBlock(x, y, z, protocol.Stone)
			}
			for y := uint8(40); y < 63; y++ {
				debugChunk.SetBlock(x, y, z, protocol.Dirt)
			}
			debugChunk.SetBlock(x, 63, z, protocol.Wool)
			debugChunk.SetBlockData(x, 63, z, 2)
			for y := uint16(63); y < 256; y++ {
				debugChunk.LightSky.Set(x, uint8(y), z, 15)
			}
		}
	}
}

func GetBlockAt(x, y, z int32) protocol.BlockType {
	return GetChunk(x>>4, z>>4).GetBlock(uint8(x&0xF), uint8(y), uint8(z&0xF))
}

func GetChunk(x, z int32) *protocol.Chunk {
	return &debugChunk
}
