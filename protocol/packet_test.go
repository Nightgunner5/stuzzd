package protocol

import (
	"bytes"
	"reflect"
	"testing"
)

func TestString(t *testing.T) {
	t.Parallel()

	unencoded := "StuzzHosting is BEST hosting"
	encoded := []byte{0, 28, 0, 83, 0, 116, 0, 117, 0, 122, 0, 122, 0, 72, 0, 111, 0, 115, 0, 116, 0, 105, 0, 110, 0, 103, 0, 32, 0, 105, 0, 115, 0, 32, 0, 66, 0, 69, 0, 83, 0, 84, 0, 32, 0, 104, 0, 111, 0, 115, 0, 116, 0, 105, 0, 110, 0, 103}

	b := stringToBytes(unencoded)
	if !reflect.DeepEqual(b, encoded) {
		t.Log("Expected: ", encoded)
		t.Log("Actual  : ", b)
		t.Error("String to bytes for packet transmission is NOT working as expected.")
	}

	s := bytesToString(bytes.NewReader(encoded))
	if s != unencoded {
		t.Log("Expected: ", unencoded)
		t.Log("Actual  : ", s)
		t.Error("Bytes to string for packet transmission is NOT working as expected.")
	}
}

func TestKeepAlive(t *testing.T) {
	t.Parallel()

	p1 := KeepAlive{ID: 123456}
	p2 := ReadKeepAlive(bytes.NewReader(p1.Packet()[1:]))

	if p1 != p2 {
		t.Log("Expected: ", p1)
		t.Log("Actual  : ", p2)
		t.Error("KeepAlive packet is not being encoded/decoded properly.")
	}
}

func TestLoginRequest(t *testing.T) {
	t.Parallel()

	p1 := LoginRequest{EntityID: 1298, LevelType: "default", ServerMode: Survival, Dimension: Overworld, Difficulty: Normal, MaxPlayers: 20}
	p2 := ReadLoginRequest(bytes.NewReader(p1.Packet()[1:]))

	if p1 != p2 {
		t.Log("Expected: ", p1)
		t.Log("Actual  : ", p2)
		t.Error("LoginRequest packet is not being encoded/decoded properly.")
	}
}
