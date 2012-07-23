package protocol

import (
	"bytes"
	"compress/zlib"
	"sync"
)

type BlockType uint8

const (
	Air                      BlockType = 0
	Stone                    BlockType = 1
	Grass                    BlockType = 2
	Dirt                     BlockType = 3
	Cobblestone              BlockType = 4
	Planks                   BlockType = 5
	Sapling                  BlockType = 6
	Bedrock                  BlockType = 7
	Water                    BlockType = 8
	StationaryWater          BlockType = 9
	Lava                     BlockType = 10
	StationaryLava           BlockType = 11
	Sand                     BlockType = 12
	Gravel                   BlockType = 13
	GoldOre                  BlockType = 14
	IronOre                  BlockType = 15
	CoalOre                  BlockType = 16
	Log                      BlockType = 17
	Leaves                   BlockType = 18
	Sponge                   BlockType = 19
	Glass                    BlockType = 20
	LapisOre                 BlockType = 21
	LapisBlock               BlockType = 22
	Dispenser                BlockType = 23
	Sandstone                BlockType = 24
	NoteBlock                BlockType = 25
	Bed                      BlockType = 26
	POWERED_RAIL                       = 27
	DETECTOR_RAIL                      = 28
	PISTON_STICKY_BASE                 = 29
	WEB                                = 30
	LONG_GRASS                         = 31
	DEAD_BUSH                          = 32
	PISTON_BASE                        = 33
	PISTON_EXTENSION                   = 34
	Wool                     BlockType = 35
	PistonMovingPiece        BlockType = 36
	YELLOW_FLOWER                      = 37
	RED_ROSE                           = 38
	BROWN_MUSHROOM                     = 39
	RED_MUSHROOM                       = 40
	GOLD_BLOCK                         = 41
	IRON_BLOCK                         = 42
	DOUBLE_STEP                        = 43
	STEP                               = 44
	BRICK                              = 45
	TNT                                = 46
	BOOKSHELF                          = 47
	MOSSY_COBBLESTONE                  = 48
	OBSIDIAN                           = 49
	TORCH                              = 50
	FIRE                               = 51
	MOB_SPAWNER                        = 52
	WOOD_STAIRS                        = 53
	CHEST                              = 54
	REDSTONE_WIRE                      = 55
	DIAMOND_ORE                        = 56
	DIAMOND_BLOCK                      = 57
	WORKBENCH                          = 58
	CROPS                              = 59
	Farm                     BlockType = 60 // TODO: We need more of this
	FURNACE                            = 61
	BURNING_FURNACE                    = 62
	SIGN_POST                          = 63
	WOODEN_DOOR                        = 64
	LADDER                             = 65
	RAILS                              = 66
	COBBLESTONE_STAIRS                 = 67
	WALL_SIGN                          = 68
	LEVER                              = 69
	STONE_PLATE                        = 70
	IRON_DOOR_BLOCK                    = 71
	WOOD_PLATE                         = 72
	REDSTONE_ORE                       = 73
	GLOWING_REDSTONE_ORE               = 74
	REDSTONE_TORCH_OFF                 = 75
	REDSTONE_TORCH_ON                  = 76
	STONE_BUTTON                       = 77
	SNOW                               = 78
	ICE                                = 79
	SNOW_BLOCK                         = 80
	CACTUS                             = 81
	CLAY                               = 82
	SUGAR_CANE_BLOCK                   = 83
	JUKEBOX                            = 84
	FENCE                              = 85
	PUMPKIN                            = 86
	NETHERRACK                         = 87
	SOUL_SAND                          = 88
	GLOWSTONE                          = 89
	PORTAL                             = 90
	JACK_O_LANTERN                     = 91
	CAKE_BLOCK                         = 92
	DIODE_BLOCK_OFF                    = 93
	DIODE_BLOCK_ON                     = 94
	LOCKED_CHEST                       = 95
	TrapDoor                 BlockType = 96
	StoneBrickWithSilverfish BlockType = 97
	StoneBrick               BlockType = 98
	HugeMushroom1            BlockType = 99
	HugeMushroom2            BlockType = 100
	IronFence                BlockType = 101
	GlassPane                BlockType = 102
	Melon                    BlockType = 103
	PumpkinStem              BlockType = 104
	MelonStem                BlockType = 105
	Vines                    BlockType = 106
	FenceGate                BlockType = 107
	BrickStairs              BlockType = 108
	StoneStairs              BlockType = 109
	Mycelium                 BlockType = 110
	LilyPad                  BlockType = 111
	NetherBrick              BlockType = 112
	NetherFence              BlockType = 113
	NetherBrickStairs        BlockType = 114
	NetherWart               BlockType = 115
	EnchantingTable          BlockType = 116
	BrewingStand             BlockType = 117
	Cauldron                 BlockType = 118
	EndPortal                BlockType = 119
	EndPortalFrame           BlockType = 120
	EndStone                 BlockType = 121
	DragonEgg                BlockType = 122
	RedstoneLampOff          BlockType = 123
	RedstoneLampOn           BlockType = 124
)

type BlockSection [4096]BlockType

func (section *BlockSection) Set(x, y, z uint8, block BlockType) {
	section[uint32(y&15)<<8|uint32(z&15)<<4|uint32(x&15)] = block
}

func (section *BlockSection) Get(x, y, z uint8) BlockType {
	return section[uint32(y&15)<<8|uint32(z&15)<<4|uint32(x&15)]
}

type BlockChunk [16]BlockSection

func (chunk *BlockChunk) Set(x, y, z uint8, block BlockType) {
	chunk[y>>4].Set(x, y&0xF, z, block)
}

func (chunk *BlockChunk) Get(x, y, z uint8) BlockType {
	return chunk[y>>4].Get(x, y&0xF, z)
}

type NibbleSection [2048]uint8

func (section *NibbleSection) Set(x, y, z, nibble uint8) {
	if nibble > 0xF {
		panic("Illegal nibble value")
	}
	index := uint32(y)<<7 | uint32(z)<<3 | uint32(x)>>1
	if x&1 == 1 {
		section[index] = section[index]&0xF | nibble<<4
	} else {
		section[index] = section[index]&0xF0 | nibble
	}
}

func (section *NibbleSection) Get(x, y, z uint8) uint8 {
	index := uint32(y)<<7 | uint32(z)<<3 | uint32(x)>>1
	if x&1 == 1 {
		return section[index] >> 4
	}
	return section[index] & 0xF
}

type NibbleChunk [16]NibbleSection

func (chunk *NibbleChunk) Set(x, y, z, nibble uint8) {
	chunk[y>>4].Set(x, y&0xF, z, nibble)
}

func (chunk *NibbleChunk) Get(x, y, z uint8) uint8 {
	return chunk[y>>4].Get(x, y&0xF, z)
}

type Biome uint8

const (
	Ocean               Biome = 0
	Plains              Biome = 1
	Desert              Biome = 2
	ExtremeHills        Biome = 3
	Forest              Biome = 4
	Taiga               Biome = 5
	Swampland           Biome = 6
	River               Biome = 7
	Hell                Biome = 8
	Sky                 Biome = 9
	FrozenOcean         Biome = 10
	FrozenRiver         Biome = 11
	IcePlains           Biome = 12
	IceMountains        Biome = 13
	MushroomIsland      Biome = 14
	MushroomIslandShore Biome = 15
	Beach               Biome = 16
	DesertHills         Biome = 17
	ForestHills         Biome = 18
	TaigaHills          Biome = 19
	ExtremeHillsEdge    Biome = 20
	Jungle              Biome = 21
	JungleHills         Biome = 22
)

type Face uint8

const (
	FaceDown Face = iota
	FaceUp
	FaceWest
	FaceEast
	FaceSouth
	FaceNorth
)

type Chunk struct {
	Blocks      BlockChunk
	BlockData   NibbleChunk
	LightBlock  NibbleChunk
	LightSky    NibbleChunk
	Biomes      [16][16]Biome
	dirty       bool
	compressed  []byte
	compressing sync.Mutex
}

func (c *Chunk) SetBlock(x, y, z uint8, block BlockType) {
	c.Blocks.Set(x, y, z, block)
	c.dirty = true
}

func (c *Chunk) GetBlock(x, y, z uint8) BlockType {
	return c.Blocks.Get(x, y, z)
}

func (c *Chunk) SetBlockData(x, y, z, data uint8) {
	c.BlockData.Set(x, y, z, data)
	c.dirty = true
}

func (c *Chunk) GetBlockData(x, y, z uint8) uint8 {
	return c.BlockData.Get(x, y, z)
}

func (c *Chunk) Compressed() []byte {
	if c.dirty || c.compressed == nil {
		c.compressing.Lock()
		defer c.compressing.Unlock()
		if !c.dirty && c.compressed != nil {
			return c.compressed
		}
		var buf bytes.Buffer
		w := zlib.NewWriter(&buf)

		for _, blocks := range c.Blocks {
			for _, block := range blocks {
				w.Write([]byte{byte(block)})
			}
		}
		for _, data := range c.BlockData {
			w.Write(data[:])
		}
		for _, light := range c.LightBlock {
			w.Write(light[:])
		}
		for _, light := range c.LightSky {
			w.Write(light[:])
		}
		for _, biomes := range c.Biomes {
			for _, biome := range biomes {
				w.Write([]byte{byte(biome)})
			}
		}

		w.Close()
		c.compressed = buf.Bytes()
		c.dirty = false
	}
	return c.compressed
}
