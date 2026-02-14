# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```bash
# Build
go build -o syn .

# Install to PATH
ln -sf "$(pwd)/syn" ~/go/bin/syn

# Run tests
go test ./...

# Run single package tests
go test ./internal/app

# Build + symlink (required after code changes)
go build -o syn . && ln -sf "$(pwd)/syn" ~/go/bin/syn
```

## Architecture

CLI tool for Synthetic.new AI API (OpenAI-compatible). Built with Cobra/Viper/Lipgloss.

```
main.go                    # Entry: config.SetDefaults() → cmd.Execute()
cmd/
  root.go                  # One-shot mode, stdin support, flag handling
  chat.go                  # Interactive REPL with context (20 msg window)
  search.go                # Web search via /v2/search endpoint
  vision.go                # Image analysis via vision-capable model
  embed.go                 # Text embeddings via nomic-embed-text
  model.go                 # Model listing
  theme.go                 # Lipgloss styles + spinner
internal/
  app/
    client.go              # HTTP client with retry logic (exp backoff + jitter)
    types.go               # Request/response types, model aliases
  config/
    config.go              # Viper defaults
```

## Key Patterns

**Model aliases** (`internal/app/types.go`):
- `kimi` → `hf:moonshotai/Kimi-K2.5`
- `qwen` → `hf:Qwen/Qwen3-235B-A22B-Thinking-2507`
- `coder` → `hf:Qwen/Qwen3-Coder-480B-A35B-Instruct`
- `glm`/`zai` → `hf:zai-org/GLM-4.7`
- `gpt`/`gptoss` → `hf:openai/gpt-oss-120b`
- `deepseek`/`ds` → `hf:deepseek-ai/DeepSeek-V3.2` (default)
- `r1` → `hf:deepseek-ai/DeepSeek-R1-0528`
- `minimax` → `hf:MiniMaxAI/MiniMax-M2.1`
- `llama` → `hf:meta-llama/Llama-3.3-70B-Instruct`

**API endpoints**:
- Chat: `https://api.synthetic.new/openai/v1/chat/completions`
- Search: `https://api.synthetic.new/v2/search` (different prefix)
- Embeddings: `https://api.synthetic.new/openai/v1/embeddings`

**Configuration**: `~/.config/syn/config.yaml` or `SYN_API_KEY` env var

**Retry logic**: 3 attempts, exponential backoff (1s-30s), retries on 429/502/503/504

## Interface Segregation

Client implements multiple interfaces for testability:
- `ChatClient` - Chat()
- `ModelClient` - ListModels()
- `EmbeddingClient` - Embed()
- `VisionClient` - Vision()
- `SearchClient` - Search()
- `HTTPDoer` - Injected http.Client for mocking

## CLI Usage

```bash
syn "prompt"                    # One-shot
syn -m kimi "prompt"            # Use Kimi model
syn -f main.go "explain"        # Include file
echo "text" | syn "summarize"   # Stdin
syn --json "prompt"             # JSON output

syn chat                        # Interactive REPL
syn search "query"              # Web search
syn vision image.jpg "describe" # Image analysis
syn embed "text"                # Embeddings
syn model list                  # List models
```

## Documentation

- **README.md** - User quick-start, setup, usage examples
- **CHANGELOG.md** - Version history
- **docs/ARCHITECTURE.md** - System design, patterns, interfaces
- **docs/API.md** - Complete API reference, CLI commands, troubleshooting, examples
- **.work/** - Internal artifacts (specs, guides) - not committed to git

## Operational Context

- Eval command exists as `syn eval` wired like other Cobra subcommands in `cmd/eval.go`.
- Eval dataset loader expects paired files in one directory: `source_<id>.txt` + `gold_<id>.json`.
- Gold JSON schema for eval cases: `{ "id": "..", "title": "..", "key_insights": [..] }`.
- `syn eval --format json` / `--json` keeps stdout JSON-only; human status lines are md-mode only.
- Eval response artifact directories are timestamped to the second (`YYYYMMDD-HHMMSS`).
- Eval spec is in `.work/specs/eval-spec.md` (design doc, not user-facing).
