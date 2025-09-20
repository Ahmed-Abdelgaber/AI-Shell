# AI Shell (aish)

`aish` is a Go-based interactive shell that layers AI assistance and reusable command snippets on top of your normal terminal workflow. It provides:

- **Interactive shell bootstrap** that wraps your preferred shell (`bash`, `zsh`) while logging history, session transcripts, and snippet storage in `~/.aish`.
- **AI helpers** (`ai ask`, `ai why`, `ai fix`) that can analyse recent session context and run suggested commands with safety checks.
- **Snippet management** (`snip add`, `snip run`, `snip view`, `snip delete`) for saving command sequences with templated variables and replaying them later.
- **Consistent, colour-aware error handling** via a shared printer so warnings and failures stand out regardless of which component raises them.

## Getting Started

### Prerequisites

- Go 1.23 or newer (the module targets Go 1.23 with the Go 1.24 toolchain).
- An AI provider key (`OPENAI_API_KEY`) if you intend to use the AI commands. OpenAI is the default provider at the moment.

### Build

```bash
# Build the CLI executable
go build -o aish ./cmd/aish
```

### Run

```bash
# Launches the interactive shell wrapper
./aish
```

On first launch the app creates `~/.aish/<session-id>/` to hold the session log (`session.log`) and history (`history.jsonl`), plus a shared `~/.aish/snippets.yaml` database for snippets.

## Environment Variables

| Variable                     | Purpose                                             | Default                 |
| ---------------------------- | --------------------------------------------------- | ----------------------- |
| `AI_PROVIDER`                | Selects the AI backend (`openai`, `ollama`).        | `openai`                |
| `OPENAI_API_KEY`             | API key for OpenAI when `AI_PROVIDER=openai`.       | _required for AI_       |
| `AISH_SNIPPETS_FILE`         | Path to the snippets YAML store.                    | `~/.aish/snippets.yaml` |
| `AISH_SESSION_LOG`           | File used for tailing recent output in AI context.  | auto-filled per session |
| `AISH_HISTORY_FILE`          | JSONL history file used for AI context.             | auto-filled per session |
| `AISH_TAIL_LINES`            | Number of lines to read from history/log files.     | `120`                   |
| `AISH_TAIL_MAX_BYTES`        | Byte limit for tail operations.                     | `256 << 10`             |
| `AISH_HISTORY_SIZE`          | Multiplier for history lines when building context. | `5`                     |
| `AISH_NO_COLOR` / `NO_COLOR` | Disable colour output in the shared printer.        | unset                   |

You usually only need to set `OPENAI_API_KEY`. The other variables are managed automatically by the shell launcher.

## CLI Usage

Once `aish` starts it injects shell functions (`ai`, `snip`) that delegate to the executable, so you can run the commands directly from the interactive prompt. Outside the shell you can call the internal commands with `aish __ai ...` or `aish __snip ...`.

### AI Commands

- `ai ask <question> [-c|--context]` &mdash; Ask a question. The optional `-c` flag attaches recent shell history/log output.
- `ai why` &mdash; Explain why the last command failed based on recent history.
- `ai fix` &mdash; Request a single safe fix command and run it after confirmation. Dangerous commands and non-persistent builtins are rejected.

### Snippet Commands

- `snip add <name> <command...>` &mdash; Store a snippet. Commands containing `[[variable]]` placeholders register required variables automatically.
- `snip run <name> [var=value ...]` &mdash; Execute a saved snippet, prompting for confirmation. Variables are substituted before each step, and colourised errors highlight failures or missing variables.
- `snip view <name>` &mdash; Inspect the stored steps and metadata for a snippet.
- `snip ls` &mdash; List stored snippets.
- `snip delete <name>` &mdash; Remove a snippet from the YAML store.

Snippet steps can be stored either as raw shell strings (`Cmd`) or exec arrays (`Exec`). The runner streams output/interactive prompts (e.g., `sudo`) directly through to your terminal.

## Error Handling & Output

Errors that flow through `internal/errs` carry codes, severities, and metadata. The shared printer (`internal/ux/printer`) renders them consistently:

- **Errors** appear in red with the code and message.
- **Warnings** and **info** lines use contrasting colour/faint output where supported.
- Metadata fields (e.g., the snippet line that failed) are indented and dimmed for quick scanning.
- Colour can be disabled globally with `NO_COLOR` or `AISH_NO_COLOR`.

Any component can opt into richer errors by wrapping failures with `errs.Wrap(...)` and optional fields.

## Project Layout

```
cmd/aish/            # CLI entrypoint
internal/app/        # Application bootstrap (config + router + launcher)
internal/cli/        # Command handlers for AI and snippets
internal/errs/       # Error metadata helpers
internal/ux/printer/ # Colour-aware printer
internal/session/    # Session context builders (history/logs, redaction)
internal/shell/      # Shell launcher, PTY wiring, prompt templates
internal/snippets/   # Snippet parser, service, and YAML store
internal/ai/         # AI service, providers, prompt templates
```
