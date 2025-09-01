package shell

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func buildShellCommand(sh, shellName, exe string) (*exec.Cmd, func(), error) {
	aishEnv := []string{"AISH=1", "AISH_EXE=" + exe}

	switch shellName {
	case "bash":
		rcPath, cleanup, err := makeTempBashRC()
		if err != nil {
			return nil, nil, fmt.Errorf("make temp bash rc: %w", err)
		}
		cmd := exec.Command(sh, "--rcfile", rcPath, "-i")
		cmd.Env = append(os.Environ(), aishEnv...)
		return cmd, cleanup, nil

	case "zsh":
		zdot, cleanup, err := makeTempZshDir()
		if err != nil {
			return nil, nil, fmt.Errorf("make temp zsh rc: %w", err)
		}
		cmd := exec.Command(sh, "-i")
		cmd.Env = append(os.Environ(), append(aishEnv, "ZDOTDIR="+zdot)...)
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
	return strings.ReplaceAll(s, `"`, `\"`)
}

func makeTempZshDir() (string, func(), error) {
	home, _ := os.UserHomeDir()
	userRC := filepath.Join(home, ".zshrc")
	dir, err := os.MkdirTemp("", "aish-zdotdir-*")
	if err != nil {
		return "", nil, err
	}
	rcPath := filepath.Join(dir, ".zshrc")
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
	if err := os.WriteFile(rcPath, []byte(content), 0644); err != nil {
		_ = os.RemoveAll(dir)
		return "", nil, err
	}
	cleanup := func() { _ = os.RemoveAll(dir) }
	return dir, cleanup, nil
}

func makeTempBashRC() (string, func(), error) {
	home, _ := os.UserHomeDir()
	userRC := filepath.Join(home, ".bashrc")

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
