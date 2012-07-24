package networking

import (
	"fmt"
	"github.com/Nightgunner5/stuzzd/protocol"
	"log"
	"strings"
	"os"
	"bufio"
	"io"
	"sync"
)

const (
	ChatInfo         = "§7"
	ChatError        = "§c"
	ChatPayload      = "§r"
	ChatName         = "§6§l"
	ChatNameOp       = "§4§l"
	ChatNameNetAdmin = "§3§l"
)

func handleCommand(player Player, command string) {
	words := strings.Split(command, " ")
	switch words[0] {
	case "me", "pl", "players", "list", "who":
		// Don't spam the log.
	default:
		log.Printf("Command from %s: /%s", player.Username(), command)
	}
	switch words[0] {
	case "me":
		message := strings.Join(words[1:], " ")
		log.Printf("* %s %s", player.Username(), message)
		SendToAll(protocol.Chat{Message: fmt.Sprintf("%s %s", starUsername(player), message)})
	case "pl", "players", "list", "who":
		message := make([]string, 0, len(players))
		for _, p := range players {
			message = append(message, formatUsername(p))
		}
		player.SendPacket(protocol.Chat{Message: ChatInfo + "Currently online: " + strings.Join(message, ", ")})
	default:
		player.SendPacket(protocol.Chat{Message: ChatError + "Unknown command."})
	}
}

var ops = make(map[string]bool)

func init() {
	f, err := os.Open("ops.txt")
	if err != nil {
		log.Print("While trying to load ops.txt: ", err)
		return
	}
	defer f.Close()

	r := bufio.NewReader(f)
	for {
		op, err := r.ReadString('\n')
		if op != "" {
			ops[op] = true
		}
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Print("While trying to load ops.txt: ", err)
			return
		}
	}
}

func IsOp(player Player) bool {
	return ops[player.Username()]
}

var opLock sync.Mutex
func GrantOp(player Player) bool {
	opLock.Lock()
	defer opLock.Unlock()

	if ops[player.Username()] {
		return false
	}

	ops[player.Username()] = true
	saveOps()
	return true
}

func RevokeOp(player Player) bool {
	opLock.Lock()
	defer opLock.Unlock()

	if !ops[player.Username()] {
		return false
	}

	delete(ops, player.Username())
	saveOps()
	return true
}

func saveOps() {
	f, _ := os.Create("ops.txt")
	defer f.Close()
	for op, _ := range ops {
		io.WriteString(f, op + "\n")
	}
}

func formatUsername(player Player) string {
	if isNetworkAdmin(player) {
		return ChatNameNetAdmin + player.Username() + ChatInfo
	}
	if IsOp(player) {
		return ChatNameOp + player.Username() + ChatInfo
	}
	return ChatName + player.Username() + ChatInfo
}

func bracketUsername(player Player) string {
	return ChatInfo + "<" + formatUsername(player) + ">" + ChatPayload
}

func starUsername(player Player) string {
	return ChatInfo + "* " + formatUsername(player) + ChatPayload
}
