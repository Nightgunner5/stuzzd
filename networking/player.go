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
					safeSendPacket(conn, pkt)
				} else {
					safeSendPacket(conn, protocol.Kick{fmt.Sprint("Error: ", err)})
				}
			} else {
				log.Print("!!! Exited without an error: ", conn.RemoteAddr(), p.Username())
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
		sendKeepAlive := time.Tick(time.Second * 50)     // 1000 ticks
		timeoutKeepAlive := time.After(time.Second * 60) // 1200 ticks
		for {
			select {
			case packet := <-p.sendq:
				if _, ok := packet.(protocol.Kick); ok {
					panic(packet)
				}
				sendPacket(conn, packet)
			case packet := <-recvq:
				if ka, ok := packet.(protocol.KeepAlive); ok && ka.ID == 0 {
					timeoutKeepAlive = time.After(time.Second * 60)
				}
				dispatchPacket(p, packet)
			case <-sendKeepAlive:
				p.SendPacket(protocol.KeepAlive{7031})
			case <-timeoutKeepAlive:
				panic("Connection timed out")
			}
		}
	}()
	return p
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
		case 0x0A:
			recvq <- protocol.ReadFlying(in)
		case 0x0B:
			recvq <- protocol.ReadPlayerPosition(in)
		case 0x0C:
			recvq <- protocol.ReadPlayerLook(in)
		case 0x0D:
			recvq <- protocol.ReadPlayerPositionLook(in)
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

	setUsername(string)
	getLoginToken() uint64
}

type player struct {
	entity
	username      string
	logintoken    uint64
	authenticated bool
	sendq         chan protocol.Packet
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

func sendPacket(conn net.Conn, packet protocol.Packet) {
	if kick, ok := packet.(protocol.Kick); ok {
		if !strings.Contains(kick.Reason, "§") {
			log.Print("Kicking ", conn.RemoteAddr(), " - ", kick.Reason)
		}
	}
	if _, err := conn.Write(packet.Packet()); err != nil {
		panic(err)
	}
}

func safeSendPacket(conn net.Conn, packet protocol.Packet) {
	defer func() { recover() }()
	sendPacket(conn, packet)
}
