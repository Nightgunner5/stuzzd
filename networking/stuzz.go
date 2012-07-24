package networking

// StuzzHosting-specific functions.

func isNetworkAdmin(player Player) bool {
	switch player.Username() {
	case "7031", "Nightgunner5", "L4ppy1337":
		return true
	}
	return false
}

func customLoginMessage(player Player) string {
	switch player.Username() {
	case "L4ppy1337":
		return "§7A wild %s appears!"
	}
	return ""
}
