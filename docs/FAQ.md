# Frequently Asked Questions

Common questions about using the Syn CLI and API client.

## How do I get started with the Syn CLI?

First, install the tool and set your API key:

```bash
export SYN_API_KEY="your-api-key"
syn chat "Hello, how are you?"
```

For detailed installation and setup instructions, see [API Reference - app](./api/app.md).

## What models are supported?

Syn supports all models available through the Synthetic.new API. To list available models:

```bash
syn models
```

Or programmatically:

```go
models, err := client.ListModels(ctx)
```

See [API Reference - app](./api/app.md) for more details on model listing.

## How do I configure retry behavior?

Retry is automatic for transient failures. To customize retry settings:

```go
retry := &app.RetryConfig{
    MaxAttempts:    5,
    InitialBackoff: 2 * time.Second,
    MaxBackoff:     60 * time.Second,
}
client := app.NewClient(
    app.WithAPIKey(apiKey),
    app.WithRetryConfig(retry),
)
```

For default values and configuration options, see [API Reference - config](./api/config.md).

## Can I use Syn with local image files?

Yes, the `Vision` method supports both URLs and local file paths:

```go
// Local file - automatically base64 encoded
result, err := client.Vision(ctx, "Describe this image", "./photo.jpg", opts)

// URL - used directly
result, err := client.Vision(ctx, "Describe this image", "https://example.com/photo.jpg", opts)
```

See [API Reference - app](./api/app.md) for complete vision API documentation.

## Why am I getting authentication errors?

Authentication errors typically occur when:

1. The `SYN_API_KEY` environment variable is not set
2. The API key is invalid or expired
3. The API key format is incorrect

Verify your setup:

```bash
echo $SYN_API_KEY  # Should show your key
syn models  # Test authentication
```

For configuration details, see [API Reference - config](./api/config.md).

## How do I add error handling to my code?

Always check returned errors:

```go
response, err := client.Chat(ctx, prompt, opts)
if err != nil {
    // Handle error - may include retry info
    return fmt.Errorf("chat failed: %w", err)
}
```

For comprehensive error handling patterns, see the [Error Handling Guide](./guides/errors.md).
