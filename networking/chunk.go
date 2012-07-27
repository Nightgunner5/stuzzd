package networking

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"github.com/Nightgunner5/stuzzd/protocol"
	"sync"
	"time"
)

type Chunk struct {
	blocks           protocol.BlockChunk
	blockData        protocol.NibbleChunk
	lightBlock       protocol.NibbleChunk
	lightSky         protocol.NibbleChunk
	biomes           [16][16]protocol.Biome
	dirty            bool
	needsSave        bool
	compressed       []byte
	lock             sync.RWMutex
	users            int64
	interruptRecycle <-chan bool
}

func (c *Chunk) SetBlock(x, y, z uint8, block protocol.BlockType) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.blocks.Set(x, y, z, block)
	c.dirty = true
	c.needsSave = true
}

func (c *Chunk) GetBlock(x, y, z uint8) protocol.BlockType {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.blocks.Get(x, y, z)
}

func (c *Chunk) SetBlockData(x, y, z, data uint8) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.blockData.Set(x, y, z, data)
	c.dirty = true
	c.needsSave = true
}

func (c *Chunk) GetBlockData(x, y, z uint8) uint8 {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.blockData.Get(x, y, z)
}

func (c *Chunk) SetBiome(x, z uint8, biome protocol.Biome) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.biomes[z][x] = biome
	c.dirty = true
	c.needsSave = true
}

func (c *Chunk) GetBiome(x, z uint8) protocol.Biome {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.biomes[z][x]
}

func (c *Chunk) InitLighting() {
	c.lock.Lock()
	defer c.lock.Unlock()

	for x := uint8(0); x < 16; x++ {
		for z := uint8(0); z < 16; z++ {
			brightness := uint8(15)
			for y := uint8(255); y >= 0 && brightness > 0 && brightness < 16; y-- {
				c.lightSky.Set(x, y, z, brightness)
				switch c.blocks.Get(x, y, z) { // Skip the lock
				case protocol.Air, protocol.Glass:
					// Don't change the brightness
				case protocol.Water, protocol.StationaryWater:
					brightness--
				default:
					brightness -= 4
				}
			}
		}
	}

	c.dirty = true
	c.needsSave = true
}

func panicIfError(n int, err error) {
	if err != nil {
		panic(err)
	}
}

func (c *Chunk) Compressed() []byte {
	c.lock.RLock()
	if !c.dirty && c.compressed != nil {
		defer c.lock.RUnlock()
		return c.compressed
	}
	c.lock.RUnlock()

	c.lock.Lock()
	defer c.lock.Unlock()
	if !c.dirty && c.compressed != nil {
		return c.compressed
	}
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)

	for _, blocks := range c.blocks {
		panicIfError(0, binary.Write(w, binary.BigEndian, blocks))
	}
	for _, data := range c.blockData {
		panicIfError(w.Write(data[:]))
	}
	for _, light := range c.lightBlock {
		panicIfError(w.Write(light[:]))
	}
	for _, light := range c.lightSky {
		panicIfError(w.Write(light[:]))
	}
	for _, biomes := range c.biomes {
		panicIfError(0, binary.Write(w, binary.BigEndian, biomes))
	}

	w.Close()
	c.compressed = buf.Bytes()
	c.dirty = false

	return c.compressed
}

func (c *Chunk) MarkDirtyForTesting() {
	c.dirty = true
}

func (c *Chunk) MarkUsed() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.users++
	if c.interruptRecycle != nil {
		<-c.interruptRecycle
		c.interruptRecycle = nil
	}
}

func (c *Chunk) MarkUnused() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.users--
	if c.users == 0 {
		interrupt := make(chan bool, 1)
		interrupt <- false
		c.interruptRecycle = interrupt
		go func() {
			time.Sleep(30 * time.Second)
			<-interrupt
			recycleChunk(c)
		}()
	}
	if c.users < 0 {
		panic("Use count < 0")
	}
}

func (c *Chunk) Save() {
	c.lock.Lock()
	defer c.lock.Unlock()

	if !c.needsSave {
		return
	}

	// TODO: actually save the chunk
	c.needsSave = false
}

func (c *Chunk) decode(stored map[string]interface{}) {
	for _, section := range stored["Level"].(map[string]interface{})["Sections"].([]interface{}) {
		sec := section.(map[string]interface{})
		if sec["Y"] == nil {
			continue // Bug?
		}
		y := sec["Y"].(int8)

		for i, block := range sec["Blocks"].([]byte) {
			c.blocks[y][i] = protocol.BlockType(block)
		}
		copy(c.blockData[y][:], sec["Data"].([]byte))
		copy(c.lightSky[y][:], sec["SkyLight"].([]byte))
		copy(c.lightBlock[y][:], sec["BlockLight"].([]byte))
	}
	for i, biome := range stored["Level"].(map[string]interface{})["Biomes"].([]byte) {
		c.biomes[i&0xF][i>>4] = protocol.Biome(biome)
	}
}
