# Advanced Usage

This guide demonstrates advanced patterns and features for experienced users who need more control and flexibility.

## Custom Retry Configuration

Handle transient failures with custom retry logic and exponential backoff.

```go
package main

import (
    "context"
    "fmt"
    "log/slog"
    "time"

    "github.com/vampire/syn/internal/app"
    "github.com/vampire/syn/internal/config"
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

## Conversational Context Management

Build multi-turn conversations with message history.

```go
package main

import (
    "context"
    "fmt"

    "github.com/vampire/syn/internal/app"
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

## File Analysis with Vision

Combine file reading with vision capabilities for document analysis.

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/vampire/syn/internal/app"
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

## Temperature Tuning for Different Use Cases

Adjust temperature to control response randomness.

```go
package main

import (
    "context"
    "fmt"

    "github.com/vampire/syn/internal/app"
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

## Batch Embeddings for Semantic Search

Generate embeddings for multiple texts at once.

```go
package main

import (
    "context"
    "fmt"

    "github.com/vampire/syn/internal/app"
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

## Context-Aware Chat with File Context

Include file contents as context for code review or documentation tasks.

```go
package main

import (
    "context"
    "fmt"

    "github.com/vampire/syn/internal/app"
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

## Configuration File Override

Override defaults with a custom configuration file.

Create `~/.config/syn/config.yaml`:

```yaml
# Custom API configuration
api:
  key: sk-your-custom-key
  base_url: https://custom-api.example.com/v1
  model: hf:Qwen/Qwen3-VL-235B-A22B-Instruct

  # Retry settings for production
  retry:
    max_attempts: 5
    initial_backoff: 2s
    max_backoff: 60s

# Chat defaults for your use case
chat:
  temperature: 0.7
  max_tokens: 4096
  top_p: 0.95

# Enable verbose logging
verbose: true
```

The configuration will be automatically loaded and override the package defaults.
