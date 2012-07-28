package networking

import (
	"bufio"
	"fmt"
	"github.com/Nightgunner5/stuzzd/config"
	"github.com/Nightgunner5/stuzzd/protocol"
	"io"
	"log"
	"os"
	"strings"
	"sync"
)

const (
	ChatInfo         = "§7"
	ChatError        = "§c"
	ChatPayload      = "§r"
	ChatName         = "§6§l"
	ChatNameOp       = "§4§l"
	ChatNameNetAdmin = "§3§l"
	ChatNotAllowed   = ChatError + "You do not have the required permission to use that command."
)

var CommandHelp = map[string]struct {
	Description string
	OpOnly      bool
}{
	"help":  {Description: "This command.", OpOnly: false},
	"me":    {Description: "Describe an action, like \"/me eats a cupcake.\"", OpOnly: false},
	"who":   {Description: "List the players currently online.", OpOnly: false},
	"op":    {Description: "Give a player operator status.", OpOnly: true},
	"deop":  {Description: "Revoke a player's operator status.", OpOnly: true},
	"kick":  {Description: "Kick a player from the server with an optional message.", OpOnly: true},
	"tpt":   {Description: "Teleport yourself to a player.", OpOnly: true},
	"gm":    {Description: "Set a player's game mode to creative (c) or survival (s).", OpOnly: true},
	"day":   {Description: "Advance to the next day.", OpOnly: true},
	"night": {Description: "Advance to the next night.", OpOnly: true},
}

func handleCommand(player Player, command string) {
	defer func() { recover() }()
	words := strings.Split(command, " ")
	switch words[0] {
	case "me", "pl", "players", "list", "who", "help":
		// Don't spam the log.
	default:
		log.Printf("Command from %s: /%s", player.Username(), command)
	}
	switch words[0] {
	case "me":
		if len(words) < 2 {
			return
		}
		message := strings.Join(words[1:], " ")
		log.Printf("* %s %s", player.Username(), message)
		SendToAll(protocol.Chat{Message: fmt.Sprintf("%s %s", starUsername(player), message)})
	case "pl", "players", "list", "who":
		message := make([]string, 0, len(players))
		for _, p := range players {
			if p.Authenticated() {
				message = append(message, formatUsername(p))
			}
		}
		sendChat(player, ChatInfo+"Currently online: "+strings.Join(message, ", "))
	case "help":
		player.SendPacketSync(protocol.Chat{Message: ChatInfo + "=== " + ChatPayload + "Help" + ChatInfo + " ==="})
		for command, info := range CommandHelp {
			if info.OpOnly {
				if !IsOp(player) && !isNetworkAdmin(player) {
					continue
				}
				sendChat(player, ChatNameOp+command+ChatInfo+" - "+ChatPayload+info.Description)
			} else {
				sendChat(player, ChatName+command+ChatInfo+" - "+ChatPayload+info.Description)
			}
		}
	case "op":
		if !checkOp(player) {
			return
		}

		var target Player
		for _, p := range players {
			if p.Authenticated() && p.Username() == words[1] {
				target = p
				break
			}
		}
		if target == nil {
			sendChat(player, ChatError+"Could not find target.")
			return
		}

		if GrantOp(target) {
			SendToAll(protocol.Chat{Message: fmt.Sprintf("%s has been given Operator privileges by %s.", formatUsername(target), formatUsername(player))})
		} else {
			sendChat(player, ChatError+"Target already has Op!")
		}
	case "deop", "unop":
		if !checkOp(player) {
			return
		}

		var target Player
		for _, p := range players {
			if p.Authenticated() && p.Username() == words[1] {
				target = p
				break
			}
		}
		if target == nil {
			sendChat(player, ChatError+"Could not find target.")
			return
		}

		if RevokeOp(target) {
			SendToAll(protocol.Chat{Message: fmt.Sprintf("%s has had their Operator privileges revoked by %s.", formatUsername(target), formatUsername(player))})
		} else {
			sendChat(player, ChatError+"Target doesn't have Op!")
		}
	case "kick":
		if !checkOp(player) {
			return
		}

		var target Player
		for _, p := range players {
			if p.Authenticated() && p.Username() == words[1] {
				target = p
				break
			}
		}
		if target == nil {
			sendChat(player, ChatError+"Could not find target.")
			return
		}

		message := "No reason given"
		if len(words) > 2 {
			message = "\"" + strings.Join(words[2:], " ") + "\""
		}

		go target.SendPacketSync(protocol.Kick{Reason: "Kicked by admin: " + message})

	case "tpt":
		if !checkOp(player) {
			return
		}

		var target Player
		for _, p := range players {
			if p.Authenticated() && p.Username() == words[1] {
				target = p
				break
			}
		}
		if target == nil {
			sendChat(player, ChatError+"Could not find target.")
			return
		}

		player.SetPosition(target.Position())
		player.ForcePosition()

	case "gm", "gamemode":
		if !checkOp(player) {
			return
		}

		var target Player
		for _, p := range players {
			if p.Authenticated() && p.Username() == words[1] {
				target = p
				break
			}
		}
		if target == nil {
			sendChat(player, ChatError+"Could not find target.")
			return
		}

		switch words[2] {
		case "creative", "c", "1":
			target.SetGameMode(protocol.Creative)
		case "survival", "s", "0":
			target.SetGameMode(protocol.Survival)
		default:
			sendChat(player, ChatError+"Unknown game mode.")
			return
		}

		sendChat(player, ChatInfo+"You have changed "+formatUsername(target)+"'s game mode.")
		sendChat(target, formatUsername(player)+" has changed your game mode.")

	case "day":
		if !checkOp(player) {
			return
		}

		config.Tick += 24000 - (config.Tick % 24000)

		SendToAll(protocol.TimeUpdate{Time: config.Tick})

	case "night":
		if !checkOp(player) {
			return
		}

		if config.Tick%24000 < 18000 {
			config.Tick += 18000 - (config.Tick % 24000)
		} else {
			config.Tick += 24000 + 18000 - (config.Tick % 24000)
		}

		SendToAll(protocol.TimeUpdate{Time: config.Tick})

	default:
		sendChat(player, ChatError+"Unknown command.")
	}
}

func checkOp(player Player) bool {
	if !IsOp(player) {
		if isNetworkAdmin(player) {
			sendChat(player, ChatInfo+"Using network admin override...")
		} else {
			sendChat(player, ChatNotAllowed)
			return false
		}
	}
	return true
}

func sendChat(player Player, message string) {
	player.SendPacketSync(protocol.Chat{Message: message})
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
		op = strings.TrimSpace(op)
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
		io.WriteString(f, op+"\n")
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
