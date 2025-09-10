package cli

func RouteCommand(args []string) bool {
	// Check if the command is an internal command.
	// Internal commands start with "__".
	// If it is an internal command, handle it accordingly and return true.
	// If not, return false to indicate that it is not handled here.
	if len(args) > 1 {
		// Switch on the specific internal command.
		switch args[1] {
		case "__ai":
			// Handle AI-related commands.
			handleAI(args[2:])
			return true
		case "__snip":
			// Handle snippet-related commands.
			handleSnip(args[2:])
			return true
		}
	}
	// Not an internal command.
	return false
}
