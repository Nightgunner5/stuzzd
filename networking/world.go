package networking

import (
	"bytes"
	"github.com/Nightgunner5/stuzzd/block"
	"github.com/Nightgunner5/stuzzd/chunk"
	"github.com/Nightgunner5/stuzzd/player"
	"github.com/Nightgunner5/stuzzd/protocol"
	"github.com/Nightgunner5/stuzzd/storage"
	"log"
	"runtime"
	"sort"
	"sync"
	"time"
)

func GetBlockAt(x, y, z int32) block.BlockType {
	if y < 0 || y > 255 {
		return block.Air
	}
	chunk := storage.GetChunkContaining(x, z)
	defer storage.ReleaseChunkContaining(x, z)

	return chunk.GetBlock(x, y, z)
}

func GetBlockDataAt(x, y, z int32) uint8 {
	if y < 0 || y > 255 {
		return 0
	}
	chunk := storage.GetChunkContaining(x, z)
	defer storage.ReleaseChunkContaining(x, z)

	return chunk.GetData(x, y, z)
}

func PlayerSetBlockAt(x, y, z int32, block block.BlockType, data uint8) {
	if y < 0 || y > 255 {
		return
	}
	chunk := storage.GetChunkContaining(x, z)
	defer storage.ReleaseChunkContaining(x, z)

	chunk.SetBlock(x, y, z, block)
	chunk.SetData(x, y, z, data)

	startBlockUpdate(x, y, z)
	SendToAll(protocol.BlockChange{X: x, Y: uint8(y), Z: z, Block: block, Data: data})
}

func SetBlockAt(x, y, z int32, block block.BlockType, data uint8) {
	if y < 0 || y > 255 {
		return
	}
	chunk := storage.GetChunkContaining(x, z)
	defer storage.ReleaseChunkContaining(x, z)

	chunk.SetBlock(x, y, z, block)
	chunk.SetData(x, y, z, data)

	startBlockUpdate(x, y, z)
	queueBlockSend(x, y, z)
}

func DropItem(x, y, z float64, itemType int16, data uint8) {
	c := storage.GetChunkContaining(int32(x), int32(z))
	defer storage.ReleaseChunkContaining(int32(x), int32(z))

	ent := chunk.NewItemDrop(x, y, z, &player.InventoryItem{Type: itemType, Damage: int16(data), Count: 1})
	c.SpawnEntity(chunk.Entity(ent))

	var buf bytes.Buffer
	ent.SpawnPacket(&buf)
	SendToAllNearChunk(c.X, c.Z, protocol.BakedPacket(buf.Bytes()))
}

func setBlockNoUpdate(x, y, z int32, block block.BlockType, data uint8) {
	if y < 0 || y > 255 {
		return
	}
	chunk := storage.GetChunkContaining(x, z)
	defer storage.ReleaseChunkContaining(x, z)

	chunk.SetBlock(x, y, z, block)
	chunk.SetData(x, y, z, data)
}

func startBlockUpdate(x, y, z int32) {
	for X := x - 1; X <= x+1; X++ {
		for Y := y - 1; Y <= y+1; Y++ {
			for Z := z - 1; Z <= z+1; Z++ {
				chunk := storage.GetChunkContaining(X, Z)
				switch chunk.GetBlock(X, Y, Z) {
				case block.Water, block.StationaryWater:
					chunk.SetBlock(X, Y, Z, block.Water)
					queueUpdate(X, Y, Z)
				case block.Sponge:
					queueUpdate(X, Y, Z)
				case block.Gravel, block.Sand, block.LongGrass, block.RedFlower, block.YellowFlower:
					queueUpdate(X, Y, Z)
				}
				storage.ReleaseChunkContaining(X, Z)
			}
		}
	}
}

func incrementWater(x, y, z int32) {
	level := GetBlockDataAt(x, y, z)
	if b := GetBlockAt(x, y, z); b != block.Water && b != block.StationaryWater { // No water here yet
		SetBlockAt(x, y, z, block.Water, 0x7)
		return
	}
	if level&0x7 == 0x0 { // Already full
		return
	}
	SetBlockAt(x, y, z, block.Water, level&0x8|(level&0x7-1)&0x7)
}
func decrementWater(x, y, z int32) {
	level := GetBlockDataAt(x, y, z)
	if b := GetBlockAt(x, y, z); b != block.Water && b != block.StationaryWater { // No water here
		return
	}
	if level&0x7 == 0x7 { // No water left
		SetBlockAt(x, y, z, block.Air, 0)
		return
	}
	SetBlockAt(x, y, z, block.Water, level&0x8|(level&0x7+1)&0x7)
}

func getWaterLevel(x, y, z int32) uint8 {
	b := GetBlockAt(x, y, z)
	if b == block.Water || b == block.StationaryWater {
		return 8 - (GetBlockDataAt(x, y, z) & 0x7)
	}
	if b.Passable() {
		return 0
	}
	return ^uint8(0)
}

type waterLevel struct {
	x, y, z int32
	level   uint8
}
type waterLevels []waterLevel

func (w waterLevels) Len() int {
	return len(w)
}
func (w waterLevels) Less(a, b int) bool {
	return w[a].level < w[b].level
}
func (w waterLevels) Swap(a, b int) {
	w[a], w[b] = w[b], w[a]
}

func getWaterLevels(x, y, z int32) waterLevels {
	levels := make(waterLevels, 0)
	for Y := y - 1; Y <= y; Y++ {
		for X := x - 1; X <= x+1; X++ {
			for Z := z - 1; Z <= z+1; Z++ {
				if x == X && y == Y && z == Z {
					continue
				}
				level := getWaterLevel(X, Y, Z)
				if level < 8 {
					levels = append(levels, waterLevel{
						x: X, y: Y, z: Z,
						level: level,
					})
				}
			}
		}

		if len(levels) != 0 {
			break
		}
	}
	sort.Sort(levels)
	return levels
}

func spreadWater(x, y, z int32) bool {
	here := getWaterLevel(x, y, z)
	levels := getWaterLevels(x, y, z)

	if here == 0 || here > 8 { // No water here
		return false
	}

	change := false
	for i := 0; i < 3 && here > 0; i++ {
		if len(levels) == 0 || levels[0].level >= 8 {
			break
		}

		if levels[0].level < here-1 || levels[0].y < y {
			here--
			decrementWater(x, y, z)
			levels[0].level++
			incrementWater(levels[0].x, levels[0].y, levels[0].z)
			sort.Sort(levels)
			change = true
		} else {
			break
		}
	}
	return change
}

var blockSendQueue = make(map[struct{ x, z int32 }]map[struct{ x, y, z int32 }]bool)
var blockSendLock sync.Mutex

func queueBlockSend(x, y, z int32) {
	blockSendLock.Lock()
	defer blockSendLock.Unlock()

	chunk := struct{ x, z int32 }{x >> 4, z >> 4}
	if blockSendQueue[chunk] == nil {
		blockSendQueue[chunk] = make(map[struct{ x, y, z int32 }]bool)
	}
	blockSendQueue[chunk][struct{ x, y, z int32 }{x, y, z}] = true
}

var updateQueue = make(map[struct{ x, y, z int32 }]bool)
var updateLock sync.Mutex

func queueUpdate(x, y, z int32) {
	updateLock.Lock()
	defer updateLock.Unlock()

	updateQueue[struct{ x, y, z int32 }{x, y, z}] = true
}

func ticker() {
	for {
		time.Sleep(50 * time.Millisecond)

		updateLock.Lock()

		queue := updateQueue
		updateQueue = make(map[struct{ x, y, z int32 }]bool)

		updateLock.Unlock()

		updateCount := 0

		for loc, _ := range queue {
			x, y, z := loc.x, loc.y, loc.z
			blockType := GetBlockAt(x, y, z)
			switch blockType {
			case block.Water:
				if !spreadWater(x, y, z) && GetBlockAt(x, y, z) == block.Water {
					setBlockNoUpdate(x, y, z, block.StationaryWater, GetBlockDataAt(x, y, z))
				}
			case block.Sand, block.Gravel, block.LongGrass, block.RedFlower, block.YellowFlower:
				if GetBlockAt(x, y-1, z).Passable() {
					blockData := GetBlockDataAt(x, y, z)
					SetBlockAt(x, y, z, GetBlockAt(x, y-1, z), GetBlockDataAt(x, y-1, z))
					SetBlockAt(x, y-1, z, blockType, blockData)
				}
			case block.Sponge:
				switch GetBlockAt(x, y+1, z) {
				case block.Water, block.StationaryWater:
					decrementWater(x, y+1, z)
				}
			}
			updateCount++
			delete(queue, loc)
			runtime.Gosched() // Don't cause too much lag
			if updateCount >= 10000 {
				log.Print("> 10000 updates. Waiting for the next tick to resume updating.")
				break
			}
		}

		for loc, _ := range queue {
			queueUpdate(loc.x, loc.y, loc.z)
		}

		blockSendLock.Lock()
		for chunk, blocks := range blockSendQueue {
			c := storage.GetChunk(chunk.x, chunk.z)
			packet := protocol.MultiBlockChange{X: chunk.x, Z: chunk.z, Blocks: make([]uint32, 0, len(blocks))}
			for block, _ := range blocks {
				packet.Blocks = append(packet.Blocks, uint32(block.x&0xF)<<28|uint32(block.z&0xF)<<24|uint32(block.y)<<16|uint32(c.GetBlock(block.x, block.y, block.z))<<4|uint32(c.GetData(block.x, block.y, block.z)))
			}
			storage.ReleaseChunk(chunk.x, chunk.z)
			SendToAll(packet)
			delete(blockSendQueue, chunk)
		}
		blockSendLock.Unlock()
	}
}

func init() {
	go ticker()
}
