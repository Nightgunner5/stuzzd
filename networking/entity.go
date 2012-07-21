package networking

import "sync"

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

func RegisterEntity(ent Entity) {
	entities[ent.ID()] = ent
}
