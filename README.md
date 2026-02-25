# omoc

A TUI tool for configuring [Oh My OpenCode](https://github.com/code-yeongyu/oh-my-opencode) agent and category model assignments.

Reads and writes `~/.config/opencode/oh-my-opencode.json`. Available models are fetched from `opencode models`.

## Install

### Homebrew

```bash
brew tap fingergohappy/tap
brew install omoc
```

### From source

```bash
go install .
```

Or build locally:

```bash
go build -o omoc .
```

## Usage

```bash
./omoc
```

### Keybindings

| Key | Action |
|-----|--------|
| `j` / `k` | Move up / down |
| `h` / `l` | Switch to left / right panel |
| `tab` | Toggle between panels |
| `enter` | Assign selected model to current agent/category |
| `v` | Cycle variant (low → high → xhigh → max → none) |
| `d` | Clear model and variant for current item |
| `/` | Filter models (in model panel) |
| `esc` | Clear filter |
| `r` | Refresh model list |
| `s` | Save config |
| `q` | Quit |

### Layout

- Left panel: agents and categories with current model assignments
- Middle panel: available models (current model pinned to top with ★)
- Right panel: description, fallback chain, and notes from the [agent-model matching guide](https://github.com/code-yeongyu/oh-my-opencode/blob/dev/docs/guide/agent-model-matching.md)

A timestamped backup is created before each save.

## Requirements

- Go 1.21+
- `opencode` CLI installed and authenticated
