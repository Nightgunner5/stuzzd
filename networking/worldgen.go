package networking

import (
	"github.com/Nightgunner5/stuzzd/protocol"
	"github.com/Nightgunner5/stuzzd/util"
	"math/rand"
)

func river(in float64) float64 {
	out := 5 - 50*in*in
	if out < 0 {
		return 0
	}
	return out
}

func ChunkGen(chunkX, chunkZ int32) *Chunk {
	chunk := new(Chunk)

	r := rand.New(rand.NewSource(int64(uint32(chunkX))<<32 | int64(uint32(chunkZ))))

	for x := uint8(0); x < 16; x++ {
		for z := uint8(0); z < 16; z++ {
			chunk.SetBlock(x, 0, z, protocol.Bedrock)

			fx := float64(x)/16 + float64(chunkX)
			fz := float64(z)/16 + float64(chunkZ)

			stone := uint8(6 + 4*util.Noise2(fx, fz))
			land := uint8(58 + 8*util.Noise2(fx/10, fz/10))

			mountain := util.Noise3(fx/15, float64(land)/15, fz/15)
			mountain -= 0.4
			if mountain < 0 {
				mountain = 0
			}
			mountain *= 30

			land += uint8(mountain)

			river := uint8(river(util.Noise2(fx/4, fz/4)))

			for y := uint8(1); y < land-stone; y++ {
				chunk.SetBlock(x, y, z, protocol.Stone)
			}

			for y := land - stone; y < land; y++ {
				chunk.SetBlock(x, y, z, protocol.Dirt)
			}

			// Begin river
			if river != 0 {
				chunk.SetBlock(x, 46-river, z, protocol.Gravel)
				chunk.SetBlock(x, 47-river, z, protocol.Gravel)
				chunk.SetBlock(x, 48-river, z, protocol.Gravel)
				chunk.SetBlock(x, 49-river, z, protocol.Gravel)
			}
			for y := 50 - river; y < 50; y++ {
				chunk.SetBlock(x, y, z, protocol.StationaryWater)
			}

			if river == 0 || land > 50 {
				chunk.SetBlock(x, land, z, protocol.Grass)
				if r.Intn(3) == 0 {
					if r.Intn(8) == 0 {
						fy := float64(land + 1)
						if util.Noise3(fx/2, fy/2, fz/2) > 0 {
							chunk.SetBlock(x, land+1, z, protocol.RedFlower)
						} else {
							chunk.SetBlock(x, land+1, z, protocol.YellowFlower)
						}
					} else {
						chunk.SetBlock(x, land+1, z, protocol.LongGrass)
						chunk.SetBlockData(x, land+1, z, 1)
					}
				}
				chunk.SetBiome(x, z, protocol.Plains)
			} else {
				chunk.SetBiome(x, z, protocol.River)
			}

			if river != 0 {
				chunk.SetBlock(x, 50, z, protocol.Air)

				for y := uint8(51); y < 64 && r.Intn(20) != 0 && chunk.GetBlock(x, y, z) == protocol.Dirt; y++ {
					chunk.SetBlock(x, y, z, protocol.Stone)
				}

				for y := uint8(51); y < 64 && r.Intn(4) != 0 && chunk.GetBlock(x, y, z) == protocol.Stone; y++ {
					chunk.SetBlock(x, y, z, protocol.Air)
				}

			}
			// End river
		}
	}
	chunk.InitLighting()

	return chunk
}
