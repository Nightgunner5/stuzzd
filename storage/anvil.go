package storage

import (
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"github.com/bemasher/GoNBT"
	"os"
)

func ReadChunk(chunkX, chunkZ int32) (*Chunk, error) {
	regionX, regionZ := chunkX>>5, chunkZ>>5
	chunkX, chunkZ = chunkX&0x1F, chunkZ&0x1F

	region, err := os.Open(fmt.Sprintf("world/region/r.%d.%d.mca", regionX, regionZ))
	if err != nil {
		return nil, err
	}
	defer region.Close()

	region.Seek(int64(chunkZ<<5|chunkX)<<4, os.SEEK_SET)

	var location uint32
	binary.Read(region, binary.BigEndian, &location)

	offset := location >> 8
	//sectorCount := location & 0xFF

	if offset == 0 {
		return nil, fmt.Errorf("Chunk (%d, %d) in region (%d, %d) does not exist.", chunkX, chunkZ, regionX, regionZ)
	}

	region.Seek(int64(offset<<12+4 /* length */ +1 /* compression type (always zlib) */), os.SEEK_SET)

	r, err := zlib.NewReader(region)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var chunk ChunkHolder
	nbt.Read(r, &chunk)

	return &chunk.Level, nil
}
