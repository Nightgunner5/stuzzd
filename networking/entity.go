package networking

import (
	"bytes"
	"github.com/Nightgunner5/stuzzd/protocol"
	"io"
	"sync"
)

type Entity interface {
	ID() int32
	SpawnPacket(io.Writer)
	Position() (x, y, z float64)
	SetPosition(x, y, z float64)
}

var idMutex sync.Mutex
var nextID int32

func assignID() int32 {
	idMutex.Lock()
	defer idMutex.Unlock()

	nextID++
	return nextID
}

var entities = make(map[int32]Entity)
var players = make(map[int32]Player)

func RegisterEntity(ent Entity) {
	entities[ent.ID()] = ent
	if p, ok := ent.(Player); ok {
		players[ent.ID()] = p
	}
}

func RemoveEntity(ent Entity) {
	delete(entities, ent.ID())
	if _, ok := ent.(Player); ok {
		delete(players, ent.ID())
	}
	SendToAll(protocol.DestroyEntity{ID: ent.ID()})
}

func EntitySpawnPacket(ent Entity) protocol.Packet {
	var buf bytes.Buffer
	ent.SpawnPacket(&buf)
	return protocol.BakedPacket(buf.Bytes())
}
