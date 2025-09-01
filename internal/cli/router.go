package cli

func RouteCommand(args []string) bool {
	if len(args) > 1 {
		switch args[1] {
		case "__ai":
			handleAI(args[2:])
			return true
		case "__snip":
			handleSnip(args[2:])
			return true
		}
	}
	return false
}
