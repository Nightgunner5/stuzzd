package storage

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/Nightgunner5/go.nbt"
	"os"
	"sync"
)

var lock sync.Mutex

func ReadChunk(chunkX, chunkZ int32) (*Chunk, error) {
	lock.Lock()
	defer lock.Unlock()
	regionX, regionZ := chunkX>>5, chunkZ>>5
	chunkX, chunkZ = chunkX&0x1F, chunkZ&0x1F

	f, err := os.Open(fmt.Sprintf("world/region/r.%d.%d.mca", regionX, regionZ))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	f.Seek(int64(chunkZ<<5|chunkX)<<2, os.SEEK_SET)

	var location uint32
	binary.Read(f, binary.BigEndian, &location)

	offset := location >> 8
	// sectorCount := location & 0xFF

	if offset == 0 {
		return nil, fmt.Errorf("Chunk (%d, %d) in region (%d, %d) does not exist.", chunkX, chunkZ, regionX, regionZ)
	}

	f.Seek(int64(offset)<<12, os.SEEK_SET)
	var length uint32
	binary.Read(f, binary.BigEndian, &length)
	length--

	var compression nbt.Compression
	binary.Read(f, binary.BigEndian, &compression)

	buf := make([]byte, length)
	_, err = f.Read(buf)
	if err != nil {
		return nil, err
	}

	var chunk ChunkHolder

	err = nbt.Unmarshal(compression, bytes.NewReader(buf), &chunk)

	return &chunk.Level, err
}
