package chunk

import (
	"encoding/binary"
	"github.com/Nightgunner5/stuzzd/player"
	"io"
)

type ItemDrop Entity

func (e Entity) ItemDrop() ItemDrop {
	if e.Type() == "Item" {
		return ItemDrop(e)
	}
	return nil
}

func NewItemDrop(x, y, z float64, item *player.InventoryItem) ItemDrop {
	return ItemDrop{
		"id":           "Item",
		"Pos":          []float64{x, y, z},
		"Motion":       []float64{0, 0, 0},
		"Rotation":     []float32{0, 0},
		"FallDistance": float32(0),
		"Fire":         int16(0),
		"Air":          int16(0),
		"OnGround":     int8(1),
		"Health":       int16(5),
		"Age":          int16(0),
		"Item": map[string]interface{}{
			"id":     item.Type,
			"Damage": item.Damage,
			"Count":  item.Count,
			"tag":    item.Meta,
		},
	}
}

func (i ItemDrop) Item() int16 {
	return i["Item"].(map[string]interface{})["id"].(int16)
}

func (i ItemDrop) Count() int8 {
	return i["Item"].(map[string]interface{})["Count"].(int8)
}

func (i ItemDrop) Damage() int16 {
	return i["Item"].(map[string]interface{})["Damage"].(int16)
}

func (i ItemDrop) SpawnPacket(w io.Writer) {
	binary.Write(w, binary.BigEndian, uint8(0x15))
	binary.Write(w, binary.BigEndian, Entity(i).ID())
	binary.Write(w, binary.BigEndian, i.Item())
	binary.Write(w, binary.BigEndian, i.Count())
	binary.Write(w, binary.BigEndian, i.Damage())
	for j := 0; j < 3; j++ {
		binary.Write(w, binary.BigEndian, int32(i["Pos"].([]float64)[j]*32))
	}
	binary.Write(w, binary.BigEndian, uint8(0)) // Yaw
	binary.Write(w, binary.BigEndian, uint8(0)) // Pitch
	binary.Write(w, binary.BigEndian, uint8(0)) // Roll
}
