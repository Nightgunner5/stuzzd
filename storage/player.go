package storage

import (
	"github.com/Nightgunner5/go.nbt"
	"github.com/Nightgunner5/stuzzd/player"
	"os"
	"sync"
)

var players = make(map[string]*player.Player)

var playerLock sync.RWMutex

func GetPlayer(name string) *player.Player {
	playerLock.RLock()
	if player, ok := players[name]; ok {
		playerLock.RUnlock()
		return player
	}
	playerLock.RUnlock()

	playerLock.Lock()
	defer playerLock.Unlock()

	players[name] = loadPlayer(name)
	return players[name]
}

func loadPlayer(name string) *player.Player {
	player := new(player.Player)

	f, err := os.Open("world/players/" + name + ".dat")
	if err != nil {
		spawnChunk := GetChunk(0, 0)
		player.Position = []float64{8.5, float64(spawnChunk.GetHighestBlockYAt(8, 8)) + 1, 8.5}
		player.Motion = []float64{0, 0, 0}
		player.Rotation = []float32{0, 0}
		ReleaseChunk(0, 0)
		return player
	}
	defer f.Close()

	err = nbt.Unmarshal(nbt.GZip, f, player)
	if err != nil {
		panic(err)
	}
	return player
}

func SavePlayer(name string, player *player.Player) error {
	f, err := os.Create("world/players/" + name + ".dat")
	if err != nil {
		return err
	}
	defer f.Close()

	err = nbt.Marshal(nbt.GZip, f, player)

	return err
}
