package app

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"os"
	"path/filepath"
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

// EmbeddingClient interface for embeddings (ISP compliance).
type EmbeddingClient interface {
	Embed(ctx context.Context, texts []string, model string) (*EmbeddingResponse, error)
}

// VisionClient interface for vision/image analysis (ISP compliance).
type VisionClient interface {
	Vision(ctx context.Context, prompt string, imageSource string, opts ChatOptions) (string, error)
}

// SearchClient interface for web search (ISP compliance).
type SearchClient interface {
	Search(ctx context.Context, query string) (*SearchResponse, error)
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
	messages := c.buildMessages(content)

	// Prepend context messages if provided
	if len(opts.Context) > 0 {
		messages = append(opts.Context, messages...)
	}

	return messages
}

// buildMessages constructs the messages array for the API.
func (c *Client) buildMessages(content string) []Message {
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
		Model:    ResolveModel(c.config.Model),
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
		reqData.Model = ResolveModel(opts.Model)
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

// Vision analyzes an image with a prompt using a vision-capable model.
// imageSource can be a URL (http/https) or a local file path.
func (c *Client) Vision(ctx context.Context, prompt string, imageSource string, opts ChatOptions) (string, error) {
	if err := c.requireAPIKey(); err != nil {
		return "", err
	}

	// Build image content based on source type
	var imageURL string
	if strings.HasPrefix(imageSource, "http://") || strings.HasPrefix(imageSource, "https://") {
		imageURL = imageSource
	} else {
		// Local file - encode as base64 data URI
		data, err := os.ReadFile(imageSource)
		if err != nil {
			return "", fmt.Errorf("failed to read image file: %w", err)
		}

		// Detect MIME type from extension
		mimeType := "image/jpeg"
		ext := strings.ToLower(filepath.Ext(imageSource))
		switch ext {
		case ".png":
			mimeType = "image/png"
		case ".gif":
			mimeType = "image/gif"
		case ".webp":
			mimeType = "image/webp"
		}

		imageURL = fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(data))
	}

	// Use Qwen3-VL model for vision
	model := ResolveModel("qwen")
	if opts.Model != "" {
		model = ResolveModel(opts.Model)
	}

	// Build multimodal message content
	content := []map[string]interface{}{
		{
			"type": "image_url",
			"image_url": map[string]string{
				"url": imageURL,
			},
		},
		{
			"type": "text",
			"text": prompt,
		},
	}

	reqData := map[string]interface{}{
		"model": model,
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": content,
			},
		},
		"max_tokens": 4096,
	}

	if opts.Temperature != nil {
		reqData["temperature"] = *opts.Temperature
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/chat/completions", c.config.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))

	c.logger.Debug("sending vision request", "url", url, "model", model)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", &APIError{StatusCode: resp.StatusCode, Body: string(body)}
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return chatResp.Choices[0].Message.Content, nil
}

// Search performs a web search using the /v2/search endpoint.
// Note: This API is under development and may have breaking changes.
func (c *Client) Search(ctx context.Context, query string) (*SearchResponse, error) {
	if err := c.requireAPIKey(); err != nil {
		return nil, err
	}

	if query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	reqData := SearchRequest{Query: query}
	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Search uses /v2/ prefix, different from chat endpoints
	url := strings.Replace(c.config.BaseURL, "/openai/v1", "/v2/search", 1)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))

	c.logger.Debug("sending search request", "url", url, "query", query)

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

	var searchResp SearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal search response: %w", err)
	}

	c.logger.Debug("search complete", "results", len(searchResp.Results))

	return &searchResp, nil
}
