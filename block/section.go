package block

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
