package networking

import (
	"fmt"
	"github.com/Nightgunner5/stuzzd/block"
	"github.com/Nightgunner5/stuzzd/chunk"
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
			p.SendPacketSync(protocol.Kick{Reason: "Your minecraft version isn't the one I expected."})
			return
		}
		if pkt.Username != p.Username() {
			p.SendPacketSync(protocol.Kick{Reason: fmt.Sprint("Your username doesn't match the one you told me earlier. (", pkt.Username, " != ", p.Username(), ")")})
			return
		}
		req, _ := http.Get(fmt.Sprintf("http://session.minecraft.net/game/checkserver.jsp?user=%s&serverId=%x", p.Username(), p.getLoginToken()))
		buf := make([]byte, 3)
		req.Body.Read(buf)
		if string(buf) == "YES" {
			p.(*_player).authenticated = true
			if OnlinePlayerCount >= config.Config.NumSlots {
				p.SendPacketSync(protocol.Kick{Reason: "Server is full!"})
				return
			}
			OnlinePlayerCount++
			p.SendPacketSync(protocol.LoginRequest{
				EntityID:   p.ID(),
				LevelType:  "default",
				ServerMode: protocol.Survival,
				Dimension:  protocol.Overworld,
				Difficulty: protocol.Peaceful,
				MaxPlayers: uint8(config.Config.NumSlots), // If you have more than 255 slots, I applaud you.
			})
			p.(*_player).stored = storage.GetPlayer(p.Username())
			if p.(*_player).stored.Abilities.InstaBuild {
				p.SetGameMode(protocol.Creative)
			} else {
				p.SetGameMode(protocol.Survival)
			}
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
			if p.(*_player).stored.Abilities.InstaBuild {
				if GetBlockAt(pkt.X, int32(pkt.Y), pkt.Z) != block.Bedrock {
					PlayerSetBlockAt(pkt.X, int32(pkt.Y), pkt.Z, block.Air, 0)
				}
			}
		case 2:
			blockType := GetBlockAt(pkt.X, int32(pkt.Y), pkt.Z)
			if blockType != block.Bedrock {
				item := blockType.ItemDrop()

				if item != 0 {
					DropItem(float64(pkt.X)+0.5, float64(pkt.Y)+0.5, float64(pkt.Z)+0.5, item, GetBlockDataAt(pkt.X, int32(pkt.Y), pkt.Z))
				}

				PlayerSetBlockAt(pkt.X, int32(pkt.Y), pkt.Z, block.Air, 0)
			}
		}
		// TODO: validation
	case protocol.Animation:
		if pkt.EID == p.ID() && pkt.Animation == 1 {
			SendToAllExcept(p, pkt)
		}
	case protocol.PlayerAbilities:
		p.(*_player).stored.Abilities.Flying = p.(*_player).stored.Abilities.MayFly && pkt.Flying
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

func sendChunk(p Player, x, z int32, chunk *chunk.Chunk) {
	if chunk == nil {
		p.SendPacketSync(protocol.ChunkAllocation{X: x, Z: z, Init: false})
	} else {
		p.SendPacketSync(protocol.ChunkAllocation{X: x, Z: z, Init: true})
		p.SendPacketSync(chunk)
		p.SendPacketSync(chunk.EntitySpawnPacket())
	}
}

func SaveAllTheThings() {
	SendToAll(protocol.Kick{Reason: "Server is shutting down!"})
	storage.SaveAndUnloadAllChunks()
	storage.SaveAllPlayers()
	time.Sleep(time.Second) // Give the kicks a little time to be recieved so the players get a useful message instead of "connection reset".
}
