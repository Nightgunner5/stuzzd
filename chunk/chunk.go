package chunk

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"github.com/Nightgunner5/stuzzd/block"
	"github.com/Nightgunner5/stuzzd/protocol"
	"reflect"
	"sync"
)

const MAX_HEIGHT int32 = 255

type Chunk struct {
	X int32 `nbt:"xPos"`
	Z int32 `nbt:"zPos"`

	TerrainPopulated bool

	LastUpdate   uint64
	Sections     SectionList
	Entities     []Entity
	TileEntities []TileEntity
	TileTicks    []TileTick
	Biomes       [256]protocol.Biome
	HeightMap    [256]int32

	lock          sync.RWMutex `nbt:"-"`
	lightingDirty bool         `nbt:"-"`
	NeedsSave     bool         `nbt:"-"`
	packetDirty   bool         `nbt:"-"`
	packet        []byte       `nbt:"-"`
}

func (c *Chunk) GetBlock(x, y, z int32) block.BlockType {
	if y < 0 || y > MAX_HEIGHT {
		return block.Air
	}
	if x>>4 != c.X || z>>4 != c.Z {
		panic(fmt.Sprintf("GetBlock() called on chunk %d, %d but should have been called on chunk %d, %d!", c.X, c.Z, x>>4, z>>4))
	}

	c.lock.RLock()
	defer c.lock.RUnlock()
	if !c.Sections.Has(byte(y >> 4)) {
		return block.Air
	}

	return c.Sections.Get(byte(y>>4)).Blocks.Get(x, y, z)
}

func (c *Chunk) SetBlock(x, y, z int32, blockType block.BlockType) {
	if y < 0 || y > MAX_HEIGHT {
		return
	}
	if x>>4 != c.X || z>>4 != c.Z {
		panic(fmt.Sprintf("SetBlock() called on chunk %d, %d but should have been called on chunk %d, %d!", c.X, c.Z, x>>4, z>>4))
	}

	c.lock.Lock()
	defer c.lock.Unlock()
	if blockType == block.Air && !c.Sections.Has(byte(y>>4)) {
		// Don't allocate a section and then deallocate it a few lines later in the same
		// function for no reason.
		return
	}

	section := c.Sections.Get(byte(y >> 4))
	if section.Blocks.Get(x, y, z) == blockType {
		// Don't bother assigning memory to itself and marking the chunk as dirty if nothing
		// actually happened.
		return
	}

	section.Blocks.Set(x, y, z, blockType)

	if blockType == block.Air {
		c.Sections.Compact(byte(y >> 4))
	}

	if y >= c.HeightMap[(z&0xF)<<4|(x&0xF)] {
		if blockType == block.Air {
			c.recalculateHeight(x, z)
		} else {
			c.HeightMap[(z&0xF)<<4|(x&0xF)] = y
		}
	}

	c.dirty()
}

func (c *Chunk) recalculateHeight(x, z int32) {
	for y := MAX_HEIGHT; y >= 0; y-- {
		if c.Sections.Has(byte(y >> 4)) {
			if c.Sections.Get(byte(y>>4)).Blocks.Get(x, y, z) != block.Air {
				c.HeightMap[(z&0xF)<<4|(x&0xF)] = y
				return
			}
		}
	}
	c.HeightMap[(z&0xF)<<4|(x&0xF)] = 0
}

func (c *Chunk) GetData(x, y, z int32) uint8 {
	if y < 0 || y > MAX_HEIGHT {
		return 0
	}
	if x>>4 != c.X || z>>4 != c.Z {
		panic(fmt.Sprintf("GetData() called on chunk %d, %d but should have been called on chunk %d, %d!", c.X, c.Z, x>>4, z>>4))
	}

	c.lock.RLock()
	defer c.lock.RUnlock()
	if !c.Sections.Has(byte(y >> 4)) {
		return 0
	}

	return c.Sections.Get(byte(y>>4)).Data.Get(x, y, z)
}

func (c *Chunk) SetData(x, y, z int32, data uint8) {
	if y < 0 || y > MAX_HEIGHT {
		return
	}
	if x>>4 != c.X || z>>4 != c.Z {
		panic(fmt.Sprintf("SetData() called on chunk %d, %d but should have been called on chunk %d, %d!", c.X, c.Z, x>>4, z>>4))
	}

	c.lock.Lock()
	defer c.lock.Unlock()
	if !c.Sections.Has(byte(y >> 4)) {
		// Don't allocate a section just to set the data on air.
		return
	}

	section := c.Sections.Get(byte(y >> 4))
	if section.Data.Get(x, y, z) == data {
		// Don't bother assigning memory to itself and marking the chunk as dirty if nothing
		// actually happened.
		return
	}

	section.Data.Set(x, y, z, data)

	c.dirty()
}

func (c *Chunk) GetBiome(x, z int32) protocol.Biome {
	if x>>4 != c.X || z>>4 != c.Z {
		panic(fmt.Sprintf("GetBiome() called on chunk %d, %d but should have been called on chunk %d, %d!", c.X, c.Z, x>>4, z>>4))
	}

	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.Biomes[(z&0xF)<<4|(x&0xF)]
}

func (c *Chunk) SetBiome(x, z int32, biome protocol.Biome) {
	if x>>4 != c.X || z>>4 != c.Z {
		panic(fmt.Sprintf("SetBiome() called on chunk %d, %d but should have been called on chunk %d, %d!", c.X, c.Z, x>>4, z>>4))
	}

	c.lock.RLock()
	defer c.lock.RUnlock()

	c.Biomes[(z&0xF)<<4|(x&0xF)] = biome

	c.dirtyGeneric()
}

func (c *Chunk) dirty() {
	c.lightingDirty = true
	c.dirtyGeneric()
}

func (c *Chunk) dirtyGeneric() {
	c.packetDirty = true
	c.NeedsSave = true
}

func (c *Chunk) InitLighting() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.processLighting()
}

func (c *Chunk) processLighting() {
	// Zero out the current values
	for Y := byte(0); Y < 16; Y++ {
		if c.Sections.Has(Y) {
			section := c.Sections.Get(Y)
			for i := range section.SkyLight {
				section.SkyLight[i] = 0
			}
		}
	}

	for x := c.X << 4; x < (c.X<<4)+16; x++ {
		for z := c.Z << 4; z < (c.Z<<4)+16; z++ {
			for y := int32(255); y >= c.HeightMap[(z&0xF)<<4|(x&0xF)]; y-- {
				Y := byte(y >> 4)
				if c.Sections.Has(Y) {
					c.Sections.Get(Y).SkyLight.Set(x, y, z, 15)
				}
			}
			light := byte(15)
			for y := c.HeightMap[(z&0xF)<<4|(x&0xF)]; y > 0 && light > 0 && light < 16; y-- {
				Y := byte(y >> 4)
				if c.Sections.Has(Y) {
					section := c.Sections.Get(Y)
					section.SkyLight.Set(x, y, z, light)
					switch section.Blocks.Get(x, y, z) {
					case block.Air:
						// No change
					case block.Water, block.StationaryWater:
						light -= 2
					default:
						if !section.Blocks.Get(x, y, z).Passable() {
							light -= 5
						}
					}
				}
			}
		}
	}

	c.lightingDirty = false
	c.dirtyGeneric()
}

func (c *Chunk) Packet() []byte {
	c.lock.RLock()
	if !c.packetDirty && c.packet != nil {
		defer c.lock.RUnlock()
		return c.packet
	}
	c.lock.RUnlock()

	c.lock.Lock()
	if c.lightingDirty {
		c.processLighting()
	}
	defer c.lock.Unlock()
	if !c.packetDirty && c.packet != nil {
		return c.packet
	}
	var payload bytes.Buffer
	w := zlib.NewWriter(&payload)

	for _, s := range c.Sections {
		// I believe this is the fastest way to convert []SomeTypeThatIsAByte to []byte. Please correct me if I'm wrong.
		w.Write(reflect.ValueOf(s.Blocks[:]).Bytes())
	}
	for _, s := range c.Sections {
		w.Write(s.Data[:])
	}
	for _, s := range c.Sections {
		w.Write(s.BlockLight[:])
	}
	for _, s := range c.Sections {
		w.Write(s.SkyLight[:])
	}
	w.Write(reflect.ValueOf(c.Biomes[:]).Bytes())

	w.Close()

	var buf bytes.Buffer

	binary.Write(&buf, binary.BigEndian, uint8(0x33)) // Packet ID
	binary.Write(&buf, binary.BigEndian, c.X)         // Chunk X coordinate
	binary.Write(&buf, binary.BigEndian, c.Z)         // Chunk Z coordinate
	binary.Write(&buf, binary.BigEndian, uint8(1))    // Contains entire chunk (yes)
	have := uint16(0)
	for i := byte(0); i < 16; i++ {
		if c.Sections.Has(i) {
			have |= 1 << i
		}
	}
	binary.Write(&buf, binary.BigEndian, have)                 // Bitmask
	binary.Write(&buf, binary.BigEndian, uint16(0))            // Add bitmask (we don't use this)
	binary.Write(&buf, binary.BigEndian, int32(payload.Len())) // Payload length
	binary.Write(&buf, binary.BigEndian, int32(0))             // Add palyoad length (we don't use this)
	buf.Write(payload.Bytes())

	c.packet = buf.Bytes()
	c.packetDirty = false

	return c.packet
}

type Entity map[string]interface{}
type TileEntity map[string]interface{}

type TileTick struct {
	Type  uint32 `nbt:"i"`
	Ticks int32  `nbt:"t"`
	X     int32  `nbt:"x"`
	Y     int32  `nbt:"y"`
	Z     int32  `nbt:"z"`
}
