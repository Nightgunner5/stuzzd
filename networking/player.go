package networking

import (
	"fmt"
	"github.com/Nightgunner5/stuzzd/protocol"
	"io"
	"log"
	"math/rand"
	"net"
	"strings"
	"time"
)

var OnlinePlayerCount uint8

func HandlePlayer(conn net.Conn) Player {
	p := new(player)
	p.id = assignID()
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
				OnlinePlayerCount--
			}
			time.Sleep(1 * time.Second)
			conn.Close()
		}()
		recvq := make(chan protocol.Packet)
		go recv(p, conn, recvq)
		sendKeepAlive := time.Tick(time.Second * 50) // 1000 ticks
		//timeoutKeepAlive := time.After(time.Second * 60) // 1200 ticks
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
				//if ka, ok := packet.(protocol.KeepAlive); ok && ka.ID == 0 {
				//	timeoutKeepAlive = time.After(time.Second * 60)
				//}
				dispatchPacket(p, packet)
				if _, ok := packet.(protocol.Kick); ok {
					return
				}
			case <-sendKeepAlive:
				p.SendPacket(protocol.KeepAlive{7031})
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
			p.SendPacket(protocol.Kick{fmt.Sprint("Error: ", err)})
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
		case 0x13:
			in.Read(make([]byte, 5)) // TODO
		case 0x47: // When the server sends this, it's a lightning bolt. When the client sends it, it means they're doing a HTTP GET. Switch over to the HTTP handler.
			recvq <- SwitchToHttp{}
		case 0xCA:
			recvq <- protocol.ReadPlayerAbilities(in)
		case 0xFE:
			recvq <- protocol.ReadServerListPing(in)
		case 0xFF:
			recvq <- protocol.ReadKick(in)
		default:
			panic(fmt.Sprint("Unknown packet ID dropped: ", id[0]))
		}
	}
}

type Player interface {
	Entity

	// Adds a packet to the send queue for a player.
	// This function may be called as a new goroutine if the code sending the packet is not
	// something that should wait for each player.
	SendPacket(protocol.Packet)

	Username() string

	Authenticated() bool

	Position() (x, y, z float64)
	SetPosition(x, y, z float64)

	// Sets the position and sends the offset to all online players.
	SendPosition(x, y, z float64)

	makeSpawnPacket() protocol.SpawnNamedEntity
	sendSpawnPacket()
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
	pitch, yaw    float32 // radians, not degrees (!)
	movecounter   uint8
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

func (p *player) SendPacket(packet protocol.Packet) {
	go func() {
		p.sendq <- packet
	}()
}

func (p *player) Authenticated() bool {
	return p.authenticated
}

func (p *player) SendPosition(x, y, z float64) {
	defer func() {
		if recover() != nil {
			SendToAll(protocol.EntityTeleport{
				ID:    p.id,
				X:     p.x,
				Y:     p.y,
				Z:     p.z,
				Yaw:   p.yaw,
				Pitch: p.pitch,
			})
			p.SendPacket(protocol.PlayerPositionLook{X: p.x, Y1: p.y + 1, Y2: p.y, Z: p.z, Yaw: deg(p.yaw), Pitch: deg(p.pitch)})
		}
	}()
	// TEMPORARY
	if y < 63.5 {
		if p.y < 64 {
			p.y = 64
		}
		panic("TEMPORARY")
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

func (p *player) SetPosition(x, y, z float64) {
	p.x, p.y, p.z = x, y, z
}

func (p *player) Position() (x, y, z float64) {
	return p.x, p.y, p.z
}

func (p *player) makeSpawnPacket() protocol.SpawnNamedEntity {
	return protocol.SpawnNamedEntity{EID: p.id, Name: p.username, X: p.x, Y: p.y, Z: p.z}
}

func (p *player) sendSpawnPacket() {
	p.SendPacket(protocol.PlayerPositionLook{X: p.x, Y1: p.y + 2.7, Y2: p.y + 3.7, Z: p.z, Ground: true})
}

func sendPacket(p Player, conn net.Conn, packet protocol.Packet) {
	if kick, ok := packet.(protocol.Kick); ok {
		if !strings.Contains(kick.Reason, "§") {
			log.Print("Kicking ", conn.RemoteAddr(), " - ", kick.Reason)
			SendToAll(protocol.Chat{Message: fmt.Sprintf("%s was kicked: %s", p.Username(), kick.Reason)})
		}
	}
	if _, err := conn.Write(packet.Packet()); err != nil {
		panic(err)
	}
}

func safeSendPacket(p Player, conn net.Conn, packet protocol.Packet) {
	defer func() { recover() }()
	sendPacket(p, conn, packet)
}

func SendToAll(packet protocol.Packet) {
	for _, player := range players {
		if player.Authenticated() {
			player.SendPacket(packet)
		}
	}
}

func SendToAllExcept(exclude Player, packet protocol.Packet) {
	for _, player := range players {
		if player.ID() == exclude.ID() {
			continue
		}
		if player.Authenticated() {
			player.SendPacket(packet)
		}
	}
}
