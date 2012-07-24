package storage

import (
	"compress/gzip"
	"github.com/Nightgunner5/stuzzd/protocol"
	"github.com/bemasher/GoNBT"
	"os"
	"sync"
)

var players = make(map[string]*Player)

var playerLock sync.RWMutex

func GetPlayer(name string) *Player {
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

func loadPlayer(name string) *Player {
	player := new(Player)

	f, err := os.Open("world/players/" + name + ".dat")
	if err != nil {
		player.defaults()
		return player
	}
	gz, _ := gzip.NewReader(f)
	defer gz.Close()

	nbt.Read(gz, &player)
	return player
}

type Player struct {
	OnGround       bool
	Sleeping       bool
	Air            int16
	AttackTime     int16
	DeathTime      int16
	Fire           int16
	Health         int16
	HurtTime       int16
	Dimension      protocol.Dimension
	FoodLevel      int32               `nbt:"foodLevel"`
	FoodTickTimer  int32               `nbt:"foodTickTimer"`
	GameMode       protocol.ServerMode `nbt:"playerGameType"`
	XpLevel        int32
	XpTotal        int32
	FallDistance   float32
	FoodExhaustion float32 `nbt:"foodExhastionLevel"`
	FoodSaturation float32 `nbt:"foodSaturationLevel"`
	XpP            float32
	Inventory      map[string]*InventoryItem
	Motion         []float64
	Position       []float64
	Rotation       []float32
}

func (p *Player) defaults() {
	p.Motion = []float64{0, 0, 0}
	p.Position = []float64{8.5, 70, 8.5}
	p.Rotation = []float32{0, 0}
}

func (p *Player) Save() {
	// MASSIVE TODO
}

type InventoryItem struct {
	Count  int8
	Slot   int8
	Damage int16
	ID     int16 `nbt:"id"`
}
