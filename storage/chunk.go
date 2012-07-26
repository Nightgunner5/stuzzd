package storage

import "github.com/Nightgunner5/stuzzd/protocol"

type Chunk struct {
	Populated    bool
	X            int32 `nbt:"xPos"`
	Z            int32 `nbt:"yPos"`
	LastUpdate   uint64
	Biomes       []protocol.Biome
	Entities     []Entity
	Sections     []Section
	TileEntities []TileEntity
	HeightMap    [16 * 16]int32
}

type Section struct {
	Y          byte
	BlockLight protocol.NibbleSection
	Blocks     protocol.BlockSection
	Data       protocol.NibbleSection
	SkyLight   protocol.NibbleSection
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
