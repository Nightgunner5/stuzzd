package protocol

import (
	"encoding/binary"
	"github.com/Nightgunner5/go.nbt"
	"io"
)

func ReadSlot(in io.Reader) (id int16, count int8, damage int16, meta map[string]interface{}) {
	binary.Read(in, binary.BigEndian, &id)
	if id == -1 {
		return
	}

	binary.Read(in, binary.BigEndian, &count)
	binary.Read(in, binary.BigEndian, &damage)

	if 256 <= id && id <= 259 || 267 <= id && id <= 279 || 283 <= id && id <= 286 || 290 <= id && id <= 294 || 298 <= id && id <= 317 || id == 261 || id == 359 || id == 346 {
		var more int8
		binary.Read(in, binary.BigEndian, &more)
		if more > -1 {
			nbt.Unmarshal(nbt.GZip, &io.LimitedReader{R: in, N: int64(more)}, &meta)
		}
	}
	return
}
