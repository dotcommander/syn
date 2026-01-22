# Technical Debt Inventory - SYN CLI

**Date:** 2026-01-22
**Debt Score:** 42/100 (Lower is better)

---

## Priority Matrix

| ID | Issue | Priority | Impact | Effort | Score |
|----|-------|----------|--------|--------|-------|
| D1 | Hardcoded API key in config | P0 | Critical | 1h | 100 |
| D2 | Zero unit test coverage | P0 | Critical | 16h | 95 |
| D3 | Goroutine leak in chat input | P0 | High | 2h | 90 |
| D4 | No input validation | P1 | High | 4h | 80 |
| D5 | Fragile string-based retry detection | P1 | Medium | 2h | 70 |
| D6 | Missing file size limits | P1 | High | 2h | 75 |
| D7 | No rate limiting | P1 | Medium | 4h | 65 |
| D8 | root.go violates SRP (296 LOC) | P2 | Low | 4h | 50 |
| D9 | Fragile search URL construction | P2 | Medium | 1h | 60 |
| D10 | No circuit breaker pattern | P2 | Low | 6h | 45 |
| D11 | Logger not interface-based | P2 | Low | 2h | 40 |
| D12 | No HTTP connection pooling config | P2 | Low | 1h | 35 |
| D13 | Context truncation without full view | P3 | Low | 2h | 30 |
| D14 | No custom model aliases | P3 | Low | 3h | 25 |
| D15 | No streaming response support | P3 | Low | 8h | 20 |

**Debt Categories:**
- **Security:** D1, D4, D6
- **Reliability:** D3, D5, D7, D10
- **Testability:** D2, D11
- **Maintainability:** D8, D9
- **Performance:** D12, D15
- **UX:** D13, D14

---

## P0 - Critical (Fix Before Production)

### D1: Hardcoded API Key

**Location:** `internal/config/config.go:12`

**Issue:**
```go
viper.SetDefault("api.key", "syn_2afa3d6ae1d48878694a13cbbe35d76c")
```

**Impact:**
- ✅ API key in git history (public exposure)
- ✅ Compromised key can't be rotated without code change
- ✅ Key visible in compiled binary (strings dump)
- ✅ Security audit failure

**Root Cause:** Convenience for testing/demo

**Fix:**
```go
// config.go
func SetDefaults() {
    // Remove hardcoded key
    viper.SetDefault("api.key", "")
}

// root.go:207
func initConfig() error {
    // ... existing code ...

    key := viper.GetString("api.key")
    if key == "" {
        return fmt.Errorf(`API key required. Set via:
  - Environment: export SYN_API_KEY="your_key"
  - Config file: ~/.config/syn/config.yaml`)
    }

    // Validate key format
    if !strings.HasPrefix(key, "syn_") {
        return fmt.Errorf("invalid API key format (expected syn_*)")
    }

    return nil
}
```

**Testing:**
```bash
# Should fail
./syn "test" 2>&1 | grep "API key required"

# Should succeed
export SYN_API_KEY="syn_test"
./syn "test"
```

**Estimated Effort:** 1 hour
**Debt Score:** 100 (Critical)

---

### D2: Zero Unit Test Coverage

**Location:** All packages

**Issue:**
- No tests for `internal/app/client.go` (613 LOC)
- No tests for retry logic
- No tests for model resolution
- Only doc examples exist

**Impact:**
- ✅ Refactoring dangerous (no safety net)
- ✅ Bugs undetected until production
- ✅ Confidence low for changes
- ✅ Regression risk high

**Coverage Gaps:**

| Component | LOC | Tests | Coverage |
|-----------|-----|-------|----------|
| client.go | 613 | 0 | 0% |
| types.go | 178 | 0 | 0% |
| config.go | 27 | 0 | 0% |
| cmd/*.go | 1111 | 0 | 0% |

**Fix Plan:**

#### Phase 1: Core Logic (8h)

```go
// internal/app/client_test.go
package app_test

import (
    "bytes"
    "context"
    "io"
    "net/http"
    "testing"

    "github.com/dotcommander/syn/internal/app"
    "github.com/stretchr/testify/assert"
)

type mockHTTPClient struct {
    DoFunc func(*http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
    return m.DoFunc(req)
}

func TestClient_Chat_Success(t *testing.T) {
    mockResp := `{
        "choices": [{"message": {"content": "test response"}}],
        "usage": {"total_tokens": 10}
    }`

    mock := &mockHTTPClient{
        DoFunc: func(req *http.Request) (*http.Response, error) {
            return &http.Response{
                StatusCode: 200,
                Body:       io.NopCloser(bytes.NewBufferString(mockResp)),
            }, nil
        },
    }

    cfg := app.ClientConfig{
        APIKey:  "test_key",
        BaseURL: "https://api.test.com/v1",
        Model:   "test-model",
    }

    client := app.NewClient(cfg, app.NewLogger(false), mock)
    resp, err := client.Chat(context.Background(), "test", app.DefaultChatOptions())

    assert.NoError(t, err)
    assert.Equal(t, "test response", resp)
}

func TestClient_Chat_APIError(t *testing.T) {
    mock := &mockHTTPClient{
        DoFunc: func(req *http.Request) (*http.Response, error) {
            return &http.Response{
                StatusCode: 500,
                Body:       io.NopCloser(bytes.NewBufferString("internal error")),
            }, nil
        },
    }

    client := app.NewClient(app.ClientConfig{APIKey: "test"}, app.NewLogger(false), mock)
    _, err := client.Chat(context.Background(), "test", app.DefaultChatOptions())

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "500")
}

func TestRetryLogic(t *testing.T) {
    attempts := 0
    mock := &mockHTTPClient{
        DoFunc: func(req *http.Request) (*http.Response, error) {
            attempts++
            if attempts < 3 {
                return &http.Response{
                    StatusCode: 429,
                    Body:       io.NopCloser(bytes.NewBufferString("rate limit")),
                }, nil
            }
            return &http.Response{
                StatusCode: 200,
                Body:       io.NopCloser(bytes.NewBufferString(`{"choices":[{"message":{"content":"ok"}}],"usage":{}}`)),
            }, nil
        },
    }

    cfg := app.ClientConfig{
        APIKey:  "test",
        BaseURL: "https://test.com/v1",
        RetryConfig: app.RetryConfig{
            MaxAttempts:    3,
            InitialBackoff: 1 * time.Millisecond,
            MaxBackoff:     10 * time.Millisecond,
        },
    }

    client := app.NewClient(cfg, app.NewLogger(false), mock)
    _, err := client.Chat(context.Background(), "test", app.DefaultChatOptions())

    assert.NoError(t, err)
    assert.Equal(t, 3, attempts, "should retry twice")
}
```

#### Phase 2: Helper Functions (4h)

```go
func TestResolveModel(t *testing.T) {
    tests := []struct {
        input string
        want  string
    }{
        {"kimi", "hf:moonshotai/Kimi-K2-Thinking"},
        {"gpt", "hf:openai/gpt-oss-120b"},
        {"unknown", "unknown"},
        {"", ""},
    }

    for _, tt := range tests {
        t.Run(tt.input, func(t *testing.T) {
            got := app.ResolveModel(tt.input)
            assert.Equal(t, tt.want, got)
        })
    }
}

func TestCalculateBackoff(t *testing.T) {
    // Test exponential growth
    b1 := calculateBackoff(1, 1*time.Second, 30*time.Second)
    b2 := calculateBackoff(2, 1*time.Second, 30*time.Second)
    b3 := calculateBackoff(3, 1*time.Second, 30*time.Second)

    assert.True(t, b2 > b1, "backoff should increase")
    assert.True(t, b3 > b2, "backoff should increase")

    // Test max cap
    b10 := calculateBackoff(10, 1*time.Second, 5*time.Second)
    assert.LessOrEqual(t, b10, 6*time.Second, "should respect max + jitter")
}
```

#### Phase 3: Integration Tests (4h)

```go
func TestClient_IntegrationSearch(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    client := app.NewClient(/* real config */, app.NewLogger(true), nil)
    resp, err := client.Search(context.Background(), "golang testing")

    assert.NoError(t, err)
    assert.NotEmpty(t, resp.Results)
}
```

**Test Coverage Goal:** 80% for `internal/app/`, 60% for `cmd/`

**Estimated Effort:** 16 hours
**Debt Score:** 95

---

### D3: Goroutine Leak in Chat Input

**Location:** `cmd/chat.go:78-90`

**Issue:**
```go
for {
    go func() {
        if scanner.Scan() {
            inputCh <- inputResult{text: scanner.Text()}
        } else {
            inputCh <- inputResult{err: scanner.Err()}
        }
    }()

    select {
    case <-ctx.Done():
        return nil // Goroutine still running!
    case result := <-inputCh:
        // ...
    }
}
```

**Problem:**
- Goroutine spawned on each iteration
- If context cancelled, goroutine leaks
- `scanner.Scan()` blocks until input, never respects context

**Impact:**
- ✅ Memory leak on Ctrl-C
- ✅ Goroutine accumulation
- ✅ Resource exhaustion in long sessions

**Fix:**
```go
import "golang.org/x/sync/errgroup"

func runInteractiveChat() error {
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    g, ctx := errgroup.WithContext(ctx)
    inputCh := make(chan inputResult, 1)

    // Single goroutine for input reading
    g.Go(func() error {
        scanner := bufio.NewScanner(os.Stdin)
        for scanner.Scan() {
            select {
            case <-ctx.Done():
                return ctx.Err()
            case inputCh <- inputResult{text: scanner.Text()}:
            }
        }
        if err := scanner.Err(); err != nil {
            return err
        }
        return nil
    })

    for {
        select {
        case <-ctx.Done():
            g.Wait() // Ensure goroutine cleanup
            return nil
        case result := <-inputCh:
            // ... process input
        }
    }
}
```

**Testing:**
```bash
# Before fix
go test -run TestChatGoroutineLeak -count 100 -race

# Should detect race or leak
```

**Estimated Effort:** 2 hours
**Debt Score:** 90

---

## P1 - High Priority

### D4: No Input Validation

**Locations:**
- `cmd/root.go:254` - File path from `-f` flag
- `cmd/vision.go:463` - Image file reading
- `cmd/search.go:70` - Search query
- `internal/app/client.go:394` - Embed text array

**Issue:**
- No check if file exists/readable
- No URL validation
- No size limits
- No array length limits

**Impact:**
- ✅ Unclear error messages
- ✅ Potential DoS (huge files)
- ✅ Command injection risk
- ✅ Poor UX

**Fix:**
```go
// internal/validation/validation.go
package validation

import (
    "fmt"
    "net/url"
    "os"
)

func ValidateFilePath(path string) error {
    if path == "" {
        return nil
    }

    info, err := os.Stat(path)
    if err != nil {
        return fmt.Errorf("file not accessible: %w", err)
    }

    if info.IsDir() {
        return fmt.Errorf("path is directory, expected file: %s", path)
    }

    // 10MB limit
    if info.Size() > 10*1024*1024 {
        return fmt.Errorf("file too large: %d bytes (max 10MB)", info.Size())
    }

    return nil
}

func ValidateURL(rawURL string) error {
    u, err := url.Parse(rawURL)
    if err != nil {
        return fmt.Errorf("invalid URL: %w", err)
    }

    if u.Scheme != "http" && u.Scheme != "https" {
        return fmt.Errorf("invalid URL scheme: %s (expected http/https)", u.Scheme)
    }

    return nil
}

func ValidateTextArray(texts []string, maxLen int) error {
    if len(texts) == 0 {
        return fmt.Errorf("empty text array")
    }

    if len(texts) > maxLen {
        return fmt.Errorf("too many texts: %d (max %d)", len(texts), maxLen)
    }

    for i, text := range texts {
        if len(text) > 100000 {
            return fmt.Errorf("text[%d] too long: %d chars (max 100k)", i, len(text))
        }
    }

    return nil
}
```

**Usage:**
```go
// cmd/root.go
opts.FilePath = viper.GetString("file")
if err := validation.ValidateFilePath(opts.FilePath); err != nil {
    return fmt.Errorf("file validation failed: %w", err)
}

// cmd/vision.go
imageSource := args[0]
if strings.HasPrefix(imageSource, "http") {
    if err := validation.ValidateURL(imageSource); err != nil {
        return err
    }
} else {
    if err := validation.ValidateFilePath(imageSource); err != nil {
        return err
    }
}

// internal/app/client.go
if err := validation.ValidateTextArray(texts, 100); err != nil {
    return nil, fmt.Errorf("invalid texts: %w", err)
}
```

**Estimated Effort:** 4 hours
**Debt Score:** 80

---

### D5: Fragile String-Based Retry Detection

**Location:** `internal/app/client.go:299-329`

**Issue:**
```go
retryablePatterns := []string{
    "connection refused",
    "timeout",
    "429", "503",
}

for _, pattern := range retryablePatterns {
    if strings.Contains(strings.ToLower(errStr), pattern) {
        return true
    }
}
```

**Problems:**
- String matching fragile (language changes, error message updates)
- No type safety
- HTTP status in string form
- Can't distinguish transient vs permanent errors

**Fix:**
```go
func isRetryableError(err error) bool {
    if err == nil {
        return false
    }

    // Check API errors first
    var apiErr *APIError
    if errors.As(err, &apiErr) {
        switch apiErr.StatusCode {
        case http.StatusTooManyRequests,      // 429
             http.StatusBadGateway,           // 502
             http.StatusServiceUnavailable,   // 503
             http.StatusGatewayTimeout:       // 504
            return true
        default:
            return false
        }
    }

    // Check network errors
    var netErr interface{ Timeout() bool }
    if errors.As(err, &netErr) && netErr.Timeout() {
        return true
    }

    // Check specific network errors
    var opErr *net.OpError
    if errors.As(err, &opErr) {
        if opErr.Temporary() || opErr.Timeout() {
            return true
        }
    }

    return false
}
```

**Testing:**
```go
func TestIsRetryableError(t *testing.T) {
    tests := []struct {
        err  error
        want bool
    }{
        {&APIError{StatusCode: 429}, true},
        {&APIError{StatusCode: 503}, true},
        {&APIError{StatusCode: 404}, false},
        {&APIError{StatusCode: 401}, false},
        {context.DeadlineExceeded, true},
        {fmt.Errorf("random error"), false},
    }

    for _, tt := range tests {
        got := isRetryableError(tt.err)
        assert.Equal(t, tt.want, got)
    }
}
```

**Estimated Effort:** 2 hours
**Debt Score:** 70

---

### D6: Missing File Size Limits

**Location:** `internal/app/client.go:129, 463`

**Issue:**
```go
// No size check before reading
data, err := os.ReadFile(filePath)

// Image encoding without limit
data, err := os.ReadFile(imageSource)
imageURL = fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(data))
```

**Impact:**
- ✅ OOM on large files (multi-GB images)
- ✅ DoS vulnerability
- ✅ Poor error message ("out of memory")

**Fix:**
```go
// internal/app/client.go
const (
    MaxFileSize  = 10 * 1024 * 1024  // 10MB
    MaxImageSize = 20 * 1024 * 1024  // 20MB
)

func readFileWithLimit(path string, maxSize int64) ([]byte, error) {
    info, err := os.Stat(path)
    if err != nil {
        return nil, fmt.Errorf("stat file: %w", err)
    }

    if info.Size() > maxSize {
        return nil, fmt.Errorf("file too large: %d bytes (max %d)", info.Size(), maxSize)
    }

    return os.ReadFile(path)
}

// Usage
func (c *Client) buildContent(prompt, filePath string) (string, error) {
    if filePath == "" {
        return prompt, nil
    }

    data, err := readFileWithLimit(filePath, MaxFileSize)
    if err != nil {
        return "", fmt.Errorf("failed to read file %s: %w", filePath, err)
    }

    return fmt.Sprintf("%s\n\nFile contents (%s):\n```\n%s\n```", prompt, filePath, string(data)), nil
}

func (c *Client) Vision(...) {
    // ...
    data, err := readFileWithLimit(imageSource, MaxImageSize)
    if err != nil {
        return "", fmt.Errorf("failed to read image: %w", err)
    }
    // ...
}
```

**Estimated Effort:** 2 hours
**Debt Score:** 75

---

### D7: No Rate Limiting

**Location:** `internal/app/client.go` (missing)

**Issue:**
- Client can send unlimited requests
- No local rate limiting
- API may ban aggressive clients
- No backpressure mechanism

**Impact:**
- ✅ API rate limit bans
- ✅ Cost overruns (pay-per-token)
- ✅ Poor multi-user support

**Fix:**
```go
import "golang.org/x/time/rate"

type Client struct {
    config     ClientConfig
    httpClient HTTPDoer
    logger     *slog.Logger
    limiter    *rate.Limiter // New field
}

func NewClient(cfg ClientConfig, logger *slog.Logger, httpClient HTTPDoer) *Client {
    // ...
    limiter := rate.NewLimiter(rate.Limit(cfg.RateLimit), cfg.RateBurst)

    return &Client{
        config:     cfg,
        httpClient: httpClient,
        logger:     logger,
        limiter:    limiter,
    }
}

func (c *Client) doRequest(ctx context.Context, ...) (string, Usage, error) {
    // Wait for rate limit
    if err := c.limiter.Wait(ctx); err != nil {
        return "", Usage{}, fmt.Errorf("rate limit wait: %w", err)
    }

    // ... existing code
}
```

**Config:**
```go
// config/config.go
viper.SetDefault("api.rate_limit", 10)   // 10 req/sec
viper.SetDefault("api.rate_burst", 5)    // Burst of 5

// types.go
type ClientConfig struct {
    // ...
    RateLimit float64
    RateBurst int
}
```

**Estimated Effort:** 4 hours
**Debt Score:** 65

---

## P2 - Medium Priority

### D8: root.go Violates SRP (296 LOC)

**Location:** `cmd/root.go`

**Issue:**
- Mixes 5 responsibilities:
  1. Flag/command setup
  2. Config loading
  3. Client creation
  4. Help rendering
  5. One-shot execution

**Current Structure:**
```
root.go (296 LOC)
├── Flag definitions (20 LOC)
├── Help rendering (70 LOC)
├── Config init (30 LOC)
├── Client factory (20 LOC)
├── Stdin handling (30 LOC)
└── One-shot execution (50 LOC)
```

**Refactor:**
```
cmd/
├── root.go (80 LOC)       # Command setup only
├── help.go (70 LOC)       # Styled help
├── factory.go (40 LOC)    # Client creation
└── oneshot.go (60 LOC)    # One-shot mode

internal/config/
├── config.go (50 LOC)     # Viper setup
└── loader.go (40 LOC)     # Config loading
```

**Benefits:**
- Easier to test each component
- Clear file-per-responsibility
- Reduced cognitive load

**Estimated Effort:** 4 hours
**Debt Score:** 50

---

### D9: Fragile Search URL Construction

**Location:** `internal/app/client.go:579`

**Issue:**
```go
url := strings.Replace(c.config.BaseURL, "/openai/v1", "/v2/search", 1)
```

**Problems:**
- Breaks if base URL format changes
- Magic string "/openai/v1"
- No validation

**Fix:**
```go
// config/config.go
viper.SetDefault("api.search_url", "https://api.synthetic.new/v2/search")

// types.go
type ClientConfig struct {
    BaseURL      string
    SearchURL    string // New field
    // ...
}

// client.go
func (c *Client) Search(ctx context.Context, query string) (*SearchResponse, error) {
    // ...
    url := c.config.SearchURL
    // ...
}
```

**Estimated Effort:** 1 hour
**Debt Score:** 60

---

### D10: No Circuit Breaker Pattern

**Location:** `internal/app/client.go` (missing)

**Issue:**
- Retries indefinitely across requests
- No failure threshold
- No automatic recovery detection
- Cascading failures possible

**Impact:**
- ✅ Wasted retries on dead API
- ✅ Slow failure detection
- ✅ Resource exhaustion

**Fix:**
```go
import "github.com/sony/gobreaker"

type Client struct {
    config     ClientConfig
    httpClient HTTPDoer
    logger     *slog.Logger
    cb         *gobreaker.CircuitBreaker // New
}

func NewClient(cfg ClientConfig, logger *slog.Logger, httpClient HTTPDoer) *Client {
    cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
        Name:        "synthetic-api",
        MaxRequests: 3,
        Interval:    10 * time.Second,
        Timeout:     60 * time.Second,
        ReadyToTrip: func(counts gobreaker.Counts) bool {
            failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
            return counts.Requests >= 3 && failureRatio >= 0.6
        },
    })

    return &Client{
        config:     cfg,
        httpClient: httpClient,
        logger:     logger,
        cb:         cb,
    }
}

func (c *Client) doRequest(...) (string, Usage, error) {
    result, err := c.cb.Execute(func() (interface{}, error) {
        // ... existing request logic
        return response, nil
    })

    if err != nil {
        return "", Usage{}, err
    }

    return result.(string), usage, nil
}
```

**Estimated Effort:** 6 hours
**Debt Score:** 45

---

### D11: Logger Not Interface-Based

**Location:** `internal/app/client.go:54,76`

**Issue:**
```go
type Client struct {
    logger *slog.Logger // Concrete type
}

func NewLogger(verbose bool) *slog.Logger { // Concrete return
    // ...
}
```

**Impact:**
- ✅ Can't mock logger in tests
- ✅ Tight coupling to slog
- ✅ Hard to test log output

**Fix:**
```go
// internal/app/logger.go
package app

type Logger interface {
    Debug(msg string, args ...any)
    Info(msg string, args ...any)
    Error(msg string, args ...any)
}

type slogAdapter struct {
    logger *slog.Logger
}

func (s *slogAdapter) Debug(msg string, args ...any) {
    s.logger.Debug(msg, args...)
}

func (s *slogAdapter) Info(msg string, args ...any) {
    s.logger.Info(msg, args...)
}

func (s *slogAdapter) Error(msg string, args ...any) {
    s.logger.Error(msg, args...)
}

func NewLogger(verbose bool) Logger {
    level := slog.LevelInfo
    if verbose {
        level = slog.LevelDebug
    }
    opts := &slog.HandlerOptions{Level: level}
    return &slogAdapter{
        logger: slog.New(slog.NewTextHandler(os.Stderr, opts)),
    }
}

// Testing
type mockLogger struct {
    DebugCalls []string
}

func (m *mockLogger) Debug(msg string, args ...any) {
    m.DebugCalls = append(m.DebugCalls, msg)
}
```

**Estimated Effort:** 2 hours
**Debt Score:** 40

---

## P3 - Low Priority

### D12: No HTTP Connection Pooling Config

**Location:** `internal/app/client.go:65`

**Issue:**
```go
httpClient = &http.Client{Timeout: timeout}
// No Transport configuration
```

**Impact:**
- Uses default connection pool (unlimited)
- Potential resource exhaustion
- No keep-alive tuning

**Fix:**
```go
if httpClient == nil {
    transport := &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        MaxConnsPerHost:     50,
        IdleConnTimeout:     90 * time.Second,
        DisableCompression:  false,
        ForceAttemptHTTP2:   true,
    }

    httpClient = &http.Client{
        Timeout:   timeout,
        Transport: transport,
    }
}
```

**Estimated Effort:** 1 hour
**Debt Score:** 35

---

### D13: Context Truncation Without Full View

**Location:** `cmd/chat.go:262-269`

**Issue:**
```go
func truncateString(s string, maxLen int) string {
    // No way to see full content
    if len(s) <= maxLen {
        return s
    }
    return s[:maxLen-3] + "..."
}
```

**Impact:**
- Can't review full conversation history
- Truncated messages hard to debug

**Fix:**
```go
// Add command
case "/context", "/ctx":
    full, _ := cmd.Flags().GetBool("full")
    printContextStyled(*context, full)
    return true

func printContextStyled(ctx []app.Message, full bool) {
    for _, msg := range ctx {
        content := msg.Content
        if !full && len(content) > 100 {
            content = truncateString(content, 100)
        }
        fmt.Printf("  %s %s\n", styledRole, theme.Dim.Render(content))
    }
}
```

**Estimated Effort:** 2 hours
**Debt Score:** 30

---

### D14: No Custom Model Aliases

**Location:** `internal/app/types.go:129`

**Issue:**
- Aliases hardcoded
- Can't add user-defined aliases
- No per-project aliases

**Fix:**
```go
// config/config.go
viper.SetDefault("model.aliases", map[string]string{})

// types.go
func GetAllAliases() map[string]string {
    aliases := make(map[string]string)

    // Built-in aliases
    for k, v := range ModelAliases {
        aliases[k] = v
    }

    // User aliases from config
    userAliases := viper.GetStringMapString("model.aliases")
    for k, v := range userAliases {
        aliases[k] = v
    }

    return aliases
}

func ResolveModel(model string) string {
    aliases := GetAllAliases()
    if resolved, ok := aliases[model]; ok {
        return resolved
    }
    return model
}
```

**Config:**
```yaml
# ~/.config/syn/config.yaml
model:
  aliases:
    fast: hf:some/fast-model
    accurate: hf:some/accurate-model
```

**Estimated Effort:** 3 hours
**Debt Score:** 25

---

### D15: No Streaming Response Support

**Location:** `internal/app/client.go:173`

**Issue:**
```go
reqData := ChatRequest{
    Stream: false, // Hardcoded
}
```

**Impact:**
- Long responses feel slow
- No progressive output
- Poor UX for chat mode

**Fix:**
```go
// types.go
type ChatOptions struct {
    // ...
    Stream bool
}

// client.go
func (c *Client) ChatStream(ctx context.Context, prompt string, opts ChatOptions, handler func(string)) error {
    reqData := ChatRequest{
        Model:    ResolveModel(c.config.Model),
        Messages: messages,
        Stream:   true,
    }

    // ... send request

    reader := bufio.NewReader(resp.Body)
    for {
        line, err := reader.ReadBytes('\n')
        if err == io.EOF {
            break
        }
        if err != nil {
            return err
        }

        // Parse SSE
        if bytes.HasPrefix(line, []byte("data: ")) {
            data := bytes.TrimPrefix(line, []byte("data: "))
            // ... unmarshal chunk
            handler(chunk.Content)
        }
    }

    return nil
}
```

**Estimated Effort:** 8 hours
**Debt Score:** 20

---

## Debt Metrics Summary

### By Priority

| Priority | Items | Total Effort | Avg Debt Score |
|----------|-------|--------------|----------------|
| P0 | 3 | 19h | 95 |
| P1 | 4 | 14h | 73 |
| P2 | 4 | 13h | 49 |
| P3 | 4 | 14h | 28 |

### By Category

| Category | Items | Debt Score |
|----------|-------|------------|
| Security | 3 | 252 |
| Reliability | 4 | 270 |
| Testability | 2 | 135 |
| Maintainability | 2 | 110 |
| Performance | 2 | 55 |
| UX | 2 | 55 |

### Resolution Plan

**Week 1: Critical Issues**
- Day 1: D1 (API key removal)
- Day 2-3: D2 (Unit tests foundation)
- Day 4: D3 (Goroutine leak)
- Day 5: D4 (Input validation)

**Week 2: High Priority**
- Day 1: D5 (Retry detection)
- Day 2: D6 (File size limits)
- Day 3-4: D7 (Rate limiting)

**Week 3: Medium Priority**
- Day 1: D8 (root.go refactor)
- Day 2: D9 (URL construction)
- Day 3-4: D10 (Circuit breaker)
- Day 5: D11 (Logger interface)

**Week 4: Low Priority**
- Day 1: D12 (Connection pooling)
- Day 2: D13 (Context view)
- Day 3: D14 (Custom aliases)
- Day 4-5: D15 (Streaming support)

**Total Estimated Effort:** 60 hours (3 weeks @ 4h/day)

---

## Debt Velocity

**Current Accumulation Rate:** Low (codebase stable)
**Paydown Rate:** 0 (no active work)
**Net Change:** +5 debt/month (feature additions)

**Target State:**
- P0 debt: 0
- P1 debt: <10
- P2 debt: <30
- P3 debt: Acceptable

**Recommendation:** Allocate 20% of dev time to debt paydown
