package storage

import (
	"fmt"
	"github.com/jteeuwen/nbt"
)

func FromNBT(in nbt.Tag) interface{} {
	if in == nil {
		return nil
	}
	switch tag := in.(type) {
	case *nbt.CompoundTag:
		out := make(map[string]interface{})
		for _, item := range tag.Items {
			out[item.Name()] = FromNBT(item)
		}
		return out

	case *nbt.ListTag:
		out := make([]interface{}, 0, len(tag.Items))
		for _, item := range tag.Items {
			out = append(out, FromNBT(item))
		}
		return out

	case *nbt.ByteTag:
		return tag.Value

	case *nbt.ShortTag:
		return tag.Value

	case *nbt.IntTag:
		return tag.Value

	case *nbt.LongTag:
		return tag.Value

	case *nbt.FloatTag:
		return tag.Value

	case *nbt.DoubleTag:
		return tag.Value

	case *nbt.StringTag:
		return tag.Value

	case *nbt.ByteArrayTag:
		// I have yet to see a case where signed bytes in an array are useful.
		out := make([]byte, 0, len(tag.Value))
		for _, b := range tag.Value {
			out = append(out, byte(b))
		}
		return out

	case *nbt.IntArrayTag:
		return tag.Value
	}

	panic(fmt.Sprintf("Unknown tag type: %T!", in))
}
