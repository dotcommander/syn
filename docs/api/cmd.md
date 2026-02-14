# cmd Package

CLI commands and application entry point for the syn tool.

## Commands

### root
```bash
syn [prompt]
```
One-shot mode: send a prompt and get a response.

**Examples:**
```bash
syn "Explain quantum computing"
syn -f main.go "Review this code"
echo "text" | syn "summarize"
```

**Flags:**
- `-m, --model <name>` - Model to use (aliases: kimi, glm, qwen, gpt)
- `-f, --file <path>` - Include file contents in prompt
- `--json` - Output as JSON
- `-v, --verbose` - Show debug info
- `-h, --help` - Show help

### chat
```bash
syn chat
```
Interactive chat session (REPL mode).

See [chat command documentation](#) for details.

### vision
```bash
syn vision <image> <prompt>
```
Analyze images with AI.

See [vision command documentation](#) for details.

### search
```bash
syn search <query>
```
Search the web using Synthetic's /v2/search endpoint.

**Note:** This API is under development. Zero-data-retention policy applies.

**Examples:**
```bash
syn search "golang error handling"
syn search --json "react hooks"
echo "python async" | syn search
```

### embed
```bash
syn embed <text...>
```
Generate text embeddings.

See [embed command documentation](#) for details.

### model
```bash
syn model [list|info]
```
Model management commands.

See [model command documentation](#) for details.

## Configuration

Configuration is loaded from (in order of precedence):
1. Command-line flags
2. Environment variables (`SYN_API_KEY`, `SYN_API_BASE_URL`, etc.)
3. Config file (`~/.config/syn/config.yaml`)

### Config File Example

```yaml
api:
  key: syn_your_api_key_here
  base_url: https://api.synthetic.new/openai/v1
  anthropic_base_url: https://api.synthetic.new/anthropic/v1
  model: hf:deepseek-ai/DeepSeek-V3.2
  embedding_model: hf:nomic-ai/nomic-embed-text-v1.5
  retry:
    max_attempts: 3
    initial_backoff: 1s
    max_backoff: 30s

chat:
  temperature: 0.6
  max_tokens: 8192
  top_p: 0.9
```

## Functions

### Execute
```go
func Execute()
```
Entry point that executes the root command tree.

### styledHelp
```go
func styledHelp(cmd *cobra.Command, args []string)
```
Custom help formatter with styled output.

## Environment Variables

| Variable | Description |
|----------|-------------|
| `SYN_API_KEY` | API key (required) |
| `SYN_API_BASE_URL` | OpenAI-compatible API base URL |
| `SYN_API_ANTHROPIC_BASE_URL` | Anthropic-compatible API base URL |
| `SYN_API_MODEL` | Default model |
| `SYN_API_EMBEDDING_MODEL` | Default embedding model |
