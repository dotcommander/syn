package app

import (
	"fmt"
	"maps"
	"time"
)

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
	Role    string `json:"role"` // "user", "assistant", "system"
	Content string `json:"content"`
}

// ChatRequest represents the /chat/completions API request.
type ChatRequest struct {
	Model         string         `json:"model"`
	Messages      []Message      `json:"messages"`
	Temperature   float64        `json:"temperature,omitempty"`
	MaxTokens     int            `json:"max_tokens,omitempty"`
	TopP          float64        `json:"top_p,omitempty"`
	Stream        bool           `json:"stream,omitempty"`
	StreamOptions *StreamOptions `json:"stream_options,omitempty"`
}

// StreamOptions configures streaming behavior.
type StreamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

// StreamChunk represents a single SSE chunk from a streaming response.
type StreamChunk struct {
	ID      string         `json:"id"`
	Choices []StreamChoice `json:"choices"`
	Usage   *Usage         `json:"usage,omitempty"`
}

// StreamChoice represents a choice delta in a streaming chunk.
type StreamChoice struct {
	Index int         `json:"index"`
	Delta StreamDelta `json:"delta"`
}

// StreamDelta represents incremental content in a streaming response.
type StreamDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// StreamResult contains the assembled result of a streaming chat request.
type StreamResult struct {
	Content string
	Usage   Usage
	TTFMS   int64 // time to first token in milliseconds
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

// EmbeddingRequest represents the /embeddings API request.
type EmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"` // Array of strings to embed
}

// EmbeddingResponse represents the /embeddings API response.
type EmbeddingResponse struct {
	Object string          `json:"object"`
	Data   []EmbeddingData `json:"data"`
	Model  string          `json:"model"`
	Usage  EmbeddingUsage  `json:"usage"`
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

// modelAliases maps short names to full Synthetic model IDs.
var modelAliases = map[string]string{ //nolint:gochecknoglobals // read-only lookup table, idiomatic Go
	"gptoss":   "hf:openai/gpt-oss-120b",
	"gpt":      "hf:openai/gpt-oss-120b",
	"kimi":     "hf:moonshotai/Kimi-K2-Thinking",
	"qwen":     "hf:Qwen/Qwen3-VL-235B-A22B-Instruct",
	"glm":      "hf:zai-org/GLM-4.7",
	"zai":      "hf:zai-org/GLM-4.7",
	"deepseek": "hf:deepseek-ai/DeepSeek-V3.2",
	"ds":       "hf:deepseek-ai/DeepSeek-V3.2",
}

// ModelAliases returns a copy of the model alias map.
func ModelAliases() map[string]string {
	result := make(map[string]string, len(modelAliases))
	maps.Copy(result, modelAliases)
	return result
}

// ResolveModel resolves a model alias to its full name, or returns the input if not an alias.
func ResolveModel(model string) string {
	if resolved, ok := modelAliases[model]; ok {
		return resolved
	}
	return model
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

// SearchRequest represents the /v2/search API request.
type SearchRequest struct {
	Query string `json:"query"`
}

// SearchResponse represents the /v2/search API response.
type SearchResponse struct {
	Results []SearchResult `json:"results"`
}

// SearchResult represents a single search result.
type SearchResult struct {
	Title     string `json:"title"`
	URL       string `json:"url"`
	Snippet   string `json:"snippet"`
	Published string `json:"published,omitempty"` // ISO 8601 timestamp
}
