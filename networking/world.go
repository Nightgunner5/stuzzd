package networking

import (
	"github.com/Nightgunner5/stuzzd/protocol"
	"log"
	"runtime"
	"sync"
	"time"
)

// Wait a tick before unmarking the chunk to stop it from hitting 0 users over and over during the same tick.
func unmarkChunkDelayed(chunk *protocol.Chunk) {
	go func() {
		time.Sleep(50 * time.Millisecond)
		chunk.MarkUnused()
	}()
}

func GetBlockAt(x, y, z int32) protocol.BlockType {
	if y < 0 || y > 255 {
		return protocol.Air
	}
	chunk := GetChunkAtMark(x, z)
	defer unmarkChunkDelayed(chunk)
	return chunk.GetBlock(uint8(x&0xF), uint8(y), uint8(z&0xF))
}

func GetBlockDataAt(x, y, z int32) uint8 {
	if y < 0 || y > 255 {
		return 0
	}
	chunk := GetChunkAtMark(x, z)
	defer unmarkChunkDelayed(chunk)
	return chunk.GetBlockData(uint8(x&0xF), uint8(y), uint8(z&0xF))
}

func PlayerSetBlockAt(x, y, z int32, block protocol.BlockType, data uint8) {
	if y < 0 || y > 255 {
		return
	}
	chunk := GetChunkAtMark(x, z)
	defer unmarkChunkDelayed(chunk)
	chunk.SetBlock(uint8(x&0xF), uint8(y), uint8(z&0xF), block)
	chunk.SetBlockData(uint8(x&0xF), uint8(y), uint8(z&0xF), data)
	startBlockUpdate(x, y, z)
	SendToAll(protocol.BlockChange{X: x, Y: uint8(y), Z: z, Block: block, Data: data})
}

func SetBlockAt(x, y, z int32, block protocol.BlockType, data uint8) {
	if y < 0 || y > 255 {
		return
	}
	chunk := GetChunkAtMark(x, z)
	defer unmarkChunkDelayed(chunk)
	chunk.SetBlock(uint8(x&0xF), uint8(y), uint8(z&0xF), block)
	chunk.SetBlockData(uint8(x&0xF), uint8(y), uint8(z&0xF), data)
	startBlockUpdate(x, y, z)
	queueBlockSend(x, y, z)
}

func setBlockNoUpdate(x, y, z int32, block protocol.BlockType, data uint8) {
	if y < 0 || y > 255 {
		return
	}
	chunk := GetChunkAtMark(x, z)
	defer unmarkChunkDelayed(chunk)
	chunk.SetBlock(uint8(x&0xF), uint8(y), uint8(z&0xF), block)
	chunk.SetBlockData(uint8(x&0xF), uint8(y), uint8(z&0xF), data)
}

func startBlockUpdate(x, y, z int32) {
	for X := x - 1; X <= x+1; X++ {
		for Y := y - 1; Y <= y+1; Y++ {
			for Z := z - 1; Z <= z+1; Z++ {
				chunk := GetChunkAtMark(X, Z)
				switch chunk.GetBlock(uint8(X&0xF), uint8(Y), uint8(Z&0xF)) {
				case protocol.Water, protocol.StationaryWater:
					chunk.SetBlock(uint8(X&0xF), uint8(Y), uint8(Z&0xF), protocol.Water)
					queueUpdate(X, Y, Z)
				case protocol.Sponge, protocol.Gravel, protocol.Sand:
					queueUpdate(X, Y, Z)
				}
				chunk.MarkUnused()
			}
		}
	}
}

func incrementWater(x, y, z int32) {
	level := GetBlockDataAt(x, y, z)
	if block := GetBlockAt(x, y, z); block != protocol.Water && block != protocol.StationaryWater { // No water here yet
		SetBlockAt(x, y, z, protocol.Water, 0x7)
		return
	}
	if level&0x7 == 0x0 { // Already full
		return
	}
	SetBlockAt(x, y, z, protocol.Water, level&0x8|(level&0x7-1)&0x7)
}
func decrementWater(x, y, z int32) {
	level := GetBlockDataAt(x, y, z)
	if block := GetBlockAt(x, y, z); block != protocol.Water && block != protocol.StationaryWater { // No water here
		return
	}
	if level&0x7 == 0x7 { // No water left
		SetBlockAt(x, y, z, protocol.Air, 0)
		return
	}
	SetBlockAt(x, y, z, protocol.Water, level&0x8|(level&0x7+1)&0x7)
}

func getWaterLevel(x, y, z int32) uint8 {
	block := GetBlockAt(x, y, z)
	if block == protocol.Water || block == protocol.StationaryWater {
		return 8 - (GetBlockDataAt(x, y, z) & 0x7)
	}
	if block.Passable() {
		return 0
	}
	return ^uint8(0)
}

func spreadWater(x, y, z int32) bool {
	here := getWaterLevel(x, y, z)
	xneg := getWaterLevel(x-1, y, z)
	xpos := getWaterLevel(x+1, y, z)
	zneg := getWaterLevel(x, y, z-1)
	zpos := getWaterLevel(x, y, z+1)
	down := getWaterLevel(x, y-1, z)

	if here == 0 || here > 8 { // No water here
		return false
	}

	change := false
	for i := 0; i < 8 && here > 0; i++ {
		if down < 8 {
			down++
			incrementWater(x, y-1, z)
			here--
			decrementWater(x, y, z)
			change = true
			continue
		}
		if xpos < here-1 && xpos <= xneg && xpos <= zpos && xpos <= zneg {
			xpos++
			incrementWater(x+1, y, z)
			here--
			decrementWater(x, y, z)
			change = true
			continue
		}
		if xneg < here-1 && xneg <= xpos && xneg <= zpos && xneg <= zneg {
			xneg++
			incrementWater(x-1, y, z)
			here--
			decrementWater(x, y, z)
			change = true
			continue
		}
		if zpos < here-1 && zpos <= xpos && zpos <= xneg && zpos <= zneg {
			zpos++
			incrementWater(x, y, z+1)
			here--
			decrementWater(x, y, z)
			change = true
			continue
		}
		if zneg < here-1 && zneg <= xpos && zneg <= xneg && zneg <= zpos {
			zneg++
			incrementWater(x, y, z-1)
			here--
			decrementWater(x, y, z)
			change = true
			continue
		}

		if xneg < xpos-1 && xneg < 8 && xpos > 0 && xpos <= 8 {
			xneg++
			incrementWater(x-1, y, z)
			xpos--
			decrementWater(x+1, y, z)
			change = true
			continue
		}
		if xpos < xneg-1 && xpos < 8 && xneg > 0 && xneg <= 8 {
			xpos++
			incrementWater(x+1, y, z)
			xneg--
			decrementWater(x-1, y, z)
			change = true
			continue
		}
		if zneg < zpos-1 && zneg < 8 && zpos > 0 && zpos <= 8 {
			zneg++
			incrementWater(x, y, z-1)
			zpos--
			decrementWater(x, y, z+1)
			change = true
			continue
		}
		if zpos < zneg-1 && zpos < 8 && zneg > 0 && zneg <= 8 {
			zpos++
			incrementWater(x, y, z+1)
			zneg--
			decrementWater(x, y, z-1)
			change = true
			continue
		}
		break
	}
	// Get rid of the tiny bit of water that stays there forever
	if here == 1 {
		if getWaterLevel(x-1, y-1, z) < 8 {
			incrementWater(x-1, y-1, z)
			decrementWater(x, y, z)
		} else if getWaterLevel(x+1, y-1, z) < 8 {
			incrementWater(x+1, y-1, z)
			decrementWater(x, y, z)
		} else if getWaterLevel(x, y-1, z-1) < 8 {
			incrementWater(x, y-1, z-1)
			decrementWater(x, y, z)
		} else if getWaterLevel(x, y-1, z+1) < 8 {
			incrementWater(x, y-1, z+1)
			decrementWater(x, y, z)
		}
	}
	return change
}

var blockSendQueue = make(map[struct{ x, z int32 }]map[struct{ x, y, z uint8 }]bool)

func queueBlockSend(x, y, z int32) {
	chunk := struct{ x, z int32 }{x >> 4, z >> 4}
	if blockSendQueue[chunk] == nil {
		blockSendQueue[chunk] = make(map[struct{ x, y, z uint8 }]bool)
	}
	blockSendQueue[chunk][struct{ x, y, z uint8 }{uint8(x & 0xF), uint8(y), uint8(z & 0xF)}] = true
}

var updateQueue = make(map[struct{ x, y, z int32 }]bool)

func queueUpdate(x, y, z int32) {
	updateQueue[struct{ x, y, z int32 }{x, y, z}] = true
}

func ticker() {
	for {
		time.Sleep(50 * time.Millisecond)

		updateCount := 0

		for block, _ := range updateQueue {
			x, y, z := block.x, block.y, block.z
			blockType := GetBlockAt(x, y, z)
			switch blockType {
			case protocol.Water:
				if !spreadWater(x, y, z) && GetBlockAt(x, y, z) == protocol.Water {
					setBlockNoUpdate(x, y, z, protocol.StationaryWater, GetBlockDataAt(x, y, z))
				}
			case protocol.Sand, protocol.Gravel:
				if GetBlockAt(x, y-1, z).Passable() {
					SetBlockAt(x, y, z, GetBlockAt(x, y-1, z), GetBlockDataAt(x, y-1, z))
					SetBlockAt(x, y-1, z, blockType, 0)
				}
			case protocol.Sponge:
				switch GetBlockAt(x, y+1, z) {
				case protocol.Water, protocol.StationaryWater:
					decrementWater(x, y+1, z)
				}
			}
			updateCount++
			delete(updateQueue, block)
			runtime.Gosched() // Don't cause too much lag
			if updateCount >= 1000 {
				log.Print("> 1000 updates. Waiting for the next tick to resume updating.")
				break
			}
		}

		for chunk, blocks := range blockSendQueue {
			c := GetChunkMark(chunk.x, chunk.z)
			packet := protocol.MultiBlockChange{X: chunk.x, Z: chunk.z, Blocks: make([]uint32, 0, len(blocks))}
			for block, _ := range blocks {
				packet.Blocks = append(packet.Blocks, uint32(block.x)<<28|uint32(block.z)<<24|uint32(block.y)<<16|uint32(c.GetBlock(block.x, block.y, block.z))<<4|uint32(c.GetBlockData(block.x, block.y, block.z)))
			}
			c.MarkUnused()
			SendToAll(packet)
			delete(blockSendQueue, chunk)
		}
	}
}

func init() {
	go ticker()
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

func GetChunkAtMark(x, z int32) *protocol.Chunk {
	return GetChunkMark(x>>4, z>>4)
}

//func GetChunk(x, z int32) *protocol.Chunk {
//	id := uint64(uint32(x))<<32 | uint64(uint32(z)) // Yes, this is required.
//	chunkLock.RLock()
//	if chunk, ok := chunks[id]; ok {
//		chunkLock.RUnlock()
//		return chunk
//	}
//	chunkLock.RUnlock()
//
//	chunkLock.Lock()
//	defer chunkLock.Unlock()
//
//	chunks[id] = loadChunk(x, z)
//	return chunks[id]
//}

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
