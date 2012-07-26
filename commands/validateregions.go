package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	dir, err := os.Open(os.Args[0])
	if err != nil {
		log.Fatal(err)
	}

	files, err := dir.Readdir(-1)
	if err != nil {
		log.Fatal(err)
	}
	dir.Close()

	for _, file := range files {
		if strings.HasPrefix(file.Name(), "r.") && strings.HasSuffix(file.Name(), ".mca") {
			sectors := int32(file.Size() >> 12)
			accountedFor := make([]string, sectors)
			accountedFor[0], accountedFor[1] = "location header", "timestamp header"
			f, _ := os.Open(os.Args[0] + "/" + file.Name())
			for z := 0; z < 32; z++ {
				for x := 0; x < 32; x++ {
					var location int32
					binary.Read(f, binary.BigEndian, &location)

					offset := location >> 8
					sectorCount := int32(int8(location & 0xFF)) // Get the sign bit to be a sign bit
					if offset == 0 && sectorCount == 0 {
						// No chunk here yet.
					} else if sectorCount <= 0 {
						log.Printf("In %s: Chunk(%d, %d) has zero or a negative number of sectors (%d)", file.Name(), x, z, sectorCount)
					} else if offset < 2 {
						log.Printf("In %s: Chunk(%d, %d) has an invalid offset (too low: %d)", file.Name(), x, z, offset)
					} else if offset+sectorCount > sectors {
						log.Printf("In %s: Chunk(%d, %d) has an invalid offset or sector count (goes past end of file: %d, %d)", file.Name(), x, z, offset, sectorCount)
					} else {
						chunkName := fmt.Sprintf("Chunk(%d, %d)", x, z)
						for i := offset; i < offset+sectorCount; i++ {
							if accountedFor[i] != "" {
								log.Printf("In %s: %s claims offset %d, but the offset is already owned by %s", file.Name(), chunkName, i, accountedFor[i])
							} else {
								accountedFor[i] = chunkName
							}
						}
					}
				}
			}
			f.Close()
		}
	}
}
