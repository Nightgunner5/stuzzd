package storage

import "github.com/Nightgunner5/stuzzd/protocol"

type ChunkHolder struct {
	Level *Chunk
}

type Chunk struct {
	TerrainPopulated byte
	X            int32 `nbt:"xPos"`
	Z            int32 `nbt:"zPos"`
	LastUpdate   uint64
	Biomes       []byte
	Entities     []Entity
	Sections     []Section
	TileEntities []TileEntity
	TileTicks    []TileTick
	HeightMap    [16 * 16]int32
}

type Section struct {
	Y          byte
	BlockLight []byte
	Blocks     []byte
	Data       []byte
	SkyLight   []byte
}

type Entity struct {
	OnGround     bool
	Air          uint16
	AttackTime   uint16
	DeathTime    uint16
	Fire         uint16
	Health       uint16
	HurtTime     uint16
	FallDistance float32
	Type         string `nbt:"id"`
	Motion       [3]float64
	Pos          [3]float64
	Rotation     [2]float32
}

type TileEntity map[string]interface{}

type TileTick struct {
	I, T, X, Y, Z int32
}
