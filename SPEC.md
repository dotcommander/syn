# SYN - Synthetic.new CLI Specification

## TL;DR - Quick Reference

| Concern | Root Cause | Solution |
|---------|------------|----------|
| New CLI for synthetic.new | No existing CLI tool for synthetic.new API | Replicate zai architecture with synthetic.new endpoints |
| OpenAI-compatible API | Synthetic uses standard OpenAI-compatible endpoints | Adapt zai's client.go with synthetic.new base URLs |
| Multiple model providers | 20+ always-on models from different providers | Implement model listing and selection via `/models` endpoint |
| Embedding support | Synthetic includes nomic-embed-text embedding model | Add embedding command (new feature vs zai) |

---

## Implementation Priority

| Priority | Task | Effort | Files/Location |
|----------|------|--------|----------------|
| **P0** | Project structure and module setup | ~10 lines | go.mod, main.go |
| **P0** | Core client with /chat/completions | ~300 lines | internal/app/client.go, types.go |
| **P0** | Root command with one-shot mode | ~150 lines | cmd/root.go |
| **P1** | Interactive chat REPL | ~200 lines | cmd/chat.go |
| **P1** | Model listing and management | ~100 lines | cmd/model.go |
| **P1** | Configuration file support | ~100 lines | internal/config/config.go |
| **P2** | Embeddings command | ~150 lines | cmd/embed.go |
| **P2** | Text completions command | ~100 lines | cmd/complete.go |
| **P2** | History storage | ~150 lines | internal/app/history.go |
| **P3** | Quota checking | ~50 lines | cmd/quota.go |
| **P3** | Shell completion | ~50 lines | cmd/completion.go |

---

## P0: Project Foundation

### Root Cause / Context

Starting a new Go CLI project requires proper module initialization and dependency management. The zai project uses:
- Go 1.25.4
- Cobra for CLI structure
- Viper for configuration
- Lipgloss for terminal UI styling
- Standard Go project layout (cmd/, internal/)

### Implementation

**Directory Structure:**
```
syn/
├── main.go                 # Entry point
├── go.mod                  # Module definition
├── cmd/                    # Commands
│   ├── root.go            # Root command + one-shot
│   ├── chat.go            # Interactive REPL
│   ├── model.go           # Model management
│   ├── embed.go           # Embeddings
│   ├── complete.go        # Text completions
│   ├── quota.go           # Quota checking
│   └── theme.go           # Lipgloss styling
├── internal/
│   ├── app/
│   │   ├── client.go      # API client
│   │   ├── types.go       # Request/response types
│   │   └── history.go     # History storage
│   └── config/
│       └── config.go      # Viper configuration
└── README.md
```

**File: `go.mod`**
```go
module github.com/yourusername/syn

go 1.25.4

require (
	github.com/charmbracelet/lipgloss v1.1.0
	github.com/spf13/cobra v1.10.2
	github.com/spf13/viper v1.21.0
	golang.org/x/sync v0.19.0
)
```

**File: `main.go`**
```go
package main

import (
	"github.com/yourusername/syn/cmd"
	"github.com/yourusername/syn/internal/config"
)

func main() {
	config.SetDefaults()
	cmd.Execute()
}
```

**File: `internal/config/config.go`**
```go
package config

import (
	"time"
	"github.com/spf13/viper"
)

// SetDefaults configures sensible defaults for the application.
func SetDefaults() {
	// API defaults
	viper.SetDefault("api.base_url", "https://api.synthetic.new/openai/v1")
	viper.SetDefault("api.anthropic_base_url", "https://api.synthetic.new/anthropic/v1")
	viper.SetDefault("api.model", "deepseek-v3.2")
	viper.SetDefault("api.embedding_model", "nomic-embed-text-v1.5")

	// Retry configuration
	viper.SetDefault("api.retry.max_attempts", 3)
	viper.SetDefault("api.retry.initial_backoff", 1*time.Second)
	viper.SetDefault("api.retry.max_backoff", 30*time.Second)

	// Chat defaults
	viper.SetDefault("chat.temperature", 0.6)
	viper.SetDefault("chat.max_tokens", 8192)
	viper.SetDefault("chat.top_p", 0.9)
}
```

### Verification

```bash
# Initialize module
cd ~/go/src/syn
go mod init github.com/yourusername/syn
go mod tidy

# Build should succeed
go build -o bin/syn .

# Verify structure
tree -L 2
```

---

## P0: Core API Client

### Root Cause / Context

The synthetic.new API is OpenAI-compatible, using standard endpoints at `https://api.synthetic.new/openai/v1`. The zai client architecture is already well-designed with:
- Dependency injection for testability
- Interface segregation (ChatClient, ModelClient, etc.)
- Retry logic with exponential backoff
- Proper error handling

We can reuse 80% of zai's client.go architecture, changing only:
1. Base URL to synthetic.new
2. Remove Z.AI-specific features (web reader, web search, vision, audio, video, image generation)
3. Add embeddings support (new feature)
4. Keep chat/completions and models endpoints

### Implementation

**File: `internal/app/types.go`**
```go
package app

import "fmt"

// ClientConfig holds all configuration for the Synthetic client.
type ClientConfig struct {
	APIKey         string
	BaseURL        string // OpenAI-compatible
	AnthropicURL   string // Anthropic-compatible
	Model          string
	EmbeddingModel string
	Timeout        time.Duration
	Verbose        bool
	RetryConfig    RetryConfig
}

// RetryConfig configures retry behavior for transient failures.
type RetryConfig struct {
	MaxAttempts    int           // Maximum number of retry attempts (default: 3)
	InitialBackoff time.Duration // Initial backoff duration (default: 1s)
	MaxBackoff     time.Duration // Maximum backoff duration (default: 30s)
}

// Message represents a chat message.
type Message struct {
	Role    string `json:"role"`    // "user", "assistant", "system"
	Content string `json:"content"`
}

// ChatRequest represents the /chat/completions API request.
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	TopP        float64   `json:"top_p,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

// ChatResponse represents the /chat/completions API response.
type ChatResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice represents a response choice.
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage represents token usage statistics.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ModelsResponse represents the /models API response.
type ModelsResponse struct {
	Object string  `json:"object"`
	Data   []Model `json:"data"`
}

// Model represents a single model.
type Model struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

// CompletionRequest represents the /completions API request.
type CompletionRequest struct {
	Model       string  `json:"model"`
	Prompt      string  `json:"prompt"`
	Temperature float64 `json:"temperature,omitempty"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
	TopP        float64 `json:"top_p,omitempty"`
	Stream      bool    `json:"stream,omitempty"`
}

// CompletionResponse represents the /completions API response.
type CompletionResponse struct {
	ID      string              `json:"id"`
	Object  string              `json:"object"`
	Created int64               `json:"created"`
	Model   string              `json:"model"`
	Choices []CompletionChoice  `json:"choices"`
	Usage   Usage               `json:"usage"`
}

// CompletionChoice represents a completion choice.
type CompletionChoice struct {
	Text         string `json:"text"`
	Index        int    `json:"index"`
	FinishReason string `json:"finish_reason"`
}

// EmbeddingRequest represents the /embeddings API request.
type EmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"` // Array of strings to embed
}

// EmbeddingResponse represents the /embeddings API response.
type EmbeddingResponse struct {
	Object string             `json:"object"`
	Data   []EmbeddingData    `json:"data"`
	Model  string             `json:"model"`
	Usage  EmbeddingUsage     `json:"usage"`
}

// EmbeddingData represents a single embedding.
type EmbeddingData struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

// EmbeddingUsage represents embedding token usage.
type EmbeddingUsage struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// ChatOptions configures chat requests.
type ChatOptions struct {
	Model       string
	Temperature *float64
	MaxTokens   *int
	TopP        *float64
	FilePath    string    // Optional file to include in context
	Context     []Message // Previous messages for context
}

// APIError represents an error response from the API.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error: %d - %s", e.StatusCode, e.Body)
}

// Helper functions for creating pointers
func Float64Ptr(v float64) *float64 { return &v }
func IntPtr(v int) *int             { return &v }
func BoolPtr(v bool) *bool          { return &v }

// DefaultChatOptions returns sensible defaults for CLI usage.
func DefaultChatOptions() ChatOptions {
	return ChatOptions{
		Temperature: Float64Ptr(0.6),
		MaxTokens:   IntPtr(8192),
		TopP:        Float64Ptr(0.9),
	}
}
```

**File: `internal/app/client.go`** (partial - key methods)
```go
package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"os"
	"strings"
	"time"
)

// ChatClient interface for testability (ISP compliance).
type ChatClient interface {
	Chat(ctx context.Context, prompt string, opts ChatOptions) (string, error)
}

// ModelClient interface for model listing (ISP compliance).
type ModelClient interface {
	ListModels(ctx context.Context) ([]Model, error)
}

// CompletionClient interface for text completions (ISP compliance).
type CompletionClient interface {
	Complete(ctx context.Context, prompt string, opts ChatOptions) (string, error)
}

// EmbeddingClient interface for embeddings (ISP compliance).
type EmbeddingClient interface {
	Embed(ctx context.Context, texts []string, model string) (*EmbeddingResponse, error)
}

// HTTPDoer interface for HTTP operations (DIP compliance, enables testing).
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client implements all client interfaces with Synthetic API.
type Client struct {
	config     ClientConfig
	httpClient HTTPDoer
	logger     *slog.Logger
}

// NewClient creates a client with injected dependencies.
func NewClient(cfg ClientConfig, logger *slog.Logger, httpClient HTTPDoer) *Client {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 60 * time.Second
	}

	if httpClient == nil {
		httpClient = &http.Client{Timeout: timeout}
	}

	return &Client{
		config:     cfg,
		httpClient: httpClient,
		logger:     logger,
	}
}

// NewLogger creates a slog.Logger for the application.
func NewLogger(verbose bool) *slog.Logger {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	opts := &slog.HandlerOptions{Level: level}
	return slog.New(slog.NewTextHandler(os.Stderr, opts))
}

// requireAPIKey validates the API key is configured.
func (c *Client) requireAPIKey() error {
	if c.config.APIKey == "" {
		return fmt.Errorf("API key is not configured. Set SYN_API_KEY or configure in ~/.config/syn/config.yaml")
	}
	return nil
}

// Chat sends a prompt and returns the response.
func (c *Client) Chat(ctx context.Context, prompt string, opts ChatOptions) (string, error) {
	if err := c.requireAPIKey(); err != nil {
		return "", err
	}

	// Build message content (with optional file)
	content, err := c.buildContent(prompt, opts.FilePath)
	if err != nil {
		return "", err
	}

	// Build messages array with context
	messages := c.buildMessagesWithContext(content, opts)

	// Execute request with retry
	response, usage, err := c.doRequestWithRetry(ctx, messages, opts)
	if err != nil {
		return "", err
	}

	c.logger.Debug("chat complete",
		"total_tokens", usage.TotalTokens,
		"prompt_tokens", usage.PromptTokens,
		"completion_tokens", usage.CompletionTokens)

	return response, nil
}

// buildContent combines prompt with optional file contents.
func (c *Client) buildContent(prompt, filePath string) (string, error) {
	if filePath == "" {
		return prompt, nil
	}

	// Read local file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	return fmt.Sprintf("%s\n\nFile contents (%s):\n```\n%s\n```", prompt, filePath, string(data)), nil
}

// buildMessagesWithContext constructs messages array including conversation context.
func (c *Client) buildMessagesWithContext(content string, opts ChatOptions) []Message {
	messages := c.buildMessages(content, opts)

	// Prepend context messages if provided
	if len(opts.Context) > 0 {
		messages = append(opts.Context, messages...)
	}

	return messages
}

// buildMessages constructs the messages array for the API.
func (c *Client) buildMessages(content string, opts ChatOptions) []Message {
	var messages []Message

	// Add system prompt
	messages = append(messages, Message{
		Role:    "system",
		Content: "Be concise and direct. Answer briefly and to the point.",
	})

	// Add current user message
	messages = append(messages, Message{
		Role:    "user",
		Content: content,
	})

	return messages
}

// doRequest executes the HTTP request to Synthetic API.
func (c *Client) doRequest(ctx context.Context, messages []Message, opts ChatOptions) (string, Usage, error) {
	reqData := ChatRequest{
		Model:    c.config.Model,
		Messages: messages,
		Stream:   false,
	}

	// Apply optional overrides
	if opts.Temperature != nil {
		reqData.Temperature = *opts.Temperature
	} else {
		reqData.Temperature = 0.6
	}

	if opts.MaxTokens != nil {
		reqData.MaxTokens = *opts.MaxTokens
	} else {
		reqData.MaxTokens = 8192
	}

	if opts.TopP != nil {
		reqData.TopP = *opts.TopP
	} else {
		reqData.TopP = 0.9
	}

	// Apply model override if provided
	if opts.Model != "" {
		reqData.Model = opts.Model
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return "", Usage{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/chat/completions", c.config.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", Usage{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))

	c.logger.Debug("sending request", "url", url)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", Usage{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", Usage{}, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", Usage{}, &APIError{StatusCode: resp.StatusCode, Body: string(body)}
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", Usage{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", Usage{}, fmt.Errorf("no choices in response")
	}

	return chatResp.Choices[0].Message.Content, chatResp.Usage, nil
}

// doRequestWithRetry executes doRequest with exponential backoff retry logic.
func (c *Client) doRequestWithRetry(ctx context.Context, messages []Message, opts ChatOptions) (string, Usage, error) {
	var lastErr error

	maxAttempts := c.config.RetryConfig.MaxAttempts
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	initialBackoff := c.config.RetryConfig.InitialBackoff
	if initialBackoff < 1 {
		initialBackoff = 1 * time.Second
	}

	maxBackoff := c.config.RetryConfig.MaxBackoff
	if maxBackoff < 1 {
		maxBackoff = 30 * time.Second
	}

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return "", Usage{}, ctx.Err()
		default:
		}

		if attempt > 1 {
			backoff := calculateBackoff(attempt, initialBackoff, maxBackoff)
			c.logger.Debug("retrying request",
				"attempt", attempt,
				"max_attempts", maxAttempts,
				"backoff", backoff,
				"error", lastErr)

			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return "", Usage{}, ctx.Err()
			}
		}

		response, usage, err := c.doRequest(ctx, messages, opts)
		if err == nil {
			return response, usage, nil
		}

		lastErr = err

		if !isRetryableError(err) || attempt == maxAttempts {
			break
		}
	}

	return "", Usage{}, fmt.Errorf("request failed after %d attempts: %w", maxAttempts, lastErr)
}

// isRetryableError checks if an error should trigger a retry.
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	var netErr interface{ Timeout() bool }
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	errStr := err.Error()
	retryablePatterns := []string{
		"connection refused",
		"connection reset",
		"temporary failure",
		"timeout",
		"429", // Too Many Requests
		"503", // Service Unavailable
		"502", // Bad Gateway
		"504", // Gateway Timeout
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(strings.ToLower(errStr), pattern) {
			return true
		}
	}

	return false
}

// calculateBackoff calculates exponential backoff with jitter.
func calculateBackoff(attempt int, initialBackoff, maxBackoff time.Duration) time.Duration {
	if attempt > 62 {
		attempt = 62
	}

	backoff := initialBackoff * time.Duration(1<<uint(attempt-1))

	if backoff > maxBackoff {
		backoff = maxBackoff
	}

	jitterRange := float64(backoff) * 0.125
	jitter := time.Duration(jitterRange * (2.0*rand.Float64() - 1.0))

	return backoff + jitter
}

// ListModels fetches available models from the API.
func (c *Client) ListModels(ctx context.Context) ([]Model, error) {
	if err := c.requireAPIKey(); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/models", c.config.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))

	c.logger.Debug("sending request", "url", url)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &APIError{StatusCode: resp.StatusCode, Body: string(body)}
	}

	var modelsResp ModelsResponse
	if err := json.Unmarshal(body, &modelsResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal models response: %w", err)
	}

	return modelsResp.Data, nil
}

// Embed generates embeddings for the given texts.
func (c *Client) Embed(ctx context.Context, texts []string, model string) (*EmbeddingResponse, error) {
	if err := c.requireAPIKey(); err != nil {
		return nil, err
	}

	if len(texts) == 0 {
		return nil, fmt.Errorf("no texts provided for embedding")
	}

	if model == "" {
		model = c.config.EmbeddingModel
	}

	reqData := EmbeddingRequest{
		Model: model,
		Input: texts,
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/embeddings", c.config.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))

	c.logger.Debug("sending embeddings request", "url", url, "texts", len(texts))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &APIError{StatusCode: resp.StatusCode, Body: string(body)}
	}

	var embedResp EmbeddingResponse
	if err := json.Unmarshal(body, &embedResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal embedding response: %w", err)
	}

	c.logger.Debug("embeddings complete",
		"embeddings", len(embedResp.Data),
		"total_tokens", embedResp.Usage.TotalTokens)

	return &embedResp, nil
}
```

### Verification

```bash
# Run unit tests
go test ./internal/app/... -v

# Test live API (requires SYN_API_KEY)
export SYN_API_KEY="your-api-key"
go run . "What is the meaning of life?"
```

---

## P0: Root Command and One-Shot Mode

### Root Cause / Context

The root command provides the entry point for CLI usage. Based on zai's architecture:
- One-shot mode: `syn "prompt"` executes immediately
- Stdin support: `echo "text" | syn` or `echo "text" | syn "prompt"`
- Flag support: `-f file.txt`, `--json`, `-v`
- No args = show help

### Implementation

**File: `cmd/root.go`**
```go
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/yourusername/syn/internal/app"
)

var (
	cfgFile    string
	verbose    bool
	filePath   string
	jsonOutput bool
)

var rootCmd = &cobra.Command{
	Use:   "syn [prompt]",
	Short: "Chat with Synthetic.new AI models",
	Long: `SYN is a CLI tool for interacting with Synthetic.new models.

One-shot mode:
  syn "Explain quantum computing"
  syn -f main.go "Explain this code"

Piped input:
  pbpaste | syn "explain this"
  cat file.txt | syn "summarize"

Interactive REPL:
  syn chat`,
	Args: cobra.ArbitraryArgs,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Name() == "completion" || cmd.Name() == "help" || cmd.Name() == "version" {
			return nil
		}
		return initConfig()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var prompt string
		var stdinData string

		// Check for stdin data
		if hasStdinData() {
			data, err := readStdin()
			if err != nil {
				return fmt.Errorf("failed to read stdin: %w", err)
			}
			stdinData = data
		}

		// Build prompt from args
		if len(args) > 0 {
			prompt = strings.Join(args, " ")
		}

		// Combine: prompt + stdin
		if stdinData != "" {
			if prompt != "" {
				prompt = prompt + "\n\n<stdin>\n" + stdinData + "\n</stdin>"
			} else {
				prompt = stdinData
			}
		}

		// Require some input
		if prompt == "" {
			return cmd.Help()
		}

		return runOneShot(prompt)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "\nError: %s\n\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default $HOME/.config/syn/config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVarP(&filePath, "file", "f", "", "include file contents in prompt")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output in JSON format")

	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	_ = viper.BindPFlag("file", rootCmd.PersistentFlags().Lookup("file"))
	_ = viper.BindPFlag("json", rootCmd.PersistentFlags().Lookup("json"))
}

func initConfig() error {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		configDir := filepath.Join(home, ".config", "syn")
		viper.AddConfigPath(configDir)
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	viper.SetEnvPrefix("SYN")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if viper.GetString("api.key") == "" {
		return fmt.Errorf("API key required: set SYN_API_KEY or configure in ~/.config/syn/config.yaml")
	}

	return nil
}

func buildClientConfig() app.ClientConfig {
	retryCfg := app.RetryConfig{
		MaxAttempts:    viper.GetInt("api.retry.max_attempts"),
		InitialBackoff: viper.GetDuration("api.retry.initial_backoff"),
		MaxBackoff:     viper.GetDuration("api.retry.max_backoff"),
	}

	return app.ClientConfig{
		APIKey:         viper.GetString("api.key"),
		BaseURL:        viper.GetString("api.base_url"),
		AnthropicURL:   viper.GetString("api.anthropic_base_url"),
		Model:          viper.GetString("api.model"),
		EmbeddingModel: viper.GetString("api.embedding_model"),
		Verbose:        viper.GetBool("verbose"),
		RetryConfig:    retryCfg,
	}
}

func newClient() *app.Client {
	cfg := buildClientConfig()
	logger := app.NewLogger(cfg.Verbose)
	return app.NewClient(cfg, logger, nil)
}

func hasStdinData() bool {
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) == 0
}

func readStdin() (string, error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func runOneShot(prompt string) error {
	client := newClient()
	opts := app.DefaultChatOptions()
	opts.FilePath = viper.GetString("file")

	if viper.GetBool("verbose") {
		fmt.Fprintf(os.Stderr, "Prompt: %s\n", prompt)
		if opts.FilePath != "" {
			fmt.Fprintf(os.Stderr, "File: %s\n", opts.FilePath)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	response, err := client.Chat(ctx, prompt, opts)
	if err != nil {
		return fmt.Errorf("failed to get response: %w", err)
	}

	if viper.GetBool("json") {
		output := map[string]interface{}{
			"prompt":    prompt,
			"response":  response,
			"model":     viper.GetString("api.model"),
			"file":      opts.FilePath,
			"timestamp": time.Now().Format(time.RFC3339),
		}

		data, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
	} else {
		fmt.Println(response)
	}

	return nil
}
```

### Verification

```bash
# Test one-shot mode
syn "Hello, world!"

# Test with file
syn -f main.go "Explain this code"

# Test with stdin
echo "Summarize this text" | syn

# Test JSON output
syn "Test" --json

# Test verbose mode
syn -v "Test" 2>&1 | grep "Prompt:"
```

---

## P1: Interactive Chat REPL

### Root Cause / Context

Interactive mode requires:
- Conversation context management (last 20 messages)
- Command parsing (special commands like `/clear`, `/help`)
- Terminal UI with lipgloss styling
- Graceful exit handling

Copy zai's chat.go architecture with minor adjustments for synthetic.new.

### Implementation

**File: `cmd/chat.go`** (key sections - full implementation similar to zai)
```go
package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/yourusername/syn/internal/app"
)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Start interactive chat session",
	Long: `Interactive REPL with conversation context.

Commands:
  /clear  - Clear conversation history
  /model  - Show current model
  /exit   - Exit chat session
  /help   - Show help`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInteractiveChat()
	},
}

func init() {
	rootCmd.AddCommand(chatCmd)
}

func runInteractiveChat() error {
	client := newClient()

	// Conversation context (last 20 messages = 10 exchanges)
	var context []app.Message
	maxContextMessages := 20

	fmt.Println(theme.Title.Render(" SYN ") + " " + theme.Description.Render("Chat Session"))
	fmt.Println(theme.Description.Render("Model: " + viper.GetString("api.model")))
	fmt.Println(theme.HelpText.Render("Type /help for commands, Ctrl+C to exit"))
	fmt.Println()

	for {
		// Prompt user
		fmt.Print(theme.UserPrompt.Render("you> "))

		var input string
		if _, err := fmt.Scanln(&input); err != nil {
			if err.Error() == "EOF" {
				fmt.Println()
				return nil
			}
			return err
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// Handle special commands
		if strings.HasPrefix(input, "/") {
			if err := handleChatCommand(input, &context); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			}
			continue
		}

		// Build chat options with context
		opts := app.DefaultChatOptions()
		opts.Context = context

		// Send message
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		response, err := client.Chat(ctx, input, opts)
		cancel()

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s %s\n",
				theme.ErrorText.Render("Error:"),
				err.Error())
			continue
		}

		// Display response
		fmt.Println(theme.AssistantPrompt.Render("assistant> ") + response)
		fmt.Println()

		// Update context
		context = append(context, app.Message{Role: "user", Content: input})
		context = append(context, app.Message{Role: "assistant", Content: response})

		// Trim context to last N messages
		if len(context) > maxContextMessages {
			context = context[len(context)-maxContextMessages:]
		}
	}
}

func handleChatCommand(cmd string, context *[]app.Message) error {
	switch cmd {
	case "/clear":
		*context = nil
		fmt.Println(theme.HelpText.Render("Context cleared"))
		return nil
	case "/model":
		fmt.Printf("%s %s\n",
			theme.HelpText.Render("Current model:"),
			viper.GetString("api.model"))
		return nil
	case "/exit":
		fmt.Println(theme.HelpText.Render("Goodbye!"))
		os.Exit(0)
		return nil
	case "/help":
		printChatHelp()
		return nil
	default:
		return fmt.Errorf("unknown command: %s (type /help for commands)", cmd)
	}
}

func printChatHelp() {
	fmt.Println(theme.Section.Render("Chat Commands"))
	commands := [][]string{
		{"/clear", "Clear conversation history"},
		{"/model", "Show current model"},
		{"/exit", "Exit chat session"},
		{"/help", "Show this help"},
	}
	for _, c := range commands {
		fmt.Printf("  %s  %s\n",
			theme.Command.Render(fmt.Sprintf("%-10s", c[0])),
			theme.Description.Render(c[1]))
	}
	fmt.Println()
}
```

**File: `cmd/theme.go`**
```go
package cmd

import "github.com/charmbracelet/lipgloss"

// Theme colors and styles
var theme = struct {
	Title            lipgloss.Style
	Section          lipgloss.Style
	Description      lipgloss.Style
	Command          lipgloss.Style
	Flag             lipgloss.Style
	Example          lipgloss.Style
	Divider          lipgloss.Style
	UserPrompt       lipgloss.Style
	AssistantPrompt  lipgloss.Style
	ErrorText        lipgloss.Style
	HelpText         lipgloss.Style
}{
	Title:            lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")),
	Section:          lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("33")),
	Description:      lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
	Command:          lipgloss.NewStyle().Foreground(lipgloss.Color("42")),
	Flag:             lipgloss.NewStyle().Foreground(lipgloss.Color("214")),
	Example:          lipgloss.NewStyle().Foreground(lipgloss.Color("246")),
	Divider:          lipgloss.NewStyle().Foreground(lipgloss.Color("238")),
	UserPrompt:       lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")),
	AssistantPrompt:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42")),
	ErrorText:        lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196")),
	HelpText:         lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
}
```

### Verification

```bash
# Start interactive chat
syn chat

# Test commands
you> /help
you> /model
you> Hello!
you> /clear
you> /exit
```

---

## P1: Model Management

### Implementation

**File: `cmd/model.go`**
```go
package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var modelCmd = &cobra.Command{
	Use:   "model",
	Short: "Model management",
	Long:  `List and manage Synthetic.new models.`,
}

var modelListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available models",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		models, err := client.ListModels(ctx)
		if err != nil {
			return fmt.Errorf("failed to list models: %w", err)
		}

		fmt.Printf("Available models: %d\n\n", len(models))

		for _, m := range models {
			fmt.Printf("  %s\n", theme.Command.Render(m.ID))
			if m.OwnedBy != "" {
				fmt.Printf("    %s: %s\n", theme.Description.Render("Owner"), m.OwnedBy)
			}
			created := time.Unix(m.Created, 0)
			fmt.Printf("    %s: %s\n", theme.Description.Render("Created"), created.Format("2006-01-02"))
			fmt.Println()
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(modelCmd)
	modelCmd.AddCommand(modelListCmd)
}
```

---

## P2: Embeddings Command

### Implementation

**File: `cmd/embed.go`**
```go
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var embedCmd = &cobra.Command{
	Use:   "embed [text...]",
	Short: "Generate text embeddings",
	Long: `Generate vector embeddings for text using nomic-embed-text-v1.5.

Examples:
  syn embed "Hello world"
  syn embed "Text 1" "Text 2" "Text 3"
  echo "Text" | syn embed`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var texts []string

		// Check for stdin
		if hasStdinData() {
			data, err := readStdin()
			if err != nil {
				return fmt.Errorf("failed to read stdin: %w", err)
			}
			texts = append(texts, data)
		}

		// Add args
		texts = append(texts, args...)

		if len(texts) == 0 {
			return fmt.Errorf("no text provided (use args or stdin)")
		}

		return runEmbed(texts)
	},
}

func init() {
	rootCmd.AddCommand(embedCmd)
}

func runEmbed(texts []string) error {
	client := newClient()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	model := viper.GetString("api.embedding_model")
	resp, err := client.Embed(ctx, texts, model)
	if err != nil {
		return fmt.Errorf("embedding failed: %w", err)
	}

	if viper.GetBool("json") {
		// Output raw JSON
		data, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	} else {
		// Human-readable output
		fmt.Printf("Generated %d embeddings\n", len(resp.Data))
		fmt.Printf("Model: %s\n", resp.Model)
		fmt.Printf("Total tokens: %d\n\n", resp.Usage.TotalTokens)

		for i, emb := range resp.Data {
			fmt.Printf("Text %d: %s\n", i+1, truncate(texts[i], 60))
			fmt.Printf("  Embedding dimensions: %d\n", len(emb.Embedding))
			fmt.Printf("  First 5 values: %v\n\n", emb.Embedding[:min(5, len(emb.Embedding))])
		}
	}

	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
```

---

## Configuration File

### Implementation

**File: `~/.config/syn/config.yaml`** (example)
```yaml
api:
  key: "your-api-key-here"
  base_url: "https://api.synthetic.new/openai/v1"
  anthropic_base_url: "https://api.synthetic.new/anthropic/v1"
  model: "deepseek-v3.2"
  embedding_model: "nomic-embed-text-v1.5"

  retry:
    max_attempts: 3
    initial_backoff: 1s
    max_backoff: 30s

chat:
  temperature: 0.6
  max_tokens: 8192
  top_p: 0.9
```

---

## API Differences from Z.AI

| Feature | Z.AI | Synthetic.new |
|---------|------|---------------|
| **Base URL** | `https://api.z.ai/api/paas/v4` | `https://api.synthetic.new/openai/v1` |
| **Chat endpoint** | `/chat/completions` | `/chat/completions` (same) |
| **Models endpoint** | `/models` | `/models` (same) |
| **Web reader** | ✓ Z.AI-specific | ✗ Not available |
| **Web search** | ✓ Z.AI-specific | ✗ Not available |
| **Vision** | ✓ GLM-4.6v | ✗ Not in OpenAI endpoint |
| **Audio** | ✓ ASR models | ✗ Not in OpenAI endpoint |
| **Video** | ✓ CogVideoX-3 | ✗ Not in OpenAI endpoint |
| **Image generation** | ✓ CogView-4 | ✗ Not in OpenAI endpoint |
| **Embeddings** | ✗ Not mentioned | ✓ nomic-embed-text-v1.5 |
| **Text completions** | ✗ Not mentioned | ✓ `/completions` |
| **Thinking mode** | ✓ GLM-4.7 reasoning | ✗ Not mentioned |
| **Anthropic API** | ✗ No | ✓ Alternative endpoint |

---

## Architecture Summary

```
syn/
├── main.go                 # Entry point → cmd.Execute()
├── cmd/
│   ├── root.go            # One-shot mode, stdin, flags
│   ├── chat.go            # Interactive REPL with context
│   ├── model.go           # List models
│   ├── embed.go           # Generate embeddings
│   ├── complete.go        # Text completions (P2)
│   ├── quota.go           # Check quotas (P3)
│   └── theme.go           # Lipgloss styling
└── internal/
    ├── app/
    │   ├── client.go      # HTTP client with retry logic
    │   ├── types.go       # Request/response structs
    │   └── history.go     # JSONL history storage (P2)
    └── config/
        └── config.go      # Viper defaults
```

**Key Design Decisions:**
1. **SOLID principles**: Interface segregation (ChatClient, ModelClient, EmbeddingClient)
2. **Dependency injection**: Client takes logger and httpClient as dependencies
3. **Retry logic**: Exponential backoff with jitter for transient failures
4. **Context management**: Last 20 messages in REPL (10 exchanges)
5. **Configuration**: Viper + environment variables (SYN_API_KEY)
6. **Styling**: Lipgloss for terminal UI consistency

---

## Files to Create

| File | Purpose | Lines |
|------|---------|-------|
| `main.go` | Entry point | 11 |
| `go.mod` | Module definition | 10 |
| `internal/config/config.go` | Viper defaults | 25 |
| `internal/app/types.go` | Request/response types | 250 |
| `internal/app/client.go` | API client implementation | 400 |
| `cmd/root.go` | Root command + one-shot | 200 |
| `cmd/chat.go` | Interactive REPL | 150 |
| `cmd/model.go` | Model management | 50 |
| `cmd/embed.go` | Embeddings | 100 |
| `cmd/theme.go` | Lipgloss styles | 30 |
| `README.md` | Documentation | 150 |
| **Total** | | **~1,376 lines** |

---

## Test Cases

### Chat Endpoint
```bash
# Input
syn "What is 2+2?"

# Expected Output
"4" (or natural language explanation)

# Verify
- API called with correct endpoint
- Response parsed correctly
- Token usage logged
```

### Embeddings
```bash
# Input
syn embed "Hello world" "Goodbye world"

# Expected Output
Generated 2 embeddings
Model: nomic-embed-text-v1.5
Total tokens: X

Text 1: Hello world
  Embedding dimensions: 768
  First 5 values: [0.123, -0.456, ...]

# Verify
- Both texts embedded
- Dimension count correct (768 for nomic-embed-text)
- JSON output works with --json flag
```

### Model Listing
```bash
# Input
syn model list

# Expected Output
Available models: 20+

  deepseek-v3.2
    Owner: Fireworks
    Created: 2025-XX-XX

  minimax-m2.1
    Owner: Synthetic
    ...

# Verify
- All models returned
- Metadata displayed
- No errors
```

### Interactive Chat
```bash
# Input
syn chat
you> What is AI?
assistant> [response]
you> Tell me more
assistant> [contextual response]

# Verify
- Context maintained across exchanges
- /clear command works
- /model shows current model
- Ctrl+C exits gracefully
```

---

## Acceptance Criteria

- [ ] One-shot mode works: `syn "prompt"`
- [ ] Stdin support works: `echo "text" | syn`
- [ ] File support works: `syn -f file.txt "explain"`
- [ ] JSON output works: `syn "test" --json`
- [ ] Interactive chat maintains context
- [ ] Model listing shows all available models
- [ ] Embeddings generate correct vectors
- [ ] Configuration file loaded from ~/.config/syn/config.yaml
- [ ] Environment variable SYN_API_KEY overrides config
- [ ] Retry logic handles transient failures
- [ ] Verbose mode shows debug logs
- [ ] Error messages are clear and actionable

---

## Known Limitations

1. **No streaming support**: API supports it, but not implemented in v1
2. **No Anthropic endpoint**: Only OpenAI-compatible endpoint used
3. **No vision/audio/video**: Not available in Synthetic's OpenAI endpoint
4. **No web reader/search**: Z.AI-specific features not replicated
5. **Context limit**: REPL keeps only last 20 messages (configurable)
6. **No persistent history**: History not saved across sessions (P2 feature)

---

## Migration Guide (from zai to syn)

| zai Command | syn Equivalent | Notes |
|-------------|----------------|-------|
| `zai "prompt"` | `syn "prompt"` | Same |
| `zai chat` | `syn chat` | Same |
| `zai search "query"` | ✗ Not available | Use external search tools |
| `zai reader <url>` | ✗ Not available | Use curl/wget |
| `zai image "prompt"` | ✗ Not available | Use dedicated image generation APIs |
| `zai vision -f img.jpg` | ✗ Not available | Use vision-specific services |
| `zai audio -f file.wav` | ✗ Not available | Use whisper/ASR services |
| `zai video "prompt"` | ✗ Not available | Use video generation services |
| `zai model list` | `syn model list` | Same |
| N/A | `syn embed "text"` | New feature |

---

## References

**Source Materials:**
- zai repository: ~/go/src/zai
- Synthetic.new API docs: https://dev.synthetic.new/docs/api/overview
- Synthetic.new models: https://dev.synthetic.new/docs/api/models
- OpenAI API compatibility: Standard endpoints

**External Dependencies:**
- Cobra: CLI framework
- Viper: Configuration management
- Lipgloss: Terminal UI styling
- Go standard library: http, context, encoding/json

---

## Sources

- [API Overview | Synthetic](https://dev.synthetic.new/docs/api/overview)
- [Synthetics REST API | New Relic Documentation](https://docs.newrelic.com/docs/apis/synthetics-rest-api/monitor-examples/manage-synthetics-monitors-rest-api/)
