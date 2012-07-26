package protocol

import (
	"bytes"
	"reflect"
	"testing"
)

func TestString(t *testing.T) {
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
