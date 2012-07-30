package networking

import (
	"fmt"
	"github.com/Nightgunner5/stuzzd/chunk"
	"github.com/Nightgunner5/stuzzd/config"
	"github.com/Nightgunner5/stuzzd/player"
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
	p := new(_player)
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
				storage.SaveAndUnloadPlayer(p.Username(), p.stored)
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

	// Sets the position and sends the offset to all online players.
	SendPosition(x, y, z float64)
	ForcePosition()

	Angles() (yaw, pitch float32)
	SetAngles(yaw, pitch float32)
	// Sets the angles and sends them to all online players.
	SendAngles(yaw, pitch float32)

	SetGameMode(protocol.ServerMode)

	sendWorldData()
	setUsername(string)
	getLoginToken() uint64
}

type _player struct {
	id            int32
	stored        *player.Player
	username      string
	logintoken    uint64
	authenticated bool
	sendq         chan protocol.Packet
	movecounter   uint8
	lastMoveTick  uint64
	gameMode      protocol.ServerMode
	chunkSet      map[uint64]*chunk.Chunk
	spawned       bool
}

func (p *_player) ID() int32 {
	return p.id
}

func (p *_player) setUsername(username string) {
	p.username = username
	p.logintoken = uint64(rand.Int63())
}

func (p *_player) Username() string {
	return p.username
}

func (p *_player) getLoginToken() uint64 {
	return p.logintoken
}

func (p *_player) SendPacketSync(packet protocol.Packet) {
	p.sendq <- packet
}

func (p *_player) Authenticated() bool {
	return p.authenticated
}

func (p *_player) SendPosition(x, y, z float64) {
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

	px, py, pz := p.Position()

	if y < 0 {
		if py < 1 {
			p.SetPosition(px, 128, pz)
		}
		panic("fell out of world")
	}

	blockX, blockY, blockZ := int32(math.Floor(x)), int32(math.Floor(y)), int32(math.Floor(z))
	if block := GetBlockAt(blockX, blockY, blockZ); !block.Passable() && !block.SemiPassable() {
		panic("inside a block")
	}

	distance := math.Sqrt((x-px)*(x-px) + (y-py)*(y-py) + (z-pz)*(z-pz))
	noFallY := y
	if noFallY < py {
		noFallY = py
	}
	distanceNoFall := math.Sqrt((x-px)*(x-px) + (noFallY-py)*(noFallY-py) + (z-pz)*(z-pz))
	if distance > 10 || distanceNoFall > 5 {
		panic("Moved too fast")
	}

	p.SetPosition(x, y, z)

	p.movecounter++
	if p.movecounter < 10 && distance < 4 {
		SendToAllExcept(p, protocol.EntityRelativeMove{
			ID: p.id,
			X:  protocol.CheckedFloatToByte(x - px),
			Y:  protocol.CheckedFloatToByte(y - py),
			Z:  protocol.CheckedFloatToByte(z - pz),
		})
	} else {
		yaw, pitch := p.Angles()
		SendToAllExcept(p, protocol.EntityTeleport{
			ID:    p.id,
			X:     x,
			Y:     y,
			Z:     z,
			Yaw:   yaw,
			Pitch: pitch,
		})
		p.movecounter = 0
	}
}

func (p *_player) ForcePosition() {
	x, y, z := p.Position()
	yaw, pitch := p.Angles()
	SendToAllExcept(p, protocol.EntityTeleport{
		ID:    p.id,
		X:     x,
		Y:     y,
		Z:     z,
		Yaw:   yaw,
		Pitch: pitch,
	})
	p.SendPacketSync(protocol.PlayerPositionLook{
		X:     x,
		Y1:    y + 1.2,
		Y2:    y + 0.2,
		Z:     z,
		Yaw:   yaw,
		Pitch: pitch,
	})

}

func (p *_player) sendWorldData() {
	go func() {
		for {
			for i, chunk := range p.chunkSet {
				x, z := chunk.X, chunk.Z
				dx, dz := (int32(p.stored.Position[0])>>4)-x, (int32(p.stored.Position[2])>>4)-z
				if dx > 10 || dx < -10 || dz > 10 || dz < -10 {
					storage.ReleaseChunk(x, z)
					sendChunk(p, x, z, nil)
					delete(p.chunkSet, i)
				}
			}

			for i := int32(1); i <= 8; i++ {
				middleX, middleZ := int32(p.stored.Position[0]/16), int32(p.stored.Position[2]/16)
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

func (p *_player) SetPosition(x, y, z float64) {
	p.stored.Position[0], p.stored.Position[1], p.stored.Position[2] = x, y, z
}

func (p *_player) Position() (x, y, z float64) {
	return p.stored.Position[0], p.stored.Position[1], p.stored.Position[2]
}

func (p *_player) SendAngles(yaw, pitch float32) {
	p.SetAngles(yaw, pitch)
	SendToAllExcept(p, protocol.EntityLook{ID: p.id, Yaw: yaw, Pitch: pitch})
	SendToAllExcept(p, protocol.EntityHeadLook{ID: p.id, Yaw: yaw})
}

func (p *_player) SetAngles(yaw, pitch float32) {
	p.stored.Rotation[0], p.stored.Rotation[1] = yaw, pitch
}

func (p *_player) Angles() (yaw, pitch float32) {
	return p.stored.Rotation[0], p.stored.Rotation[1]
}

func (p *_player) SetGameMode(mode protocol.ServerMode) {
	p.gameMode = mode
	p.stored.Abilities.InstaBuild = mode == protocol.Creative
	p.stored.Abilities.Invulnerable = mode == protocol.Creative
	p.stored.Abilities.MayFly = mode == protocol.Creative
	if !p.stored.Abilities.MayFly {
		p.stored.Abilities.Flying = false
	}
	p.SendPacketSync(&p.stored.Abilities)
	p.SendPacketSync(protocol.ChangeGameState{Type: protocol.ChangeGameMode, Mode: mode})
}

func (p *_player) SpawnPacket(w io.Writer) {
	x, y, z := p.Position()
	yaw, pitch := p.Angles()
	w.Write(protocol.SpawnNamedEntity{
		EID:   p.id,
		Name:  p.username,
		X:     x,
		Y:     y,
		Z:     z,
		Yaw:   yaw,
		Pitch: pitch,
	}.Packet())
}

func (p *_player) sendSpawnPacket() {
	x, y, z := p.Position()
	yaw, pitch := p.Angles()
	p.SendPacketSync(protocol.PlayerPositionLook{
		X:      x,
		Y1:     y + 2,
		Y2:     y + 3,
		Z:      z,
		Ground: false,
		Yaw:    yaw,
		Pitch:  pitch,
	})
}

func sendPacket(p Player, conn net.Conn, packet protocol.Packet) {
	if kick, ok := packet.(protocol.Kick); ok {
		if p.Username() != "" {
			if strings.HasPrefix(kick.Reason, "Error: ") {
				log.Print("Dropping ", conn.RemoteAddr(), " ", p.Username(), " - ", kick.Reason)
				SendToAll(protocol.Chat{Message: fmt.Sprintf("%s is error %s", formatUsername(p), strings.ToLower(kick.Reason)[7:])})
			} else {
				log.Print("Kicking ", conn.RemoteAddr(), " ", p.Username(), " - ", kick.Reason)
				SendToAll(protocol.Chat{Message: fmt.Sprintf("%s was kicked: %s", formatUsername(p), kick.Reason)})
			}
		}
		for _, chunk := range p.(*_player).chunkSet {
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
	baked := protocol.BakePacket(packet)
	for _, player := range players {
		if player.Authenticated() {
			go player.SendPacketSync(baked)
		}
	}
}

func SendToAllExcept(exclude Player, packet protocol.Packet) {
	baked := protocol.BakePacket(packet)
	for _, player := range players {
		if player.ID() == exclude.ID() {
			continue
		}
		if player.Authenticated() {
			go player.SendPacketSync(baked)
		}
	}
}

func SendToAllNearChunk(chunkX, chunkZ int32, packet protocol.Packet) {
	id := uint64(uint32(chunkX))<<32 | uint64(uint32(chunkZ))
	baked := protocol.BakePacket(packet)
	for _, player := range players {
		if _, ok := player.(*_player).chunkSet[id]; ok {
			go player.SendPacketSync(baked)
		}
	}
}
