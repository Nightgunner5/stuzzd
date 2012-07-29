package storage

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/Nightgunner5/go.nbt"
	"os"
	"sync"
)

var locks map[uint64]*sync.RWMutex
var lockLock sync.Mutex
func getLock(chunkX, chunkZ int32) *sync.RWMutex {
	lockLock.Lock()
	defer lockLock.Unlock()
	regionID := uint64(uint32(chunkX>>5))<<32 | uint64(uint32(chunkZ>>5))

	if locks[regionID] == nil {
		locks[regionID] = new(sync.RWMutex)
	}

	return locks[regionID]
}

func ReadChunk(chunkX, chunkZ int32) (*Chunk, error) {
	lock := getLock(chunkX, chunkZ)
	lock.RLock()
	defer lock.RUnlock()

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

func WriteChunk(chunk *Chunk) error {
	lock := getLock(chunk.X, chunk.Z)
	lock.Lock()
	defer lock.Unlock()

	regionX, regionZ := chunk.X>>5, chunk.Z>>5
	chunkX, chunkZ := chunk.X&0x1F, chunk.Z&0x1F

	f, err := os.OpenFile(fmt.Sprintf("world/region/r.%d.%d.mca", regionX, regionZ), os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	fstat, err := f.Stat()
	if err != nil {
		return err
	}
	size := fstat.Size()
	if size%4096 != 0 {
		needed := 4096 - (size % 4096)
		f.Write(make([]byte, needed))
		size += needed
	}

	sectors := make([]bool, size>>12)
	if len(sectors) < 2 {
		sectors = make([]bool, 2)
		f.Write(make([]byte, 8192))
	} else {
		sectors[0] = true
		sectors[1] = true

		for i := 0; i < 1024; i++ {
			var location uint32
			binary.Read(f, binary.BigEndian, &location)

			firstSector := location >> 8
			numSectors := location & 0xFF

			for sector := firstSector; sector < firstSector+numSectors; sector++ {
				sectors[sector] = true
			}
		}

		f.Seek(int64(chunkZ<<5|chunkX)<<2, os.SEEK_SET)
		var location uint32
		binary.Read(f, binary.BigEndian, &location)
		firstSector := location >> 8
		numSectors := location & 0xFF
		for sector := firstSector; sector < firstSector+numSectors; sector++ {
			sectors[sector] = false
		}
	}

	var buf bytes.Buffer
	err = nbt.Marshal(nbt.ZLib, &buf, ChunkHolder{*chunk})
	if err != nil {
		return err
	}

	encoded := append([]byte{0, 0, 0, 0, byte(nbt.ZLib)}, buf.Bytes()...)

	binary.BigEndian.PutUint32(encoded[:4], uint32(len(encoded)-4))

	numSectors := len(encoded)/4096 + 1

search:
	for firstSector := 2; firstSector+numSectors < len(sectors); firstSector++ {
		if sectors[firstSector] {
			continue search
		}

		for availableSectors := 1; availableSectors < numSectors; availableSectors++ {
			if sectors[firstSector+availableSectors] {
				firstSector += availableSectors
				continue search
			}
		}

		f.Seek(int64(chunkZ<<5|chunkX)<<2, os.SEEK_SET)
		binary.Write(f, binary.BigEndian, uint32(firstSector<<8)|uint32(numSectors))
		f.Seek(int64((chunkZ<<5|chunkX)<<2)+4096, os.SEEK_SET)
		binary.Write(f, binary.BigEndian, uint32(chunk.LastUpdate))

		f.Seek(int64(firstSector)<<12, os.SEEK_SET)
		_, err = f.Write(encoded)

		return err
	}

	f.Seek(int64(chunkZ<<5|chunkX)<<2, os.SEEK_SET)
	binary.Write(f, binary.BigEndian, uint32(len(sectors)<<8)|uint32(numSectors))
	f.Seek(int64((chunkZ<<5|chunkX)<<2)+4096, os.SEEK_SET)
	binary.Write(f, binary.BigEndian, uint32(chunk.LastUpdate))

	f.Seek(0, os.SEEK_END)
	_, err = f.Write(encoded)

	// File length is fixed on the next write.

	return err
}
