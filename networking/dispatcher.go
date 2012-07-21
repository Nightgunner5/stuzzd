package networking

import (
	"fmt"
	"github.com/Nightgunner5/stuzzd/config"
	"github.com/Nightgunner5/stuzzd/protocol"
	"strings"
	"net/http"
)

func dispatchPacket(p Player, packet protocol.Packet) {
	switch pkt := packet.(type) {
	case protocol.KeepAlive:
		return
	case protocol.LoginRequest:
		if pkt.EntityID != protocol.PROTOCOL_VERSION {
			panic("Your minecraft version isn't the one I expected.")
		}
		if pkt.Username != p.Username() {
			panic(fmt.Sprint("Your username doesn't match the one you told me earlier. (", pkt.Username, " != ", p.Username(), ")"))
		}
		req, _ := http.Get(fmt.Sprintf("http://session.minecraft.net/game/checkserver.jsp?user=%s&serverId=%x", p.Username(), p.getLoginToken()))
		buf := make([]byte, 3)
		req.Body.Read(buf)
		if string(buf) == "YES" {
			go p.SendPacket(protocol.LoginRequest{EntityID: p.ID(), LevelType: "default", ServerMode: protocol.Survival, Dimension: protocol.Overworld, Difficulty: protocol.Peaceful, MaxPlayers: config.NumSlots()})
		} else {
			go p.SendPacket(protocol.Kick{Reason: "Failed to verify username!"})
		}
	case protocol.Handshake:
		data := strings.Split(pkt.Data, ";")
		p.setUsername(data[0])
		go p.SendPacket(protocol.Handshake{fmt.Sprintf("%016x", p.getLoginToken())})
	case protocol.ServerListPing:
		go p.SendPacket(protocol.Kick{Reason: fmt.Sprintf("%s§%d§%d", config.ServerDescription(), OnlinePlayerCount, config.NumSlots())})
	default:
		panic(fmt.Sprintf("%T %v", packet, packet))
	}
}
