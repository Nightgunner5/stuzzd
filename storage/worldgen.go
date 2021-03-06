package storage

import (
	"github.com/Nightgunner5/stuzzd/block"
	"github.com/Nightgunner5/stuzzd/chunk"
	"github.com/Nightgunner5/stuzzd/protocol"
	"github.com/Nightgunner5/stuzzd/util"
	"math/rand"
)

func river(in float64) float64 {
	// This function does some math that was decided upon by "trial and oh look at that, that's cool".

	out := 5 - 50*in*in
	if out < 0 {
		return 0
	}
	return out
}

func growTree(chunk *chunk.Chunk, r *rand.Rand, x, y, z int32) {
	if x&0xF == x && z&0xF == z {
		return
	}
	if x&0xF < 2 || x&0xF > 13 || z&0xF < 2 || z&0xF > 13 {
		return
	}

	chunk.SetBlock(x, y-1, z, block.Dirt)

	height := r.Int31n(5) + 3
	for Y := y; Y < y+height; Y++ {
		chunk.SetBlock(x, Y, z, block.Log)
	}

	for X := x - 2; X <= x+2; X++ {
		for Z := z - 2; Z <= z+2; Z++ {
			for Y := y + height - 2; Y <= y+height+2; Y++ {
				if chunk.GetBlock(X, Y, Z) == block.Air {
					chunk.SetBlock(X, Y, Z, block.Leaves)
				}
			}
		}
	}
}

func ChunkGen(chunkX, chunkZ int32) *chunk.Chunk {
	chunk := &chunk.Chunk{X: chunkX, Z: chunkZ}

	r := rand.New(rand.NewSource(int64(uint32(chunkX))<<32 | int64(uint32(chunkZ))))

	for x := chunkX << 4; x < (chunkX<<4)+16; x++ {
		for z := chunkZ << 4; z < (chunkZ<<4)+16; z++ {
			chunk.SetBlock(x, 0, z, block.Bedrock)

			fx := float64(x) / 16
			fz := float64(z) / 16

			stone := int32(6 + 4*util.Noise2(fx, fz))
			land := int32(58 + 8*util.Noise2(fx/10, fz/10))

			mountain := util.Noise3(fx/15, float64(land)/15, fz/15)
			isMountain := true
			mountain -= 0.4
			if mountain < 0 {
				mountain = 0
				isMountain = false
			}
			mountain *= 30

			land += int32(mountain)

			river := int32(river(util.Noise2(fx/4, fz/4)))

			smooth := util.Noise2(fx/50, fz/50) * 3

			rough := util.Noise2(fx*2, fz*2) * 2

			for y := int32(1); y < land-stone; y++ {
				chunk.SetBlock(x, y, z, block.Stone)
			}

			for y := land - stone; y < land; y++ {
				chunk.SetBlock(x, y, z, block.Dirt)
			}

			if river != 0 {
				chunk.SetBlock(x, 46-river, z, block.Gravel)
				chunk.SetBlock(x, 47-river, z, block.Gravel)
				chunk.SetBlock(x, 48-river, z, block.Gravel)
				chunk.SetBlock(x, 49-river, z, block.Gravel)
			}
			for y := 50 - river; y < 50; y++ {
				chunk.SetBlock(x, y, z, block.StationaryWater)
			}

			if river == 0 || land > 50 {
				chunk.SetBlock(x, land, z, block.Grass)
				if smooth < rough {
					if r.Intn(8) == 0 {
						fy := float64(land + 1)
						if util.Noise3(fx/10, fy/2, fz/10) > 0 {
							chunk.SetBlock(x, land+1, z, block.RedFlower)
						} else {
							chunk.SetBlock(x, land+1, z, block.YellowFlower)
						}
					} else {
						chunk.SetBlock(x, land+1, z, block.LongGrass)
						chunk.SetData(x, land+1, z, 1)
					}
				}
				if smooth < rough/5 && mountain > 1 && r.Intn(20) == 0 {
					growTree(chunk, r, x, land+1, z)
				}
				if isMountain {
					chunk.SetBiome(x, z, protocol.ExtremeHills)
				} else {
					chunk.SetBiome(x, z, protocol.Plains)
				}
			} else {
				chunk.SetBiome(x, z, protocol.River)
			}

			if river != 0 {
				chunk.SetBlock(x, 50, z, block.Air)

				for y := int32(51); y < 64 && r.Intn(20) != 0 && chunk.GetBlock(x, y, z) == block.Dirt; y++ {
					chunk.SetBlock(x, y, z, block.Stone)
				}

				for y := int32(51); y < 64 && r.Intn(4) != 0 && chunk.GetBlock(x, y, z) == block.Stone; y++ {
					chunk.SetBlock(x, y, z, block.Air)
				}

			}
		}
	}
	chunk.InitLighting()

	return chunk
}
