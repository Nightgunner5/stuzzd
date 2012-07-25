package networking

import (
	"github.com/Nightgunner5/stuzzd/protocol"
	"log"
	"runtime"
	"sync"
	"time"
)

func GetBlockAt(x, y, z int32) protocol.BlockType {
	if y < 0 || y > 255 {
		return protocol.Air
	}
	chunk := GetChunkAtMark(x, z)
	defer chunk.MarkUnused()
	return chunk.GetBlock(uint8(x&0xF), uint8(y), uint8(z&0xF))
}

func GetBlockDataAt(x, y, z int32) uint8 {
	if y < 0 || y > 255 {
		return 0
	}
	chunk := GetChunkAtMark(x, z)
	defer chunk.MarkUnused()
	return chunk.GetBlockData(uint8(x&0xF), uint8(y), uint8(z&0xF))
}

func SetBlockAt(x, y, z int32, block protocol.BlockType, data uint8) {
	if y < 0 || y > 255 {
		return
	}
	chunk := GetChunkAtMark(x, z)
	defer chunk.MarkUnused()
	chunk.SetBlock(uint8(x&0xF), uint8(y), uint8(z&0xF), block)
	chunk.SetBlockData(uint8(x&0xF), uint8(y), uint8(z&0xF), data)
	startBlockUpdate(x, y, z)
	SendToAll(protocol.BlockChange{X: x, Y: uint8(y), Z: z, Block: block, Data: data})
}

func setBlockNoUpdate(x, y, z int32, block protocol.BlockType, data uint8) {
	if y < 0 || y > 255 {
		return
	}
	chunk := GetChunkAtMark(x, z)
	defer chunk.MarkUnused()
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
				}
				chunk.MarkUnused()
			}
		}
	}
}

func spreadWater(fromX, fromY, fromZ, toX, toY, toZ int32) bool {
	fromData := GetBlockDataAt(fromX, fromY, fromZ)
	if fromData&0x8 == 0x8 && fromY == toY { // This water is falling, don't make it go sideways!
		return false
	}
	if fromData == 0x7 { // No infinite spreading recursion!
		return false
	}

	toType := GetBlockAt(toX, toY, toZ)
	if !toType.Passable() {
		return false
	}
	toData := GetBlockDataAt(toX, toY, toZ)
	if toType != protocol.Water && toType != protocol.StationaryWater {
		if fromData&0x8 == 0x8 {
			toData = fromData & ^uint8(0x8)
			fromData = 0
		} else {
			toData = 0x7
			fromData++
		}
		if fromData&0x7 == 0x0 {
			SetBlockAt(fromX, fromY, fromZ, protocol.Air, 0)
		} else {
			SetBlockAt(fromX, fromY, fromZ, protocol.Water, fromData)
		}
		SetBlockAt(toX, toY, toZ, protocol.Water, toData)
		return true
	}
	if toData&0x7 == 0x0 { // Target block is already full.
		return false
	}

	if fromData&0x8 == 0x0 && fromData&0x7 == toData&0x7-1 { // Target block is nearly the same height as this block - spreading would cause infinite updates.
		return false
	}

	toData = (toData & 0x8) | (toData&0x7 - 1)
	fromData = (fromData & 0x8) | (fromData&0x7 + 1)
	if fromData&0x7 == 0x0 {
		SetBlockAt(fromX, fromY, fromZ, protocol.Air, 0)
	} else {
		SetBlockAt(fromX, fromY, fromZ, protocol.Water, fromData)
	}
	SetBlockAt(toX, toY, toZ, protocol.Water, toData)
	return true
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
			switch GetBlockAt(x, y, z) {
			case protocol.Water:
				a := spreadWater(x, y, z, x, y-1, z)
				b := spreadWater(x, y, z, x-1, y, z)
				c := spreadWater(x, y, z, x+1, y, z)
				d := spreadWater(x, y, z, x, y, z-1)
				e := spreadWater(x, y, z, x, y, z+1)
				if !a && !b && !c && !d && !e {
					setBlockNoUpdate(x, y, z, protocol.StationaryWater, GetBlockDataAt(x, y, z))
				} else {
					updateCount++
				}
			}
			delete(updateQueue, block)
			if updateCount >= 1000 {
				log.Print("> 1000 updates. Waiting for the next tick to resume updating.")
				break
			}
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
