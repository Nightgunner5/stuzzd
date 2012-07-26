// See http://wiki.vg/Protocol
package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
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

func CheckedFloatToByte(in float64) int8 {
	if in > 4 || in < -4 {
		panic("Out of range float")
	}
	if in*32 >= 127 {
		return 127
	}
	if in*32 <= -128 {
		return -128
	}
	return int8(in * 32)
}

func encodeDouble(d float64, out io.Writer) {
	binary.Write(out, binary.BigEndian, int32(d*32))
}

func decodeDouble(in io.Reader) float64 {
	var d int32
	binary.Read(in, binary.BigEndian, &d)
	return float64(d) / 32
}

func encodeAngle(a float32, out io.Writer) {
	out.Write([]byte{uint8(a / 180 * 128)})
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

// Chat (0x03)
type Chat struct {
	Message string
}

func (p Chat) Packet() []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, uint8(0x03))
	buf.Write(stringToBytes(p.Message))
	return buf.Bytes()
}

func ReadChat(in io.Reader) Chat {
	var p Chat
	p.Message = bytesToString(in)
	for _, r := range []rune(p.Message) {
		if !strings.ContainsRune(" !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_'abcdefghijklmnopqrstuvwxyz{|}~⌂ÇüéâäàåçêëèïîìÄÅÉæÆôöòûùÿÖÜø£Ø×ƒáíóúñÑªº¿®¬½¼¡«»", r) {
			panic("Illegal character in string")
		}
	}
	return p
}

// Time Update (0x04)
type TimeUpdate struct {
	Time uint64
}

func (p TimeUpdate) Packet() []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, uint8(0x04))
	binary.Write(&buf, binary.BigEndian, p.Time)
	return buf.Bytes()
}

// No read function as this is not sent by the client.

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

// Player Digging (0x0E)
type PlayerDigging struct {
	Status uint8
	X      int32
	Y      uint8
	Z      int32
	Face   Face
}

func (p PlayerDigging) Packet() []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, uint8(0x0E))
	binary.Write(&buf, binary.BigEndian, p.Status)
	binary.Write(&buf, binary.BigEndian, p.X)
	binary.Write(&buf, binary.BigEndian, p.Y)
	binary.Write(&buf, binary.BigEndian, p.Z)
	binary.Write(&buf, binary.BigEndian, p.Face)
	return buf.Bytes()
}

func ReadPlayerDigging(in io.Reader) PlayerDigging {
	var p PlayerDigging
	binary.Read(in, binary.BigEndian, &p.Status)
	binary.Read(in, binary.BigEndian, &p.X)
	binary.Read(in, binary.BigEndian, &p.Y)
	binary.Read(in, binary.BigEndian, &p.Z)
	binary.Read(in, binary.BigEndian, &p.Face)
	return p
}

// Animation (0x12)
type Animation struct {
	EID       uint32
	Animation uint8
}

func (p Animation) Packet() []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, uint8(0x12))
	binary.Write(&buf, binary.BigEndian, p.EID)
	binary.Write(&buf, binary.BigEndian, p.Animation)
	return buf.Bytes()
}

func ReadAnimation(in io.Reader) Animation {
	var p Animation
	binary.Read(in, binary.BigEndian, &p.EID)
	binary.Read(in, binary.BigEndian, &p.Animation)
	return p
}

// Spawn Named Entity (0x14)
type SpawnNamedEntity struct {
	EID        uint32
	Name       string
	X, Y, Z    float64
	Yaw, Pitch float32
	ItemInHand uint16
}

func (p SpawnNamedEntity) Packet() []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, uint8(0x14))
	binary.Write(&buf, binary.BigEndian, p.EID)
	buf.Write(stringToBytes(p.Name))
	encodeDouble(p.X, &buf)
	encodeDouble(p.Y, &buf)
	encodeDouble(p.Z, &buf)
	encodeAngle(p.Yaw, &buf)
	encodeAngle(p.Pitch, &buf)
	binary.Write(&buf, binary.BigEndian, p.ItemInHand)
	return buf.Bytes()
}

// No read function as this is not sent by the client.

// Destroy Entity (0x1D)
type DestroyEntity struct {
	ID uint32
}

func (p DestroyEntity) Packet() []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, uint8(0x1D))
	binary.Write(&buf, binary.BigEndian, p.ID)
	return buf.Bytes()
}

// No read function as this is not sent by the client.

// Entity Relative Move (0x1F)
type EntityRelativeMove struct {
	ID      uint32
	X, Y, Z int8 // Encoded at creation time, not write time, to prevent incorrect kicks.
}

func (p EntityRelativeMove) Packet() []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, uint8(0x1F))
	binary.Write(&buf, binary.BigEndian, p.ID)
	binary.Write(&buf, binary.BigEndian, p.X)
	binary.Write(&buf, binary.BigEndian, p.Y)
	binary.Write(&buf, binary.BigEndian, p.Z)
	return buf.Bytes()
}

// No read function as this is not sent by the client.

// Entity Look (0x20)
type EntityLook struct {
	ID         uint32
	Yaw, Pitch float32
}

func (p EntityLook) Packet() []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, uint8(0x20))
	binary.Write(&buf, binary.BigEndian, p.ID)
	encodeAngle(p.Yaw, &buf)
	encodeAngle(p.Pitch, &buf)
	return buf.Bytes()
}

// No read function as this is not sent by the client.

// Entity Teleport (0x22)
type EntityTeleport struct {
	ID         uint32
	X, Y, Z    float64
	Yaw, Pitch float32
}

func (p EntityTeleport) Packet() []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, uint8(0x22))
	binary.Write(&buf, binary.BigEndian, p.ID)
	encodeDouble(p.X, &buf)
	encodeDouble(p.Y, &buf)
	encodeDouble(p.Z, &buf)
	encodeAngle(p.Yaw, &buf)
	encodeAngle(p.Pitch, &buf)
	return buf.Bytes()
}

// No read function as this is not sent by the client.

// Entity Head Look (0x23)
type EntityHeadLook struct {
	ID  uint32
	Yaw float32
}

func (p EntityHeadLook) Packet() []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, uint8(0x23))
	binary.Write(&buf, binary.BigEndian, p.ID)
	encodeAngle(p.Yaw, &buf)
	return buf.Bytes()
}

// No read function as this is not sent by the client.

// Chunk Allocation (0x32)
type ChunkAllocation struct {
	X, Z int32
	Init bool
}

func (p ChunkAllocation) Packet() []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, uint8(0x32))
	binary.Write(&buf, binary.BigEndian, p.X)
	binary.Write(&buf, binary.BigEndian, p.Z)
	var init uint8
	if p.Init {
		init = 1
	}
	binary.Write(&buf, binary.BigEndian, init)
	return buf.Bytes()
}

// No read function as this is not sent by the client.

// Chunk Data (0x33)
type ChunkData struct {
	X     int32
	Z     int32
	Chunk *Chunk
}

func (p ChunkData) Packet() []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, uint8(0x33))
	binary.Write(&buf, binary.BigEndian, p.X)
	binary.Write(&buf, binary.BigEndian, p.Z)
	binary.Write(&buf, binary.BigEndian, uint8(1))
	binary.Write(&buf, binary.BigEndian, ^uint16(0))
	binary.Write(&buf, binary.BigEndian, uint16(0))

	payload := p.Chunk.Compressed()

	binary.Write(&buf, binary.BigEndian, int32(len(payload)))
	binary.Write(&buf, binary.BigEndian, int32(0))
	buf.Write(payload)
	return buf.Bytes()
}

// No read function as this is not sent by the client.

// Multi Block Change (0x34)
type MultiBlockChange struct {
	X, Z   int32
	Blocks []uint32
}

func (p MultiBlockChange) Packet() []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, uint8(0x34))
	binary.Write(&buf, binary.BigEndian, p.X)
	binary.Write(&buf, binary.BigEndian, p.Z)
	binary.Write(&buf, binary.BigEndian, uint16(len(p.Blocks)))
	binary.Write(&buf, binary.BigEndian, uint32(len(p.Blocks))*4)
	binary.Write(&buf, binary.BigEndian, p.Blocks)
	return buf.Bytes()
}

// No read function as this is not sent by the client.

// Block Change (0x35)
type BlockChange struct {
	X     int32
	Y     uint8
	Z     int32
	Block BlockType
	Data  uint8
}

func (p BlockChange) Packet() []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, uint8(0x35))
	binary.Write(&buf, binary.BigEndian, p.X)
	binary.Write(&buf, binary.BigEndian, p.Y)
	binary.Write(&buf, binary.BigEndian, p.Z)
	binary.Write(&buf, binary.BigEndian, p.Block)
	binary.Write(&buf, binary.BigEndian, p.Data)
	return buf.Bytes()
}

// No read function as this is not sent by the client.

// Player List Item (0xC9)
type PlayerListItem struct {
	Name   string
	Online bool
	Ping   uint16
}

func (p PlayerListItem) Packet() []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, uint8(0xC9))
	buf.Write(stringToBytes(p.Name))
	var online uint8
	if p.Online {
		online = 1
	}
	binary.Write(&buf, binary.BigEndian, online)
	binary.Write(&buf, binary.BigEndian, p.Ping)
	return buf.Bytes()
}

// No read function as this is not sent by the client.

// Player Abilities (0xCA)
type PlayerAbilities struct {
	Invulnerable   bool
	Flying         bool
	CanFly         bool
	InstantDestroy bool
}

func (p PlayerAbilities) Packet() []byte {
	var invulnerable, flying, canfly, instantdestroy uint8
	if p.Invulnerable {
		invulnerable = 1
	}
	if p.Flying {
		flying = 1
	}
	if p.CanFly {
		canfly = 1
	}
	if p.InstantDestroy {
		instantdestroy = 1
	}

	return []byte{0xCA, invulnerable, flying, canfly, instantdestroy}
}

func ReadPlayerAbilities(in io.Reader) PlayerAbilities {
	var p PlayerAbilities

	b := make([]byte, 4)
	in.Read(b)

	if b[0] == 1 {
		p.Invulnerable = true
	}
	if b[1] == 1 {
		p.Flying = true
	}
	if b[2] == 1 {
		p.CanFly = true
	}
	if b[3] == 1 {
		p.InstantDestroy = true
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
