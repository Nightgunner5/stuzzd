package block

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
	PoweredRail              BlockType = 27
	DetectorRail             BlockType = 28
	PistonBaseSticky         BlockType = 29
	SpiderWeb                BlockType = 30
	LongGrass                BlockType = 31
	DeadBush                 BlockType = 32
	PistonBase               BlockType = 33
	PistonExtension          BlockType = 34
	Wool                     BlockType = 35
	PistonMovingPiece        BlockType = 36
	YellowFlower             BlockType = 37
	RedFlower                BlockType = 38
	BrownMushroom            BlockType = 39
	RedMushroom              BlockType = 40
	GoldBlock                BlockType = 41
	IronBlock                BlockType = 42
	DoubleStep               BlockType = 43
	HalfStep                 BlockType = 44
	Bricks                   BlockType = 45
	TNT                      BlockType = 46
	Bookshelf                BlockType = 47
	MossyCobblestone         BlockType = 48
	Obsidian                 BlockType = 49
	Torch                    BlockType = 50
	Fire                     BlockType = 51
	MobSpawner               BlockType = 52
	WoodStairs               BlockType = 53
	Chest                    BlockType = 54
	RedstoneWire             BlockType = 55
	DiamondOre               BlockType = 56
	DiamondBlock             BlockType = 57
	CraftingTable            BlockType = 58
	Wheat                    BlockType = 59
	Farm                     BlockType = 60 // TODO: We need more of this
	Furnace                  BlockType = 61
	FurnaceBurning           BlockType = 62
	SignPost                 BlockType = 63
	WoodenDoor               BlockType = 64
	Ladder                   BlockType = 65
	Rails                    BlockType = 66
	CobblestoneStairs        BlockType = 67
	WallSign                 BlockType = 68
	Lever                    BlockType = 69
	StonePressurePlate       BlockType = 70
	IronDoor                 BlockType = 71
	WoodPressurePlate        BlockType = 72
	RedstoneOre              BlockType = 73
	RedstoneOreGlowing       BlockType = 74
	RedstoneTorchOff         BlockType = 75
	RedstoneTorchOn          BlockType = 76
	Button                   BlockType = 77
	Snow                     BlockType = 78
	Ice                      BlockType = 79
	SnowBlock                BlockType = 80
	Cactus                   BlockType = 81
	Clay                     BlockType = 82
	SugarCane                BlockType = 83
	Jukebox                  BlockType = 84
	Fence                    BlockType = 85
	Pumpkin                  BlockType = 86
	Netherrack               BlockType = 87
	SoulSand                 BlockType = 88
	Glowstone                BlockType = 89
	NetherPortal             BlockType = 90
	JackOLantern             BlockType = 91
	Cake                     BlockType = 92
	RedstoneRepeaterOff      BlockType = 93
	RedstoneRepeaterOn       BlockType = 94
	SteveCoChest             BlockType = 95
	Trapdoor                 BlockType = 96
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
	NetherStairs             BlockType = 114
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

func (b BlockType) Passable() bool {
	switch b {
	case Air, Water, StationaryWater, Lava, StationaryLava, Sapling, LongGrass,
		RedFlower, YellowFlower, Torch, RedstoneTorchOn, RedstoneTorchOff,
		RedstoneRepeaterOn, RedstoneRepeaterOff, SugarCane, LilyPad, NetherWart,
		Wheat, MelonStem, PumpkinStem, NetherPortal, WoodPressurePlate,
		StonePressurePlate, Button, Lever, RedstoneWire, EndPortal,
		RedMushroom, BrownMushroom, Ladder, DeadBush, Fire, Vines, Snow,
		Rails, PoweredRail, DetectorRail:
		return true
	}
	return false
}

func (b BlockType) SemiPassable() bool {
	switch b {
	case WoodStairs, CobblestoneStairs, StoneStairs, BrickStairs, NetherStairs,
		Cauldron, BrewingStand, EnchantingTable, Cake, IronFence, GlassPane,
		Fence, FenceGate, NetherFence, Bed, HalfStep, WoodenDoor, IronDoor,
		SpiderWeb, WallSign, SignPost:
		return true
	}
	return false
}

func (b BlockType) ItemDrop() int16 {
	switch b {
	case Grass:
		return int16(Dirt)
	}
	return int16(b)
}
