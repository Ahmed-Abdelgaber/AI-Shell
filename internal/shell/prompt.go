package shell

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func buildShellCommand(sh, shellName, exe string) (*exec.Cmd, func(), error) {
	// Prepare the environment variables to indicate AISH mode.
	// AISH=1 indicates that the shell is running in AISH mode.
	// AISH_EXE is the path to the AISH executable, used by shell functions.
	aishEnv := []string{"AISH=1", "AISH_EXE=" + exe}

	switch shellName {
	case "bash":
		// Create a temporary bash rc file that sources the user's .bashrc
		rcPath, cleanup, err := makeTempBashRC()
		if err != nil {
			return nil, nil, fmt.Errorf("make temp bash rc: %w", err)
		}
		// Start bash with the temporary rc file and interactive mode.
		// The --rcfile option tells bash to use the specified rc file instead of the default.
		// The -i option tells bash to run in interactive mode.
		// We also append the AISH environment variables to the command's environment.
		cmd := exec.Command(sh, "--rcfile", rcPath, "-i")
		cmd.Env = append(os.Environ(), aishEnv...)
		// Return the command, cleanup function, and no error.
		return cmd, cleanup, nil

	case "zsh":
		// Create a temporary directory to act as ZDOTDIR, containing a .zshrc
		zdot, cleanup, err := makeTempZshDir()
		if err != nil {
			return nil, nil, fmt.Errorf("make temp zsh rc: %w", err)
		}
		// Start zsh with the temporary ZDOTDIR and interactive mode.
		// Setting ZDOTDIR to the temporary directory makes zsh look for its .zshrc there.
		// The -i option tells zsh to run in interactive mode.
		// We also append the AISH environment variables to the command's environment.
		cmd := exec.Command(sh, "-i")
		cmd.Env = append(os.Environ(), append(aishEnv, "ZDOTDIR="+zdot)...)
		// Return the command, cleanup function, and no error.
		return cmd, cleanup, nil

	default:
		// Fallback: try to tag PS1 via env (some shells ignore)
		existingPS1 := os.Getenv("PS1")
		if existingPS1 == "" {
			existingPS1 = "\\u@\\h:\\w\\$ "
		}
		aishTag := "\\[\\e[1;36m\\][aish]\\[\\e[0m\\] "
		cmd := exec.Command(sh)
		cmd.Env = append(os.Environ(), append(aishEnv, "PS1="+aishTag+existingPS1)...)
		return cmd, nil, nil
	}
}

func escapeShell(s string) string {
	// Escape double quotes for safe inclusion in shell scripts.
	return strings.ReplaceAll(s, `"`, `\"`)
}

func makeTempZshDir() (string, func(), error) {
	// Get the user's home directory to locate their .zshrc
	home, _ := os.UserHomeDir()
	// Path to the user's .zshrc
	userRC := filepath.Join(home, ".zshrc")
	// Create a temporary directory to act as ZDOTDIR
	// The directory will be automatically cleaned up later.
	// We use os.MkdirTemp to create a unique temporary directory.
	// The pattern "aish-zdotdir-*" helps identify the purpose of the directory.
	dir, err := os.MkdirTemp("", "aish-zdotdir-*")
	if err != nil {
		return "", nil, err
	}
	// Path to the temporary .zshrc inside the temporary directory
	rcPath := filepath.Join(dir, ".zshrc")
	// Content of the temporary .zshrc
	// It sources the user's .zshrc if it exists, sets the AISH environment variable,
	// modifies the prompt to include an [aish] tag, and defines shell functions for AI interaction.
	// We use escapeShell to safely include the path to the user's .zshrc.
	// The PROMPT variable is modified to include the [aish] tag in cyan and bold.
	// The shell functions `ai` and `snip` call the AISH executable with appropriate arguments.
	// We check if the AISH_EXE environment variable is set before defining the functions.
	// This ensures that the functions only exist when AISH is properly set up.
	// Finally, we write the content to the temporary .zshrc file.
	// If writing fails, we clean up the temporary directory and return an error.
	// On success, we return the path to the temporary directory, a cleanup function, and no error.
	content := `# aish: temporary zsh rc (auto-generated)
if [ -f "` + escapeShell(userRC) + `" ]; then
  source "` + escapeShell(userRC) + `"
fi
export AISH=1
PROMPT="%B%F{cyan}[aish]%f%b ${PROMPT}"
# aish shell functions
if [[ -n "$AISH_EXE" ]]; then
  function ai()   { "$AISH_EXE" __ai "$@"; }
  function snip() { "$AISH_EXE" __snip "$@"; }
fi
`
	// Write the content to the temporary .zshrc file
	if err := os.WriteFile(rcPath, []byte(content), 0644); err != nil {
		_ = os.RemoveAll(dir)
		return "", nil, err
	}
	// Cleanup function to remove the temporary directory
	cleanup := func() { _ = os.RemoveAll(dir) }
	// Return the path to the temporary directory, cleanup function, and no error
	return dir, cleanup, nil
}

func makeTempBashRC() (string, func(), error) {
	// Get the user's home directory to locate their .bashrc
	home, _ := os.UserHomeDir()
	// Path to the user's .bashrc
	userRC := filepath.Join(home, ".bashrc")
	// Content of the temporary bash rc file
	// It sources the user's .bashrc if it exists, sets the AISH environment variable,
	// modifies the prompt to include an [aish] tag, and defines shell functions for AI interaction.
	// We use escapeShell to safely include the path to the user's .bashrc.
	// The PS1 variable is modified to include the [aish] tag in cyan and bold.
	// The shell functions `ai` and `snip` call the AISH executable with appropriate arguments.
	// We check if the AISH_EXE environment variable is set before defining the functions.
	// This ensures that the functions only exist when AISH is properly set up.
	// We create a temporary file to hold this content using os.CreateTemp.
	// The pattern "aish-bashrc-*" helps identify the purpose of the file.
	// If creating or writing to the file fails, we clean up and return an error.
	// On success
	content := `# aish: temporary bash rc (auto-generated)
if [ -f "` + escapeShell(userRC) + `" ]; then
  . "` + escapeShell(userRC) + `"
fi
export AISH=1
if [ -z "$PS1" ]; then PS1="\u@\h:\w\$ "; fi
PS1='\[\e[1;36m\][aish]\[\e[0m\] '"$PS1"
# aish shell functions
if [ -n "$AISH_EXE" ]; then
  ai()   { "$AISH_EXE" __ai "$@"; }
  snip() { "$AISH_EXE" __snip "$@"; }
fi
`
	f, err := os.CreateTemp("", "aish-bashrc-*")
	if err != nil {
		return "", nil, err
	}
	if _, err := f.WriteString(content); err != nil {
		_ = f.Close()
		_ = os.Remove(f.Name())
		return "", nil, err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(f.Name())
		return "", nil, err
	}
	cleanup := func() { _ = os.Remove(f.Name()) }
	return f.Name(), cleanup, nil
}
