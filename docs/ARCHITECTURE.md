# Architecture

System design and patterns for the Syn CLI tool.

## Overview

Syn is a CLI tool for the Synthetic.new AI API, providing an OpenAI-compatible interface with additional features like web search, vision analysis, and embeddings.

**Built with:**
- Cobra (CLI framework)
- Viper (configuration management)
- Lipgloss (terminal styling)

## Directory Structure

```
main.go                    # Entry: config.SetDefaults() → cmd.Execute()
cmd/
  root.go                  # One-shot mode, stdin support, flag handling
  chat.go                  # Interactive REPL with context (20 msg window)
  search.go                # Web search via /v2/search endpoint
  vision.go                # Image analysis via Qwen3-VL
  embed.go                 # Text embeddings via nomic-embed-text
  eval.go                  # Model evaluation framework
  model.go                 # Model listing
  theme.go                 # Lipgloss styles + spinner
internal/
  app/
    client.go              # HTTP client with retry logic (exp backoff + jitter)
    types.go               # Request/response types, model aliases
  config/
    config.go              # Viper defaults
```

## Core Patterns

### Model Aliases

Defined in `internal/app/types.go`:

| Alias | Full Model Name |
|-------|-----------------|
| `kimi` | `hf:moonshotai/Kimi-K2-Thinking` |
| `qwen` | `hf:Qwen/Qwen3-VL-235B-A22B-Instruct` (vision) |
| `glm`, `zai` | `hf:zai-org/GLM-4.7` |
| `gpt`, `gptoss` | `hf:openai/gpt-oss-120b` |
| `deepseek`, `ds` | `hf:deepseek-ai/DeepSeek-V3.2` (default) |

### API Endpoints

- **Chat:** `https://api.synthetic.new/openai/v1/chat/completions`
- **Search:** `https://api.synthetic.new/v2/search` (different prefix)
- **Embeddings:** `https://api.synthetic.new/openai/v1/embeddings`

### Configuration

Configuration loaded from (in order of precedence):

1. Command-line flags
2. Environment variables (`SYN_API_KEY`, `SYN_API_BASE_URL`, etc.)
3. Config file (`~/.config/syn/config.yaml`)

### Retry Logic

Automatic retry for transient failures (429/502/503/504):

- **Max attempts:** 3
- **Exponential backoff:** 1s-30s
- **Jitter:** Randomized delay to avoid thundering herd

## Interface Segregation

The client implements multiple focused interfaces for testability:

```go
type ChatClient interface {
    Chat(ctx context.Context, prompt string, opts ChatOptions) (string, error)
}

type ModelClient interface {
    ListModels(ctx context.Context) ([]Model, error)
}

type EmbeddingClient interface {
    Embed(ctx context.Context, texts []string, model string) (*EmbeddingResponse, error)
}

type VisionClient interface {
    Vision(ctx context.Context, prompt string, imageSource string, opts ChatOptions) (string, error)
}

type SearchClient interface {
    Search(ctx context.Context, query string) (*SearchResponse, error)
}

type HTTPDoer interface {
    Do(*http.Request) (*http.Response, error)
}
```

**Benefits:**
- Easy to mock in tests
- Single responsibility per interface
- Explicit dependencies

## Dependency Injection

The client uses constructor injection for testability:

```go
func NewClient(cfg ClientConfig, logger *slog.Logger, httpClient HTTPDoer) *Client
```

**Injected dependencies:**
- `logger` - Structured logging
- `httpClient` - HTTP transport (allows mocking)

## Conversational Context

The chat command maintains a sliding window of 20 messages:

- User messages and assistant responses stored in memory
- Context passed to API for multi-turn conversations
- Cleared with `/clear` command

## Error Handling

All errors are wrapped with context:

```go
if err != nil {
    return fmt.Errorf("failed to read file: %w", err)
}
```

API errors include status code and response body:

```go
type APIError struct {
    StatusCode int
    Body       string
}
```

## Testing Strategy

- **Table-driven tests** for multiple cases
- **Interface mocking** for HTTP client
- **Dependency injection** for logger and HTTP client
- **Edge cases:** empty inputs, nil values, errors
