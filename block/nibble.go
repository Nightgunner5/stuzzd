package block

type NibbleSection [2048]uint8

func (section *NibbleSection) Set(x, y, z int32, nibble uint8) {
	if nibble > 0xF {
		panic("Illegal nibble value")
	}
	index := uint32(y&0xF)<<7 | uint32(z&0xF)<<3 | uint32(x&0xF)>>1
	if x&1 == 1 {
		section[index] = section[index]&0xF | nibble<<4
	} else {
		section[index] = section[index]&0xF0 | nibble
	}
}

func (section *NibbleSection) Get(x, y, z int32) uint8 {
	index := uint32(y&0xF)<<7 | uint32(z&0xF)<<3 | uint32(x&0xF)>>1
	if x&1 == 1 {
		return section[index] >> 4
	}
	return section[index] & 0xF
}

type NibbleChunk [16]NibbleSection

func (chunk *NibbleChunk) Set(x, y, z int32, nibble uint8) {
	chunk[y>>4].Set(x, y, z, nibble)
}

func (chunk *NibbleChunk) Get(x, y, z int32) uint8 {
	return chunk[y>>4].Get(x, y, z)
}
