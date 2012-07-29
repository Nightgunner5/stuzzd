package block

type BlockSection [4096]BlockType

func (section *BlockSection) Set(x, y, z int32, block BlockType) {
	section[uint32(y&0xF)<<8|uint32(z&0xF)<<4|uint32(x&0xF)] = block
}

func (section *BlockSection) Get(x, y, z int32) BlockType {
	return section[uint32(y&0xF)<<8|uint32(z&0xF)<<4|uint32(x&0xF)]
}

type BlockChunk [16]BlockSection

func (chunk *BlockChunk) Set(x, y, z int32, block BlockType) {
	chunk[y>>4].Set(x, y, z, block)
}

func (chunk *BlockChunk) Get(x, y, z int32) BlockType {
	return chunk[y>>4].Get(x, y, z)
}
