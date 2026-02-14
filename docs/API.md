# API Reference

Complete reference for the Syn CLI and client library.

## Client Library

### Interfaces

#### ChatClient

```go
type ChatClient interface {
    Chat(ctx context.Context, prompt string, opts ChatOptions) (string, error)
}
```

Interface for chat completion operations.

#### ModelClient

```go
type ModelClient interface {
    ListModels(ctx context.Context) ([]Model, error)
}
```

Interface for listing available models.

#### EmbeddingClient

```go
type EmbeddingClient interface {
    Embed(ctx context.Context, texts []string, model string) (*EmbeddingResponse, error)
}
```

Interface for generating text embeddings.

#### VisionClient

```go
type VisionClient interface {
    Vision(ctx context.Context, prompt string, imageSource string, opts ChatOptions) (string, error)
}
```

Interface for vision/image analysis operations.

#### SearchClient

```go
type SearchClient interface {
    Search(ctx context.Context, query string) (*SearchResponse, error)
}
```

Interface for web search operations. Uses `/v2/search` endpoint with zero-data-retention.

### Types

#### Client

```go
type Client struct {
    // contains filtered or exported fields
}
```

Main client implementing all API interfaces. Create with `NewClient()`.

#### ClientConfig

```go
type ClientConfig struct {
    APIKey         string
    BaseURL        string
    AnthropicURL   string
    Model          string
    EmbeddingModel string
    Timeout        time.Duration
    Verbose        bool
    RetryConfig    RetryConfig
}
```

Configuration for the API client.

#### RetryConfig

```go
type RetryConfig struct {
    MaxAttempts    int           // Default: 3
    InitialBackoff time.Duration // Default: 1s
    MaxBackoff     time.Duration // Default: 30s
}
```

Retry behavior configuration for transient failures.

#### Message

```go
type Message struct {
    Role    string // "user", "assistant", "system"
    Content string
}
```

Chat message representation.

#### ChatOptions

```go
type ChatOptions struct {
    Model       string
    Temperature *float64
    MaxTokens   *int
    TopP        *float64
    FilePath    string
    Context     []Message
}
```

Options for chat requests.

#### Usage

```go
type Usage struct {
    PromptTokens     int
    CompletionTokens int
    TotalTokens      int
}
```

Token usage statistics.

#### Model

```go
type Model struct {
    ID      string
    Object  string
    Created int64
    OwnedBy string
}
```

Model information.

#### APIError

```go
type APIError struct {
    StatusCode int
    Body       string
}
```

API error response.

#### SearchResponse

```go
type SearchResponse struct {
    Results []SearchResult
}
```

Response from the /v2/search endpoint.

#### SearchResult

```go
type SearchResult struct {
    Title     string
    URL       string
    Snippet   string
    Published string // ISO 8601 timestamp (optional)
}
```

Individual search result with metadata.

### Functions

#### NewClient

```go
func NewClient(cfg ClientConfig, logger *slog.Logger, httpClient HTTPDoer) *Client
```

Creates a new API client with injected dependencies.

#### NewLogger

```go
func NewLogger(verbose bool) *slog.Logger
```

Creates a structured logger for the application.

#### ResolveModel

```go
func ResolveModel(model string) string
```

Resolves model aliases (kimi, qwen, coder, glm, gpt, deepseek, r1, minimax, llama) to full model names.

#### DefaultChatOptions

```go
func DefaultChatOptions() ChatOptions
```

Returns sensible default options for chat requests.

### Methods

#### (*Client).Chat

```go
func (c *Client) Chat(ctx context.Context, prompt string, opts ChatOptions) (string, error)
```

Sends a chat prompt and returns the response. Supports file inclusion via `opts.FilePath`.

#### (*Client).ListModels

```go
func (c *Client) ListModels(ctx context.Context) ([]Model, error)
```

Fetches available models from the API.

#### (*Client).Embed

```go
func (c *Client) Embed(ctx context.Context, texts []string, model string) (*EmbeddingResponse, error)
```

Generates embeddings for the given texts.

#### (*Client).Vision

```go
func (c *Client) Vision(ctx context.Context, prompt string, imageSource string, opts ChatOptions) (string, error)
```

Analyzes an image with a prompt. `imageSource` can be a URL or local file path.

#### (*Client).Search

```go
func (c *Client) Search(ctx context.Context, query string) (*SearchResponse, error)
```

Performs a web search. Returns results with title, URL, snippet, and optional published timestamp.

### Model Aliases

| Alias | Full Model Name |
|-------|-----------------|
| kimi | hf:moonshotai/Kimi-K2.5 |
| qwen | hf:Qwen/Qwen3-235B-A22B-Thinking-2507 |
| coder | hf:Qwen/Qwen3-Coder-480B-A35B-Instruct |
| glm, zai | hf:zai-org/GLM-4.7 |
| gpt, gptoss | hf:openai/gpt-oss-120b |
| deepseek, ds | hf:deepseek-ai/DeepSeek-V3.2 |
| r1 | hf:deepseek-ai/DeepSeek-R1-0528 |
| minimax | hf:MiniMaxAI/MiniMax-M2.1 |
| llama | hf:meta-llama/Llama-3.3-70B-Instruct |

## CLI Commands

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

**Commands:**

- `/help` - Show available commands
- `/clear` - Clear conversation history
- `/model` - Show current model
- `/context` - Show conversation context
- `/exit` - Exit chat session

### vision

```bash
syn vision <image> <prompt>
```

Analyze images with AI.

**Examples:**

```bash
syn vision screenshot.png "What's in this image?"
syn vision https://example.com/diagram.png "Explain this diagram"
```

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

**Examples:**

```bash
syn embed "Hello world"
syn embed "Text 1" "Text 2" "Text 3"
syn embed --json "For vector storage"
```

### model

```bash
syn model [list|info]
```

Model management commands.

**Examples:**

```bash
syn model list
```

### eval

```bash
syn eval --dataset <path>
```

Run insight-extraction evaluation across models.

**Examples:**

```bash
# Run eval across all listed models
syn eval --dataset testdata/eval/walter_lewin

# Save machine-readable report
syn eval --format json --out analysis-results/eval-report.json

# Evaluate only selected models
syn eval --models "hf:deepseek-ai/DeepSeek-V3.2,hf:moonshotai/Kimi-K2.5"

# Keep persistent score history + leaderboard
syn eval --history analysis-results/eval-history.jsonl --leaderboard-out analysis-results/eval-leaderboard.md
```

## Configuration

### Config File Location

`~/.config/syn/config.yaml`

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

### Default Values

#### API Defaults

| Setting | Default Value |
|---------|---------------|
| `api.key` | *(must be set by user via `SYN_API_KEY` or config file)* |
| `api.base_url` | https://api.synthetic.new/openai/v1 |
| `api.anthropic_base_url` | https://api.synthetic.new/anthropic/v1 |
| `api.model` | hf:deepseek-ai/DeepSeek-V3.2 |
| `api.embedding_model` | hf:nomic-ai/nomic-embed-text-v1.5 |

#### Retry Configuration

| Setting | Default Value |
|---------|---------------|
| `api.retry.max_attempts` | 3 |
| `api.retry.initial_backoff` | 1s |
| `api.retry.max_backoff` | 30s |

#### Chat Defaults

| Setting | Default Value |
|---------|---------------|
| `chat.temperature` | 0.6 |
| `chat.max_tokens` | 8192 |
| `chat.top_p` | 0.9 |

### Environment Variables

| Variable | Description |
|----------|-------------|
| `SYN_API_KEY` | API key (required) |
| `SYN_API_BASE_URL` | OpenAI-compatible API base URL |
| `SYN_API_ANTHROPIC_BASE_URL` | Anthropic-compatible API base URL |
| `SYN_API_MODEL` | Default model |
| `SYN_API_EMBEDDING_MODEL` | Default embedding model |

## Troubleshooting

### API key is not configured

**Error Message:**

```text
API key is not configured. Set SYN_API_KEY or configure in ~/.config/syn/config.yaml
```

**Cause:**

The `SYN_API_KEY` environment variable is not set or no API key is configured in the config file.

**Resolution:**

1. Set the `SYN_API_KEY` environment variable:

   ```bash
   export SYN_API_KEY=your_api_key_here
   ```

2. Or configure it in `~/.config/syn/config.yaml`:

   ```yaml
   api_key: your_api_key_here
   ```

### failed to read file

**Error Message:**

```text
failed to read file /path/to/file: open /path/to/file: no such file or directory
```

**Cause:**

The specified file path does not exist or is not readable.

**Resolution:**

1. Verify the file path is correct
2. Check file permissions: `ls -la /path/to/file`
3. Use absolute paths if relative paths are not resolving correctly

### API error: 401

**Error Message:**

```text
API error: 401 - {"error":"Unauthorized"}
```

**Cause:**

The API key is invalid, expired, or malformed.

**Resolution:**

1. Verify your API key is correct
2. Check if the key has expired
3. Ensure there are no extra spaces or characters in the key
4. Regenerate the API key if necessary

### API error: 429

**Error Message:**

```text
API error: 429 - {"error":"Too Many Requests"}
```

**Cause:**

Rate limit exceeded. The API has received too many requests in a short period.

**Resolution:**

1. Wait a few seconds before retrying
2. The client will automatically retry with exponential backoff
3. Configure retry settings in `~/.config/syn/config.yaml`:

   ```yaml
   retry_config:
     max_attempts: 5
     initial_backoff: 2s
     max_backoff: 60s
   ```

### failed to send request

**Error Message:**

```text
failed to send request: Post "https://api.example.com/chat/completions":
dial tcp: lookup api.example.com: no such host
```

**Cause:**

DNS resolution failure or network connectivity issue.

**Resolution:**

1. Check your internet connection
2. Verify the API base URL in your config is correct
3. Try pinging the host: `ping api.example.com`
4. Check if a VPN or proxy is interfering

### API error: 500 / 502 / 503 / 504

**Error Message:**

```text
API error: 503 - Service Unavailable
```

**Cause:**

The API server is experiencing issues or is temporarily unavailable.

**Resolution:**

1. The client will automatically retry with exponential backoff
2. Wait a few minutes and try again
3. Check the API status page if available
4. If the issue persists, contact API support

### no choices in response

**Error Message:**

```text
no choices in response
```

**Cause:**

The API returned a response without any completion choices, possibly due to an invalid model or request format.

**Resolution:**

1. Verify the model name is correct: `syn model list`
2. Check if the model is available in your region
3. Try with a different model using the `--model` flag

### no texts provided for embedding

**Error Message:**

```text
no texts provided for embedding
```

**Cause:**

The embed command was called without providing any text input.

**Resolution:**

1. Provide text input via stdin or file:

   ```bash
   echo "your text here" | syn embed
   syn embed < input.txt
   ```

### failed to read image file

**Error Message:**

```text
failed to read image file: open image.png: no such file or directory
```

**Cause:**

The image file specified for vision analysis does not exist.

**Resolution:**

1. Verify the image file path is correct
2. Check the file exists: `ls -la image.png`
3. Use absolute paths if needed
4. Ensure the image format is supported (PNG, JPEG, GIF, WebP)

### Debug Mode

Enable verbose logging to get more detailed error information:

```bash
export SYN_VERBOSE=1
syn chat "your prompt"
```

Or in `~/.config/syn/config.yaml`:

```yaml
verbose: true
```

### Getting Help

If you encounter an error not covered in this guide:

1. Enable verbose mode to see detailed logs
2. Check the API documentation for the specific endpoint
3. Report issues at: <https://github.com/dotcommander/syn/issues>

## Examples

### Custom Retry Configuration

Handle transient failures with custom retry logic and exponential backoff.

```go
package main

import (
    "context"
    "fmt"
    "log/slog"
    "time"

    "github.com/dotcommander/syn/internal/app"
    "github.com/dotcommander/syn/internal/config"
)

func main() {
    config.SetDefaults()

    // Custom retry configuration for unreliable networks
    customClient := app.NewClient(app.ClientConfig{
        APIKey:    "your_api_key",
        BaseURL:   "https://api.synthetic.new/openai/v1",
        Model:     "hf:deepseek-ai/DeepSeek-V3.2",
        Timeout:   60 * time.Second,
        RetryConfig: app.RetryConfig{
            MaxAttempts:    5,              // Increase retries
            InitialBackoff: 2 * time.Second, // Start with 2s
            MaxBackoff:     60 * time.Second, // Cap at 60s
        },
    }, slog.Default(), nil)

    ctx := context.Background()
    response, err := customClient.Chat(ctx, "Explain quantum computing", app.ChatOptions{})
    if err != nil {
        // Handle API errors with proper error checking
        if apiErr, ok := err.(*app.APIError); ok {
            log.Fatalf("API error %d: %s", apiErr.StatusCode, apiErr.Body)
        }
        log.Fatalf("Request failed: %v", err)
    }

    fmt.Println(response)
}
```

### Conversational Context Management

Build multi-turn conversations with message history.

```go
package main

import (
    "context"
    "fmt"

    "github.com/dotcommander/syn/internal/app"
)

func main() {
    client := app.NewClient(app.ClientConfig{
        APIKey:  "your_api_key",
        BaseURL: "https://api.synthetic.new/openai/v1",
        Model:   "hf:deepseek-ai/DeepSeek-V3.2",
    }, nil, nil)

    ctx := context.Background()

    // Build conversation history
    conversation := []app.Message{
        {Role: "system", Content: "You are a helpful coding assistant."},
        {Role: "user", Content: "What is a closure in Go?"},
    }

    // First turn
    response1, err := client.Chat(ctx, "", app.ChatOptions{
        Context: conversation,
    })
    if err != nil {
        panic(err)
    }

    // Append assistant response to history
    conversation = append(conversation, app.Message{
        Role:    "assistant",
        Content: response1,
    })

    // Second turn - user asks follow-up
    conversation = append(conversation, app.Message{
        Role:    "user",
        Content: "Can you show me an example?",
    })

    response2, err := client.Chat(ctx, "", app.ChatOptions{
        Context: conversation,
    })
    if err != nil {
        panic(err)
    }

    fmt.Println(response2)
}
```

### File Analysis with Vision

Combine file reading with vision capabilities for document analysis.

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/dotcommander/syn/internal/app"
)

func main() {
    client := app.NewClient(app.ClientConfig{
        APIKey:  "your_api_key",
        BaseURL: "https://api.synthetic.new/openai/v1",
        Model:   "hf:deepseek-ai/DeepSeek-V3.2",
    }, nil, nil)

    ctx := context.Background()

    // Analyze a screenshot or diagram
    imagePath := "/path/to/screenshot.png"

    // Verify file exists before sending to API
    if _, err := os.Stat(imagePath); os.IsNotExist(err) {
        fmt.Printf("Image file not found: %s\n", imagePath)
        return
    }

    response, err := client.Vision(ctx, "Describe what you see in this image", imagePath, app.ChatOptions{})
    if err != nil {
        // Handle vision-specific errors
        if apiErr, ok := err.(*app.APIError); ok {
            switch apiErr.StatusCode {
            case 400:
                fmt.Println("Invalid image format or file too large")
            case 413:
                fmt.Println("Image file exceeds size limit")
            default:
                fmt.Printf("Vision API error: %s\n", apiErr.Body)
            }
            return
        }
        panic(err)
    }

    fmt.Println("Analysis:", response)
}
```

### Temperature Tuning for Different Use Cases

Adjust temperature to control response randomness.

```go
package main

import (
    "context"
    "fmt"

    "github.com/dotcommander/syn/internal/app"
)

func main() {
    client := app.NewClient(app.ClientConfig{
        APIKey:  "your_api_key",
        BaseURL: "https://api.synthetic.new/openai/v1",
        Model:   "hf:deepseek-ai/DeepSeek-V3.2",
    }, nil, nil)

    ctx := context.Background()

    prompt := "Write a short story about a robot"

    // Low temperature (0.0-0.3): Deterministic, factual
    tempLow := 0.1
    responseLow, _ := client.Chat(ctx, prompt, app.ChatOptions{
        Temperature: &tempLow,
    })
    fmt.Println("Low temperature:", responseLow)

    // Medium temperature (0.4-0.7): Balanced creativity
    tempMid := 0.6
    responseMid, _ := client.Chat(ctx, prompt, app.ChatOptions{
        Temperature: &tempMid,
    })
    fmt.Println("Medium temperature:", responseMid)

    // High temperature (0.8-2.0): Maximum creativity
    tempHigh := 1.2
    responseHigh, _ := client.Chat(ctx, prompt, app.ChatOptions{
        Temperature: &tempHigh,
    })
    fmt.Println("High temperature:", responseHigh)
}
```

### Batch Embeddings for Semantic Search

Generate embeddings for multiple texts at once.

```go
package main

import (
    "context"
    "fmt"

    "github.com/dotcommander/syn/internal/app"
)

func main() {
    client := app.NewClient(app.ClientConfig{
        APIKey:         "your_api_key",
        BaseURL:        "https://api.synthetic.new/openai/v1",
        EmbeddingModel: "hf:nomic-ai/nomic-embed-text-v1.5",
    }, nil, nil)

    ctx := context.Background()

    // Batch process multiple documents
    documents := []string{
        "The quick brown fox jumps over the lazy dog",
        "Machine learning is a subset of artificial intelligence",
        "Go is a statically typed programming language",
    }

    resp, err := client.Embed(ctx, documents, "")
    if err != nil {
        panic(err)
    }

    // Each document now has a vector embedding
    for i, embedding := range resp.Data {
        fmt.Printf("Document %d: %d dimensions\n", i, len(embedding.Embedding))
        // Use embeddings for similarity search, clustering, etc.
    }
}
```

### Context-Aware Chat with File Context

Include file contents as context for code review or documentation tasks.

```go
package main

import (
    "context"
    "fmt"

    "github.com/dotcommander/syn/internal/app"
)

func main() {
    client := app.NewClient(app.ClientConfig{
        APIKey:  "your_api_key",
        BaseURL: "https://api.synthetic.new/openai/v1",
        Model:   "hf:deepseek-ai/DeepSeek-V3.2",
    }, nil, nil)

    ctx := context.Background()

    // Include file content as context
    response, err := client.Chat(ctx, "Review this code for potential issues", app.ChatOptions{
        FilePath: "main.go",
        MaxTokens: ptr(4096), // Limit response length
    })
    if err != nil {
        // Handle file-specific errors
        if err.Error() == "failed to read file" {
            fmt.Println("Could not read the specified file")
            return
        }
        panic(err)
    }

    fmt.Println(response)
}

func ptr[T any](v T) *T {
    return &v
}
```
