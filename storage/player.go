package storage

import (
	"github.com/Nightgunner5/go.nbt"
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
		player.name = name
		player.defaults()
		return player
	}
	defer f.Close()

	err = nbt.Unmarshal(nbt.GZip, f, player)
	if err != nil {
		panic(err)
	}
	player.name = name
	return player
}

type Player struct {
	FoodSaturationLevel float32 `nbt:"foodSaturationLevel"`
	FoodExhaustionLevel float32 `nbt:"foodExhaustionLevel"`
	FoodTickTimer       uint32  `nbt:"foodTickTimer"`
	FoodLevel           uint32  `nbt:"foodLevel"`

	XpLevel uint32
	XpP     float32
	XpTotal uint32

	Health uint16
	Fire   uint16
	Air    uint16

	AttackTime uint16
	DeathTime  uint16
	HurtTime   uint16

	FallDistance float32

	Sleeping   bool
	SleepTimer uint16

	SpawnX int32
	SpawnY int32
	SpawnZ int32

	OnGround  bool
	Position  []float64 `nbt:"Pos"`
	Motion    []float64
	Rotation  []float32
	Dimension int32

	GameType  uint32          `nbt:"playerGameType"`
	Abilities PlayerAbilities `nbt:"abilities"`

	Inventory  []InventoryItem
	EnderItems []InventoryItem

	name string `nbt:"-"`
}

type PlayerAbilities struct {
	MayFly bool `nbt:"mayfly"`
	Flying bool `nbt:"flying"`

	FlySpeed  float32 `nbt:"flySpeed"`
	WalkSpeed float32 `nbt:"walkSpeed"`

	InstaBuild   bool `nbt:"instabuild"`
	Invulnerable bool `nbt:"invulnerable"`
	MayBuild     bool `nbt:"mayBuild"`
}

type InventoryItem struct {
	Type   uint16 `nbt:"id"`
	Damage uint16
	Count  uint8
	Slot   uint8
}

func (p *Player) defaults() {
	p.Motion = []float64{0, 0, 0}
	p.Position = []float64{8.5, 60, 8.5}
	p.Rotation = []float32{0, 0}
}

func (p *Player) Save() error {
	f, err := os.Create("world/players/" + p.name + ".dat")
	if err != nil {
		return err
	}
	defer f.Close()

	err = nbt.Marshal(nbt.GZip, f, p)

	return err
}
