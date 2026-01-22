# syn

CLI for [Synthetic.new](https://synthetic.new) AI API.

## Install

```bash
go install github.com/dotcommander/syn@latest
```

## Setup

```bash
export SYN_API_KEY="your-api-key"
```

Or create `~/.config/syn/config.yaml`:

```yaml
api:
  key: "your-api-key"
  model: "hf:deepseek-ai/DeepSeek-V3.2"
```

## Usage

### Chat

```bash
# One-shot
syn "Explain goroutines"

# With file context
syn -f main.go "Review this code"

# Pipe input
pbpaste | syn "Summarize this"
cat error.log | syn "What went wrong?"

# Select model
syn -m kimi "Complex reasoning task"
syn -m qwen "Describe this architecture"

# JSON output
syn --json "List 3 facts about Go"
```

### Interactive REPL

```bash
syn chat
```

Commands: `/help`, `/clear`, `/model`, `/context`, `/exit`

### Web Search

```bash
syn search "golang error handling best practices"
syn search --json "react server components"
```

### Vision

```bash
syn vision screenshot.png "What's in this image?"
syn vision https://example.com/diagram.png "Explain this diagram"
```

### Embeddings

```bash
syn embed "Hello world"
syn embed "Text 1" "Text 2" "Text 3"
syn embed --json "For vector storage"
```

### Models

```bash
syn model list
```

## Model Aliases

| Alias | Model |
|-------|-------|
| `kimi` | Kimi-K2-Thinking |
| `qwen` | Qwen3-VL-235B (vision) |
| `glm` | GLM-4.7 |
| `gpt` | GPT-OSS-120B |
| `deepseek` | DeepSeek-V3.2 (default) |

## Flags

| Flag | Description |
|------|-------------|
| `-m, --model` | Model name or alias |
| `-f, --file` | Include file in prompt |
| `--json` | JSON output |
| `-v, --verbose` | Debug output |

## License

MIT
