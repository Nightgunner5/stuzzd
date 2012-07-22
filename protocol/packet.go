// See http://wiki.vg/Protocol
package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// Functions in this file panic instead of returning errors. This allows us to jump straight to the networking goroutine
// when there's a problem. Operations that never return errors except when the universe breaks, such as writing to
// in-memory byte buffers, are not error-checked.

type Packet interface {
	Packet() []byte
}

func errorCheck(err error) {
	if err != nil {
		panic(err.Error())
	}
}

// Minecraft's "string" format is a big endian UCS-2 string.
// The length in characters is represented as a big endian signed 16 bit integer before the string.

func stringToBytes(in string) []byte {
	var buf bytes.Buffer
	runes := []rune(in)

	binary.Write(&buf, binary.BigEndian, int16(len(runes)))
	for _, r := range runes {
		binary.Write(&buf, binary.BigEndian, uint16(r))
	}
	return buf.Bytes()
}

func bytesToString(in io.Reader) string {
	var length int16
	errorCheck(binary.Read(in, binary.BigEndian, &length))
	if length > 120 { // Minecraft uses 120 as the maximum length for a client->server string.
		panic(fmt.Sprintf("String too long (%d > 120)", length))
	}
	out := make([]rune, 0, length)
	for i := int16(0); i < length; i++ {
		var r uint16
		errorCheck(binary.Read(in, binary.BigEndian, &r))
		out = append(out, rune(r))
	}
	return string(out)
}

// Keep Alive (0x00)
// Server must send every 1000 ticks with a nonzero ID. Client must send with ID 0.
type KeepAlive struct {
	ID int32
}

func (p KeepAlive) Packet() []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, int8(0x00))
	binary.Write(&buf, binary.BigEndian, p.ID)
	return buf.Bytes()
}

func ReadKeepAlive(in io.Reader) KeepAlive {
	var p KeepAlive
	errorCheck(binary.Read(in, binary.BigEndian, &p.ID))
	return p
}

// http://wiki.vg/Protocol#Login_Request_.280x01.29
const PROTOCOL_VERSION uint32 = 29

type ServerMode int32

const (
	Survival ServerMode = 0
	Creative ServerMode = 1
)

type Dimension int32

const (
	Nether    Dimension = -1
	Overworld Dimension = 0
	TheEnd    Dimension = 1
)

type Difficulty uint8

const (
	Peaceful Difficulty = 0
	Easy     Difficulty = 1
	Normal   Difficulty = 2
	Hard     Difficulty = 3
)

// Login Request (0x01)
// Sent by client after the handshake to finish logging in.
// Sent by server if the client is accepted, otherwise kick is sent.
type LoginRequest struct {
	EntityID   uint32     // Version on client->server
	Username   string     // Not used on server->client
	LevelType  string     // Not used in client->server, "default" in server->client
	ServerMode ServerMode // Not used in client->server
	Dimension  Dimension  // Not used in client->server
	Difficulty Difficulty // Not used in client->server
	Unused     uint8      // Always 0
	MaxPlayers uint8      // Not used in client->server, values 127 < x cause player list to not be drawn, values 60 < x < 128 are visually buggy.
}

func (p LoginRequest) Packet() []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, uint8(0x01))
	binary.Write(&buf, binary.BigEndian, p.EntityID)
	buf.Write(stringToBytes(p.Username))
	buf.Write(stringToBytes(p.LevelType))
	binary.Write(&buf, binary.BigEndian, p.ServerMode)
	binary.Write(&buf, binary.BigEndian, p.Dimension)
	binary.Write(&buf, binary.BigEndian, p.Difficulty)
	binary.Write(&buf, binary.BigEndian, p.Unused)
	binary.Write(&buf, binary.BigEndian, p.MaxPlayers)
	return buf.Bytes()
}

func ReadLoginRequest(in io.Reader) LoginRequest {
	var p LoginRequest
	errorCheck(binary.Read(in, binary.BigEndian, &p.EntityID))
	p.Username = bytesToString(in)
	p.LevelType = bytesToString(in)
	errorCheck(binary.Read(in, binary.BigEndian, &p.ServerMode))
	errorCheck(binary.Read(in, binary.BigEndian, &p.Dimension))
	errorCheck(binary.Read(in, binary.BigEndian, &p.Difficulty))
	errorCheck(binary.Read(in, binary.BigEndian, &p.Unused))
	errorCheck(binary.Read(in, binary.BigEndian, &p.MaxPlayers))
	return p
}

// Handshake (0x02)
// client->server, the data is "Username;server:port"
// server->client, it's a 64 bit unsigned integer in hex form that is used to verify identity
type Handshake struct {
	Data string
}

func (p Handshake) Packet() []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, uint8(0x02))
	buf.Write(stringToBytes(p.Data))
	return buf.Bytes()
}

func ReadHandshake(in io.Reader) Handshake {
	var p Handshake
	p.Data = bytesToString(in)
	return p
}

// Flying (0x0A)
type Flying struct {
	Ground bool
}

func (p Flying) Packet() []byte {
	var ground byte
	if p.Ground {
		ground = 1
	}
	return []byte{0x0A, ground}
}

func ReadFlying(in io.Reader) Flying {
	var ground uint8
	binary.Read(in, binary.BigEndian, &ground)
	if ground == 1 {
		return Flying{Ground: true}
	}
	return Flying{Ground: false}
}

// Player Position (0x0B)
type PlayerPosition struct {
	X      float64
	Y1     float64
	Y2     float64
	Z      float64
	Ground bool
}

func (p PlayerPosition) Packet() []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, p.X)
	binary.Write(&buf, binary.BigEndian, p.Y1)
	binary.Write(&buf, binary.BigEndian, p.Y2)
	binary.Write(&buf, binary.BigEndian, p.Z)
	var ground uint8
	if p.Ground {
		ground = 1
	}
	binary.Write(&buf, binary.BigEndian, ground)
	return buf.Bytes()
}

func ReadPlayerPosition(in io.Reader) PlayerPosition {
	var p PlayerPosition
	binary.Read(in, binary.BigEndian, &p.X)
	binary.Read(in, binary.BigEndian, &p.Y1)
	binary.Read(in, binary.BigEndian, &p.Y2)
	binary.Read(in, binary.BigEndian, &p.Z)
	var ground uint8
	binary.Read(in, binary.BigEndian, &ground)
	if ground == 1 {
		p.Ground = true
	}
	return p
}

// Player Look (0x0C)
type PlayerLook struct {
	Yaw    float32
	Pitch  float32
	Ground bool
}

func (p PlayerLook) Packet() []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, uint8(0x0C))
	binary.Write(&buf, binary.BigEndian, p.Yaw)
	binary.Write(&buf, binary.BigEndian, p.Pitch)
	var ground uint8
	if p.Ground {
		ground = 1
	}
	binary.Write(&buf, binary.BigEndian, ground)
	return buf.Bytes()
}

func ReadPlayerLook(in io.Reader) PlayerLook {
	var p PlayerLook
	binary.Read(in, binary.BigEndian, &p.Yaw)
	binary.Read(in, binary.BigEndian, &p.Pitch)
	var ground uint8
	binary.Read(in, binary.BigEndian, &ground)
	if ground == 1 {
		p.Ground = true
	}
	return p
}

// Player Position/Look (0x0D)
type PlayerPositionLook struct {
	X      float64
	Y1     float64
	Y2     float64
	Z      float64
	Yaw    float32
	Pitch  float32
	Ground bool
}

func (p PlayerPositionLook) Packet() []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, uint8(0x0D))
	binary.Write(&buf, binary.BigEndian, p.X)
	binary.Write(&buf, binary.BigEndian, p.Y1)
	binary.Write(&buf, binary.BigEndian, p.Y2)
	binary.Write(&buf, binary.BigEndian, p.Z)
	binary.Write(&buf, binary.BigEndian, p.Yaw)
	binary.Write(&buf, binary.BigEndian, p.Pitch)
	var ground uint8
	if p.Ground {
		ground = 1
	}
	binary.Write(&buf, binary.BigEndian, ground)
	return buf.Bytes()
}

func ReadPlayerPositionLook(in io.Reader) PlayerPositionLook {
	var p PlayerPositionLook
	binary.Read(in, binary.BigEndian, &p.X)
	binary.Read(in, binary.BigEndian, &p.Y1)
	binary.Read(in, binary.BigEndian, &p.Y2)
	binary.Read(in, binary.BigEndian, &p.Z)
	binary.Read(in, binary.BigEndian, &p.Yaw)
	binary.Read(in, binary.BigEndian, &p.Pitch)
	var ground uint8
	binary.Read(in, binary.BigEndian, &ground)
	if ground == 1 {
		p.Ground = true
	}
	return p
}

// Server List Ping (0xFE)
type ServerListPing struct{}

func (p ServerListPing) Packet() []byte {
	return []byte{0xFE}
}

func ReadServerListPing(in io.Reader) ServerListPing {
	var p ServerListPing
	return p
}

// Disconnect/Kick (0xFF)
// If this is recieved by the server, close the connection and drop the client as they have disconnected.
// If this is sent by the server, drop the client, wait a minute, and close the connection.
type Kick struct {
	Reason string // Only read by client when this acts as a kick.
}

func (p Kick) Packet() []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, uint8(0xFF))
	buf.Write(stringToBytes(p.Reason))
	return buf.Bytes()
}

func ReadKick(in io.Reader) Kick {
	var p Kick
	p.Reason = bytesToString(in)
	return p
}
