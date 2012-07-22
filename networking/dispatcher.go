package networking

import (
	"fmt"
	"github.com/Nightgunner5/stuzzd/config"
	"github.com/Nightgunner5/stuzzd/protocol"
	"log"
	"net/http"
	"strings"
	"time"
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
			p.(*player).authenticated = true
			OnlinePlayerCount++
			p.SendPacket(protocol.LoginRequest{
				EntityID:   p.ID(),
				LevelType:  "default",
				ServerMode: protocol.Creative,
				Dimension:  protocol.Overworld,
				Difficulty: protocol.Peaceful,
				MaxPlayers: config.NumSlots(),
			})
			SendWorldData(p)
			p.SetPosition(8.5, 65, 8.5)
			p.sendSpawnPacket()
			go func() {
				for i := time.Duration(0); i < 10; i++ {
					time.Sleep(i * time.Millisecond * 10)
					p.sendSpawnPacket()
				}
			}()
			log.Print(p.Username(), " connected.")
			SendToAll(protocol.Chat{Message: fmt.Sprintf("%s connected.", p.Username())})
			for _, player := range players {
				if player != p {
					p.SendPacket(player.makeSpawnPacket())
				}
			}
			SendToAllExcept(p, p.makeSpawnPacket())
		} else {
			p.SendPacket(protocol.Kick{Reason: "Failed to verify username!"})
		}
	case protocol.Chat:
		chat := fmt.Sprintf("<%s> %s", p.Username(), pkt.Message)
		log.Print(chat)
		SendToAll(protocol.Chat{Message: chat})
	case protocol.Handshake:
		data := strings.Split(pkt.Data, ";")
		p.setUsername(data[0])
		p.SendPacket(protocol.Handshake{fmt.Sprintf("%016x", p.getLoginToken())})
	case protocol.Flying:
		// TODO
	case protocol.PlayerPosition:
		p.SendPosition(pkt.X, pkt.Y1, pkt.Z)
		// TODO: validation
	case protocol.PlayerLook:
		p.SendAngles(rad(pkt.Yaw), rad(pkt.Pitch))
	case protocol.PlayerPositionLook:
		p.SendAngles(rad(pkt.Yaw), rad(pkt.Pitch))
		p.SendPosition(pkt.X, pkt.Y1, pkt.Z)
		// TODO: validation
	case protocol.PlayerAbilities:
		// TODO
	case protocol.ServerListPing:
		p.SendPacket(protocol.Kick{Reason: fmt.Sprintf("%s§%d§%d", config.ServerDescription(), OnlinePlayerCount, config.NumSlots())})
	case protocol.Kick:
		log.Print(p.Username(), " disconnected.")
		SendToAll(protocol.Chat{Message: fmt.Sprintf("%s disconnected.", p.Username())})
	default:
		panic(fmt.Sprintf("%T %v", packet, packet))
	}
}

func SendWorldData(p Player) {
	for x := int32(-16); x < 16; x++ {
		for z := int32(-16); z < 16; z++ {
			p.SendPacket(protocol.ChunkAllocation{X: x, Z: z, Init: true})
			p.SendPacket(protocol.ChunkData{X: x, Z: z, Chunk: GetChunk(x, z)})
		}
	}
}
