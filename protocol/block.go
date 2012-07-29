package protocol

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
