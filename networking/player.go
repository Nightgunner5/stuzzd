package networking

import (
	"fmt"
	"github.com/Nightgunner5/stuzzd/chunk"
	"github.com/Nightgunner5/stuzzd/config"
	"github.com/Nightgunner5/stuzzd/protocol"
	"github.com/Nightgunner5/stuzzd/storage"
	"io"
	"log"
	"math"
	"math/rand"
	"net"
	"runtime"
	"strings"
	"time"
)

var OnlinePlayerCount uint64

func HandlePlayer(conn net.Conn) Player {
	p := new(player)
	p.id = assignID()
	p.chunkSet = make(map[uint64]*chunk.Chunk)
	p.sendq = make(chan protocol.Packet)

	go func() {
		defer func() {
			err := recover()
			if err != nil {
				if pkt, ok := err.(protocol.Kick); ok {
					safeSendPacket(p, conn, pkt)
				} else {
					safeSendPacket(p, conn, protocol.Kick{Reason: fmt.Sprint("Error: ", err)})
				}
			}
			RemoveEntity(p)
			if p.authenticated {
				stored := storage.GetPlayer(p.Username())
				x, y, z := p.Position()
				stored.Position = []float64{x, y, z}
				yaw, pitch := p.Angles()
				stored.Rotation = []float32{yaw, pitch}
				stored.Save()
				OnlinePlayerCount--
			}
			time.Sleep(1 * time.Second)
			SendToAllExcept(p, protocol.PlayerListItem{Name: p.Username(), Online: false, Ping: 0})
			conn.Close()
		}()
		recvq := make(chan protocol.Packet)
		go recv(p, conn, recvq)
		sendKeepAlive := time.Tick(time.Second * 50) // 1000 ticks
		for {
			select {
			case packet := <-p.sendq:
				if _, ok := packet.(protocol.Kick); ok {
					panic(packet)
				}
				sendPacket(p, conn, packet)
			case packet := <-recvq:
				if _, ok := packet.(SwitchToHttp); ok {
					sendHTTPResponse(conn)
					return
				}
				go dispatchPacket(p, packet)
				if _, ok := packet.(protocol.Kick); ok {
					return
				}
			case <-sendKeepAlive:
				go p.SendPacketSync(protocol.KeepAlive{7031})
				//case <-timeoutKeepAlive:
				//	panic("Connection timed out")
			}
		}
	}()
	return p
}

func sendHTTPResponse(conn net.Conn) {
	io.WriteString(conn, `HTTP/1.0 200 OK
Content-Type: text/html; charset=UTF-8
Server: StuzzD
Connection: close
`)
	page := `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>StuzzD Server Status</title>
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<link href="http://www.stuzzhosting.com/css/style.css" rel="stylesheet">

<!-- Le HTML5 shim, for IE6-8 support of HTML5 elements -->
<!--[if lt IE 9]>
<script src="http://html5shim.googlecode.com/svn/trunk/html5.js"></script>
<![endif]-->
</head>
<body>
<div class="container"><div class="row"><div class="span4"><img src="http://www.stuzzhosting.com/img/logo.png"></div></div></div>
<div class="hero-unit"><div class="container"><div class="row"><div class="span7">
<h1>StuzzD Server Status</h1>
<p>Thank you for connecting to a Minecraft server over HTTP. You now know a secret way to figure out if a StuzzHosting Minecraft server is up.</p>
</div></div></div></div>
<footer><div class="copyrights"><div class="container"><p>Copyright &copy; <a href="http://www.stuzzhosting.com" rel="tooltip" title="StuzzHosting is Best Hosting">StuzzHosting.com</a> 2012</p><p>Also, why the hell would you connect to a Minecraft server over HTTP?</p></div></div></footer>
<script src="http://www.stuzzhosting.com/js/jquery.js"></script>
<script src="http://www.stuzzhosting.com/js/bootstrap.js"></script>
</body>
</html>`
	fmt.Fprintf(conn, "Content-Length: %d\n\n%s", len(page), page)
}

type SwitchToHttp struct {
}

func (SwitchToHttp) Packet() []byte {
	panic("Connected over HTTP")
}

func recv(p Player, in io.Reader, recvq chan<- protocol.Packet) {
	defer func() {
		err := recover()
		if err != nil {
			p.SendPacketSync(protocol.Kick{fmt.Sprint("Error: ", err)})
		}
	}()
	id := make([]byte, 1)
	for {
		in.Read(id)
		switch id[0] {
		case 0x00:
			recvq <- protocol.ReadKeepAlive(in)
		case 0x01:
			recvq <- protocol.ReadLoginRequest(in)
		case 0x02:
			recvq <- protocol.ReadHandshake(in)
		case 0x03:
			recvq <- protocol.ReadChat(in)
		case 0x0A:
			recvq <- protocol.ReadFlying(in)
		case 0x0B:
			recvq <- protocol.ReadPlayerPosition(in)
		case 0x0C:
			recvq <- protocol.ReadPlayerLook(in)
		case 0x0D:
			recvq <- protocol.ReadPlayerPositionLook(in)
		case 0x0E:
			recvq <- protocol.ReadPlayerDigging(in)
		case 0x0F:
			panic("NO CLICKING!")
		case 0x10:
			in.Read(make([]byte, 2)) // TODO
		case 0x12:
			recvq <- protocol.ReadAnimation(in)
		case 0x13:
			in.Read(make([]byte, 5)) // TODO
		case 0x47: // When the server sends this, it's a lightning bolt. When the client sends it, it means they're doing a HTTP GET. Switch over to the HTTP handler.
			recvq <- SwitchToHttp{}
		case 0x65:
			in.Read(make([]byte, 1)) // TODO
		case 0xCA:
			recvq <- protocol.ReadPlayerAbilities(in)
		case 0xFE:
			recvq <- protocol.ReadServerListPing(in)
		case 0xFF:
			recvq <- protocol.ReadKick(in)
		default:
			panic(fmt.Sprintf("Unknown packet ID dropped: %x", id[0]))
		}
	}
}

type Player interface {
	Entity

	// Adds a packet to the send queue for a player.
	// This function may be called as a new goroutine if the code sending the packet is not
	// something that should wait for each player.
	//SendPacket(protocol.Packet)
	SendPacketSync(protocol.Packet)

	Username() string

	Authenticated() bool

	Position() (x, y, z float64)
	SetPosition(x, y, z float64)
	// Sets the position and sends the offset to all online players.
	SendPosition(x, y, z float64)
	ForcePosition()

	Angles() (yaw, pitch float32)
	SetAngles(yaw, pitch float32)
	// Sets the angles and sends them to all online players.
	SendAngles(yaw, pitch float32)

	SetGameMode(protocol.ServerMode)

	makeSpawnPacket() protocol.SpawnNamedEntity
	sendWorldData()
	setUsername(string)
	getLoginToken() uint64
}

type player struct {
	entity
	username      string
	logintoken    uint64
	authenticated bool
	sendq         chan protocol.Packet
	x, y, z       float64
	pitch, yaw    float32
	movecounter   uint8
	lastMoveTick  uint64
	gameMode      protocol.ServerMode
	chunkSet      map[uint64]*chunk.Chunk
	spawned       bool
}

func (p *player) setUsername(username string) {
	p.username = username
	p.logintoken = uint64(rand.Int63())
}

func (p *player) Username() string {
	return p.username
}

func (p *player) getLoginToken() uint64 {
	return p.logintoken
}

//func (p *player) SendPacket(packet protocol.Packet) {
//	go p.SendPacketSync(packet)
//}

func (p *player) SendPacketSync(packet protocol.Packet) {
	p.sendq <- packet
}

func (p *player) Authenticated() bool {
	return p.authenticated
}

func (p *player) SendPosition(x, y, z float64) {
	if !p.spawned {
		return
	}
	if p.lastMoveTick == config.Tick {
		return
	}
	p.lastMoveTick = config.Tick
	defer func() {
		if recover() != nil {
			p.ForcePosition()
		}
	}()
	if y < 0 {
		p.y = 128
		panic("fell out of world")
	}
	blockX, blockY, blockZ := int32(math.Floor(x)), int32(math.Floor(y)), int32(math.Floor(z))
	if block := GetBlockAt(blockX, blockY, blockZ); !block.Passable() && !block.SemiPassable() {
		panic("inside a block")
	}

	p.movecounter++
	if p.movecounter < 10 {
		SendToAllExcept(p, protocol.EntityRelativeMove{
			ID: p.id,
			X:  protocol.CheckedFloatToByte(x - p.x),
			Y:  protocol.CheckedFloatToByte(y - p.y),
			Z:  protocol.CheckedFloatToByte(z - p.z),
		})
	}

	p.x, p.y, p.z = x, y, z

	if p.movecounter >= 10 {
		SendToAllExcept(p, protocol.EntityTeleport{
			ID: p.id,
			X:  p.x,
			Y:  p.y, Z: p.z,
			Yaw: p.yaw, Pitch: p.pitch,
		})
		p.movecounter = 0
	}
}

func (p *player) ForcePosition() {
	SendToAllExcept(p, protocol.EntityTeleport{
		ID:    p.id,
		X:     p.x,
		Y:     p.y,
		Z:     p.z,
		Yaw:   p.yaw,
		Pitch: p.pitch,
	})
	p.SendPacketSync(protocol.PlayerPositionLook{X: p.x, Y1: p.y + 1.2, Y2: p.y + 0.2, Z: p.z, Yaw: p.yaw, Pitch: p.pitch})

}

func (p *player) sendWorldData() {
	go func() {
		for {
			for i, chunk := range p.chunkSet {
				x, z := chunk.X, chunk.Z
				dx, dz := int32(p.x)>>4-x, int32(p.z)>>4-z
				if dx > 10 || dx < -10 || dz > 10 || dz < -10 {
					storage.ReleaseChunk(x, z)
					sendChunk(p, x, z, nil)
					delete(p.chunkSet, i)
				}
			}

			for i := int32(1); i <= 8; i++ {
				middleX, middleZ := int32(p.x/16), int32(p.z/16)
				for x := middleX - i; x < middleX+i; x++ {
					for z := middleZ - i; z < middleZ+i; z++ {
						id := uint64(uint32(x))<<32 | uint64(uint32(z))
						if _, ok := p.chunkSet[id]; !ok {
							p.chunkSet[id] = storage.GetChunk(x, z)
							sendChunk(p, x, z, p.chunkSet[id])
							runtime.Gosched()
						}
					}
				}

				if i == 2 && !p.spawned {
					p.sendSpawnPacket()
					p.spawned = true
				}
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()
}

func (p *player) SetPosition(x, y, z float64) {
	p.x, p.y, p.z = x, y, z
}

func (p *player) Position() (x, y, z float64) {
	return p.x, p.y, p.z
}

func (p *player) SendAngles(yaw, pitch float32) {
	p.yaw, p.pitch = yaw, pitch
	SendToAllExcept(p, protocol.EntityLook{ID: p.id, Yaw: yaw, Pitch: pitch})
	SendToAllExcept(p, protocol.EntityHeadLook{ID: p.id, Yaw: yaw})
}

func (p *player) SetAngles(yaw, pitch float32) {
	p.yaw, p.pitch = yaw, pitch
}

func (p *player) Angles() (yaw, pitch float32) {
	return p.yaw, p.pitch
}

func (p *player) SetGameMode(mode protocol.ServerMode) {
	p.gameMode = mode
	p.SendPacketSync(protocol.PlayerAbilities{
		Invulnerable:   mode == protocol.Creative,
		CanFly:         mode == protocol.Creative,
		InstantDestroy: mode == protocol.Creative,
	})
}

func (p *player) makeSpawnPacket() protocol.SpawnNamedEntity {
	return protocol.SpawnNamedEntity{EID: p.id, Name: p.username, X: p.x, Y: p.y, Z: p.z, Yaw: p.yaw, Pitch: p.pitch}
}

func (p *player) sendSpawnPacket() {
	p.SendPacketSync(protocol.PlayerPositionLook{X: p.x, Y1: p.y + 2, Y2: p.y + 3, Z: p.z, Ground: true, Yaw: p.yaw, Pitch: p.pitch})
}

func sendPacket(p Player, conn net.Conn, packet protocol.Packet) {
	if kick, ok := packet.(protocol.Kick); ok {
		if !strings.Contains(kick.Reason, "§") {
			if strings.HasPrefix(kick.Reason, "Error: ") {
				log.Print("Dropping ", conn.RemoteAddr(), " ", p.Username(), " - ", kick.Reason)
				if p.Username() != "" {
					SendToAll(protocol.Chat{Message: fmt.Sprintf("%s is error %s", formatUsername(p), strings.ToLower(kick.Reason)[7:])})
				}
			} else {
				log.Print("Kicking ", conn.RemoteAddr(), " ", p.Username(), " - ", kick.Reason)
				if p.Username() != "" {
					SendToAll(protocol.Chat{Message: fmt.Sprintf("%s was kicked: %s", formatUsername(p), kick.Reason)})
				}
			}
		}
		for _, chunk := range p.(*player).chunkSet {
			storage.ReleaseChunk(chunk.X, chunk.Z)
		}
	}
	if _, err := conn.Write(packet.Packet()); err != nil {
		panic(err)
	}
}

// Used for ignoring network errors when responding to what could be network errors.
func safeSendPacket(p Player, conn net.Conn, packet protocol.Packet) {
	defer func() { recover() }()
	sendPacket(p, conn, packet)
}

func SendToAll(packet protocol.Packet) {
	for _, player := range players {
		if player.Authenticated() {
			go player.SendPacketSync(packet)
		}
	}
}

func SendToAllExcept(exclude Player, packet protocol.Packet) {
	for _, player := range players {
		if player.ID() == exclude.ID() {
			continue
		}
		if player.Authenticated() {
			go player.SendPacketSync(packet)
		}
	}
}
