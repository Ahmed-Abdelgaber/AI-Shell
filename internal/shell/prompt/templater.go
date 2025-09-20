package prompt

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func BuildShellCommand(sh, shellName, exe string) (*exec.Cmd, func(), error) {
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

# --- aish: quick git branch helper
__aish_git_branch() {
  command -v git >/dev/null 2>&1 || return 0
  git rev-parse --abbrev-ref HEAD 2>/dev/null | tr -d '\n'
}

# --- aish: JSON escaper
__aish_json_escape() {
  local s=$1
  s=${s//\\/\\\\}; s=${s//\"/\\\"}
  s=${s//$'\n'/\\n}; s=${s//$'\r'/\\r}; s=${s//$'\t'/\\t}
  printf '%s' "$s"
}

# --- aish: common logger (we’ll call it from precmd)
__aish_log_cmd() {
  local ec=$? ts cmd cwd git
  ts=$(date -Is -u 2>/dev/null || date -u "+%Y-%m-%dT%H:%M:%SZ")
  cmd=$(fc -ln -1 2>/dev/null) || return 0
  [[ -z "$cmd" ]] && return 0
  cmd="${cmd#"${cmd%%[!$' \t']*}"}"
  cwd="$PWD"
  git="$(__aish_git_branch)"
  if [[ -n "$AISH_HISTORY_FILE" ]]; then
    printf '{"ts":"%s","cwd":"%s","cmd":"%s","exit":%d,"git":"%s"}\n' \
      "$ts" "$(__aish_json_escape "$cwd")" \
      "$(__aish_json_escape "$cmd")" \
      "$ec" "$(__aish_json_escape "$git")" \
      >> "$AISH_HISTORY_FILE"
  fi
}

__aish_precmd() { __aish_log_cmd }
autoload -Uz add-zsh-hook
add-zsh-hook precmd __aish_precmd
`

	if err := os.WriteFile(rcPath, []byte(content), 0o644); err != nil {
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
set -o history
shopt -s cmdhist

if [ -z "$PS1" ]; then PS1="\u@\h:\w\$ "; fi
PS1='\[\e[1;36m\][aish]\[\e[0m\] '"$PS1"

# aish shell functions
if [ -n "$AISH_EXE" ]; then
  ai()   { "$AISH_EXE" __ai "$@"; }
  snip() { "$AISH_EXE" __snip "$@"; }
fi

# --- aish: quick git branch helper (empty if not a repo)
__aish_git_branch() {
  command -v git >/dev/null 2>&1 || return 0
  git rev-parse --abbrev-ref HEAD 2>/dev/null | tr -d '\n'
}

# --- aish: JSON escaper for safe JSON lines
__aish_json_escape() {
  local s=$1
  s=${s//\\/\\\\}; s=${s//\"/\\\"}
  s=${s//$'\n'/\\n}; s=${s//$'\r'/\\r}; s=${s//$'\t'/\\t}
  printf '%s' "$s"
}

# --- aish: log last command & exit to history.jsonl (before each prompt)
__aish_log_cmd() {
  local ec="$?"
  local ts cmd cwd git
  ts=$(date -Is -u 2>/dev/null || date -u "+%Y-%m-%dT%H:%M:%SZ")
  cmd=$(fc -ln -1 2>/dev/null) || return 0
  [ -z "$cmd" ] && return 0
  cmd="${cmd#"${cmd%%[!$' \t']*}"}"
  cwd="$PWD"
  git="$(__aish_git_branch)"

  if [ -n "$AISH_HISTORY_FILE" ]; then
    printf '{"ts":"%s","cwd":"%s","cmd":"%s","exit":%d,"git":"%s"}\n' \
      "$ts" "$(__aish_json_escape "$cwd")" \
      "$(__aish_json_escape "$cmd")" \
      "$ec" "$(__aish_json_escape "$git")" \
      >> "$AISH_HISTORY_FILE"
  fi
}

case ";$PROMPT_COMMAND;" in
  *";__aish_log_cmd;"*) ;;
  *) PROMPT_COMMAND="__aish_log_cmd; $PROMPT_COMMAND" ;;
esac
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
