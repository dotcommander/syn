# app Package

Core client library for interacting with Synthetic.new API.

## Interfaces

### ChatClient

```go
type ChatClient interface {
    Chat(ctx context.Context, prompt string, opts ChatOptions) (string, error)
}
```
Interface for chat completion operations.

### ModelClient
```go
type ModelClient interface {
    ListModels(ctx context.Context) ([]Model, error)
}
```
Interface for listing available models.

### EmbeddingClient
```go
type EmbeddingClient interface {
    Embed(ctx context.Context, texts []string, model string) (*EmbeddingResponse, error)
}
```
Interface for generating text embeddings.

### VisionClient
```go
type VisionClient interface {
    Vision(ctx context.Context, prompt string, imageSource string, opts ChatOptions) (string, error)
}
```
Interface for vision/image analysis operations.

### SearchClient
```go
type SearchClient interface {
    Search(ctx context.Context, query string) (*SearchResponse, error)
}
```
Interface for web search operations. Uses `/v2/search` endpoint with zero-data-retention.

## Types

### Client
```go
type Client struct {
    // contains filtered or exported fields
}
```
Main client implementing all API interfaces. Create with `NewClient()`.

### ClientConfig
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

### RetryConfig
```go
type RetryConfig struct {
    MaxAttempts    int           // Default: 3
    InitialBackoff time.Duration // Default: 1s
    MaxBackoff     time.Duration // Default: 30s
}
```
Retry behavior configuration for transient failures.

### Message
```go
type Message struct {
    Role    string // "user", "assistant", "system"
    Content string
}
```
Chat message representation.

### ChatOptions
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

### Usage
```go
type Usage struct {
    PromptTokens     int
    CompletionTokens int
    TotalTokens      int
}
```
Token usage statistics.

### Model
```go
type Model struct {
    ID      string
    Object  string
    Created int64
    OwnedBy string
}
```
Model information.

### APIError
```go
type APIError struct {
    StatusCode int
    Body       string
}
```
API error response.

### SearchResponse
```go
type SearchResponse struct {
    Results []SearchResult
}
```
Response from the /v2/search endpoint.

### SearchResult
```go
type SearchResult struct {
    Title     string
    URL       string
    Snippet   string
    Published string // ISO 8601 timestamp (optional)
}
```
Individual search result with metadata.

## Functions

### NewClient
```go
func NewClient(cfg ClientConfig, logger *slog.Logger, httpClient HTTPDoer) *Client
```
Creates a new API client with injected dependencies.

### NewLogger
```go
func NewLogger(verbose bool) *slog.Logger
```
Creates a structured logger for the application.

### ResolveModel
```go
func ResolveModel(model string) string
```
Resolves model aliases (gpt, kimi, qwen, glm, deepseek) to full model names.

### DefaultChatOptions
```go
func DefaultChatOptions() ChatOptions
```
Returns sensible default options for chat requests.

## Methods

### (*Client).Chat
```go
func (c *Client) Chat(ctx context.Context, prompt string, opts ChatOptions) (string, error)
```
Sends a chat prompt and returns the response. Supports file inclusion via `opts.FilePath`.

### (*Client).ListModels
```go
func (c *Client) ListModels(ctx context.Context) ([]Model, error)
```
Fetches available models from the API.

### (*Client).Embed
```go
func (c *Client) Embed(ctx context.Context, texts []string, model string) (*EmbeddingResponse, error)
```
Generates embeddings for the given texts.

### (*Client).Vision
```go
func (c *Client) Vision(ctx context.Context, prompt string, imageSource string, opts ChatOptions) (string, error)
```
Analyzes an image with a prompt. `imageSource` can be a URL or local file path.

### (*Client).Search
```go
func (c *Client) Search(ctx context.Context, query string) (*SearchResponse, error)
```
Performs a web search. Returns results with title, URL, snippet, and optional published timestamp.

## Model Aliases

| Alias | Full Model Name |
|-------|-----------------|
| gpt, gptoss | hf:openai/gpt-oss-120b |
| kimi | hf:moonshotai/Kimi-K2-Thinking |
| qwen | hf:Qwen/Qwen3-VL-235B-A22B-Instruct |
| glm, zai | hf:zai-org/GLM-4.7 |
| deepseek, ds | hf:deepseek-ai/DeepSeek-V3.2 |
