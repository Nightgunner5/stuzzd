package networking

import (
	"fmt"
	"github.com/Nightgunner5/stuzzd/config"
	"github.com/Nightgunner5/stuzzd/protocol"
	"github.com/Nightgunner5/stuzzd/storage"
	"log"
	"net/http"
	"strings"
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
			p.(*player).gameMode = protocol.Survival
			if isNetworkAdmin(p) {
				p.(*player).gameMode = protocol.Creative
			}
			p.SendPacket(protocol.LoginRequest{
				EntityID:   p.ID(),
				LevelType:  "default",
				ServerMode: p.(*player).gameMode,
				Dimension:  protocol.Overworld,
				Difficulty: protocol.Peaceful,
				MaxPlayers: config.NumSlots(),
			})
			stored := storage.GetPlayer(p.Username())
			p.SetPosition(stored.Position[0], stored.Position[1], stored.Position[2])
			p.SetAngles(stored.Rotation[0], stored.Rotation[1])
			p.sendWorldData()
			log.Print(p.Username(), " connected.")
			if customLoginMessage(p) != "" {
				SendToAll(protocol.Chat{Message: fmt.Sprintf(customLoginMessage(p), formatUsername(p.Username()))})
			} else {
				SendToAll(protocol.Chat{Message: fmt.Sprintf("%s connected.", formatUsername(p.Username()))})
			}
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
		if pkt.Message[0] == '/' {
			handleCommand(p, string(pkt.Message[1:]))
		} else {
			log.Printf("<%s> %s", p.Username(), pkt.Message)
			SendToAll(protocol.Chat{Message: fmt.Sprintf("%s %s", bracketUsername(p.Username()), pkt.Message)})
		}
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
	case protocol.PlayerDigging:
		switch pkt.Status {
		case 0:
			if p.(*player).gameMode != protocol.Creative {
				break
			}
			fallthrough
		case 2:
			if GetBlockAt(pkt.X, int32(pkt.Y), pkt.Z) != protocol.Bedrock {
				SetBlockAt(pkt.X, int32(pkt.Y), pkt.Z, protocol.Air, 0)
			}
		}
		// TODO: validation
	case protocol.Animation:
		if pkt.EID == p.ID() && pkt.Animation == 1 {
			SendToAllExcept(p, pkt)
		}
	case protocol.PlayerAbilities:
		// TODO
	case protocol.ServerListPing:
		p.SendPacket(protocol.Kick{Reason: fmt.Sprintf("%s§%d§%d", config.ServerDescription(), OnlinePlayerCount, config.NumSlots())})
	case protocol.Kick:
		log.Print(p.Username(), " disconnected.")
		SendToAll(protocol.Chat{Message: fmt.Sprintf("%s disconnected.", formatUsername(p.Username()))})
	default:
		panic(fmt.Sprintf("%T %v", packet, packet))
	}
}

func formatUsername(name string) string {
	return "§6§l" + name + "§r§7"
}

func bracketUsername(name string) string {
	return "§7<" + formatUsername(name) + ">§r"
}

func starUsername(name string) string {
	return "§7* " + formatUsername(name) + "§r"
}

func sendChunk(p Player, x, z int32, chunk *protocol.Chunk) {
	if chunk == nil {
		log.Printf("Unloading chunk at (%d, %d) for %s", x, z, p.Username())
		p.SendPacket(protocol.ChunkAllocation{X: x, Z: z, Init: false})
	} else {
		log.Printf("Sending chunk at (%d, %d) to %s", x, z, p.Username())
		p.SendPacketSync(protocol.ChunkAllocation{X: x, Z: z, Init: true})
		p.SendPacketSync(protocol.ChunkData{X: x, Z: z, Chunk: chunk})
	}
}

func SaveAllTheThings() {
	SendToAll(protocol.Kick{Reason: "Server is shutting down!"})
}
