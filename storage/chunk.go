package storage

type ChunkHolder struct {
	Level Chunk
}

type Chunk struct {
	TerrainPopulated byte
	X                int32 `nbt:"xPos"`
	Z                int32 `nbt:"zPos"`
	LastUpdate       uint64
	Biomes           []byte
	Entities         []Entity
	Sections         []Section
	TileEntities     []TileEntity
	TileTicks        []TileTick
	HeightMap        [16 * 16]int32
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
	Motion       []float64
	Pos          []float64
	Rotation     []float32
}

type TileEntity map[string]interface{}

type TileTick struct {
	Type  int32 `nbt:"i"`
	Ticks int32 `nbt:"t"`
	X     int32 `nbt:"x"`
	Y     int32 `nbt:"y"`
	Z     int32 `nbt:"z"`
}
