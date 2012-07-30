package player

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

func (abilities *PlayerAbilities) Packet() []byte {
	buf := []byte{0xCA, 0, 0, 0, 0}

	if abilities.Invulnerable {
		buf[1] = 1
	}
	if abilities.Flying {
		buf[2] = 1
	}
	if abilities.MayFly {
		buf[3] = 1
	}
	if abilities.InstaBuild {
		buf[4] = 1
	}

	return buf
}

type InventoryItem struct {
	Type   int16 `nbt:"id"`
	Damage int16
	Count  int8
	Slot   int8
	Meta   map[string]interface{} `nbt:"tag"`
}
