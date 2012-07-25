package networking

import (
	"github.com/Nightgunner5/stuzzd/protocol"
	"sync"
)

type Entity interface {
	ID() uint32
}

type entity struct {
	id uint32
}

var idMutex sync.Mutex
var nextID uint32

func assignID() uint32 {
	idMutex.Lock()
	defer idMutex.Unlock()

	nextID++
	return nextID
}

func (e *entity) ID() uint32 {
	return e.id
}

var entities = make(map[uint32]Entity)
var players = make(map[uint32]Player)

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
