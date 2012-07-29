package networking

import (
	"fmt"
	"github.com/Nightgunner5/stuzzd/block"
	"github.com/Nightgunner5/stuzzd/config"
	"github.com/Nightgunner5/stuzzd/protocol"
	"github.com/Nightgunner5/stuzzd/storage"
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
			if OnlinePlayerCount >= config.Config.NumSlots {
				p.SendPacketSync(protocol.Kick{Reason: "Server is full!"})
				return
			}
			OnlinePlayerCount++
			p.(*player).gameMode = protocol.Survival
			if isNetworkAdmin(p) {
				p.(*player).gameMode = protocol.Creative
			}
			p.SendPacketSync(protocol.LoginRequest{
				EntityID:   p.ID(),
				LevelType:  "default",
				ServerMode: p.(*player).gameMode,
				Dimension:  protocol.Overworld,
				Difficulty: protocol.Peaceful,
				MaxPlayers: uint8(config.Config.NumSlots), // If you have more than 255 slots, I applaud you.
			})
			stored := storage.GetPlayer(p.Username())
			p.SetPosition(stored.Position[0], stored.Position[1], stored.Position[2])
			p.SetAngles(stored.Rotation[0], stored.Rotation[1])
			p.sendWorldData()
			log.Print(p.Username(), " connected.")
			if customLoginMessage(p) != "" {
				SendToAll(protocol.Chat{Message: fmt.Sprintf(customLoginMessage(p), formatUsername(p))})
			} else {
				SendToAll(protocol.Chat{Message: fmt.Sprintf("%s connected.", formatUsername(p))})
			}
			for _, player := range players {
				if player.Authenticated() && player != p {
					// Spawn them for the new guy
					p.SendPacketSync(player.makeSpawnPacket())
				}
			}
			SendToAllExcept(p, p.makeSpawnPacket())
		} else {
			p.SendPacketSync(protocol.Kick{Reason: "Failed to verify username!"})
		}
	case protocol.Chat:
		if pkt.Message[0] == '/' {
			handleCommand(p, string(pkt.Message[1:]))
		} else {
			log.Printf("<%s> %s", p.Username(), pkt.Message)
			SendToAll(protocol.Chat{Message: fmt.Sprintf("%s %s", bracketUsername(p), pkt.Message)})
		}
	case protocol.Handshake:
		data := strings.Split(pkt.Data, ";")
		p.setUsername(data[0])
		p.SendPacketSync(protocol.Handshake{fmt.Sprintf("%016x", p.getLoginToken())})
	case protocol.Flying:
		// TODO
	case protocol.PlayerPosition:
		p.SendPosition(pkt.X, pkt.Y1, pkt.Z)
		// TODO: validation
	case protocol.PlayerLook:
		p.SendAngles(pkt.Yaw, pkt.Pitch)
	case protocol.PlayerPositionLook:
		p.SendAngles(pkt.Yaw, pkt.Pitch)
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
			if GetBlockAt(pkt.X, int32(pkt.Y), pkt.Z) != block.Bedrock {
				PlayerSetBlockAt(pkt.X, int32(pkt.Y), pkt.Z, block.Air, 0)
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
		p.SendPacketSync(protocol.Kick{Reason: fmt.Sprintf("%s§%d§%d", config.Config.ServerDescription, OnlinePlayerCount, config.Config.NumSlots)})
	case protocol.Kick:
		log.Print(p.Username(), " disconnected.")
		SendToAll(protocol.Chat{Message: fmt.Sprintf("%s disconnected.", formatUsername(p))})
	default:
		panic(fmt.Sprintf("%T %v", packet, packet))
	}
}

func init() {
	go func() {
		for {
			time.Sleep(1 * time.Second)
			for _, player := range players {
				if player.Authenticated() {
					SendToAll(protocol.PlayerListItem{Name: player.Username(), Online: true, Ping: 50})
				}
			}
		}
	}()
}

func sendChunk(p Player, x, z int32, chunk *Chunk) {
	if chunk == nil {
		p.SendPacketSync(protocol.ChunkAllocation{X: x, Z: z, Init: false})
	} else {
		p.SendPacketSync(protocol.ChunkAllocation{X: x, Z: z, Init: true})
		p.SendPacketSync(protocol.ChunkData{X: x, Z: z, Payload: chunk.Compressed()})
	}
}

func SaveAllTheThings() {
	SendToAll(protocol.Kick{Reason: "Server is shutting down!"})
	chunkLock.Lock()
	for _, chunk := range chunks {
		chunk.Save()
	}
	chunkLock.Unlock()
	time.Sleep(time.Second)
}
