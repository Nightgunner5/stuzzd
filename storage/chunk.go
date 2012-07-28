package storage

import "github.com/Nightgunner5/stuzzd/protocol"

type ChunkHolder struct {
	Level Chunk
}

type Chunk struct {
	X int32 `nbt:"xPos"`
	Z int32 `nbt:"zPos"`

	TerrainPopulated bool

	LastUpdate   uint64
	Sections     []Section
	Entities     []Entity
	TileEntities []TileEntity
	TileTicks    []TileTick
	Biomes       [256]protocol.Biome
	HeightMap    [256]int32
}

type Section struct {
	Y byte

	Blocks protocol.BlockSection
	Data   protocol.NibbleSection

	SkyLight   protocol.NibbleSection
	BlockLight protocol.NibbleSection
}

type Entity map[string]interface{}

type TileEntity struct {
}

type TileTick struct {
	Type  uint32 `nbt:"i"`
	Ticks int32  `nbt:"t"`
	X     int32  `nbt:"x"`
	Y     int32  `nbt:"y"`
	Z     int32  `nbt:"z"`
}
