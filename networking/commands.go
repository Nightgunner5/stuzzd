package networking

import (
	"fmt"
	"github.com/Nightgunner5/stuzzd/protocol"
	"log"
	"strings"
)

func handleCommand(player Player, command string) {
	words := strings.Split(command, " ")
	switch words[0] {
	case "me", "pl":
		// Don't spam the log.
	default:
		log.Printf("Command from %s: /%s", player.Username(), command)
	}
	switch words[0] {
	case "me":
		message := strings.Join(words[1:], " ")
		log.Printf("* %s %s", player.Username(), message)
		SendToAll(protocol.Chat{Message: fmt.Sprintf("%s %s", starUsername(player.Username()), message)})
	default:
		player.SendPacket(protocol.Chat{Message: "§cUnknown command."})
	}
}
