package shell

import (
	"bufio"
	"os"
	"strings"
)

func ConfirmFromStdin() bool {
	in := bufio.NewReader(os.Stdin)
	line, _ := in.ReadString('\n')
	line = strings.ToLower(strings.TrimSpace(line))
	return line == "y" || line == "yes"
}
