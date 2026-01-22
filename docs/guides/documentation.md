# Documentation Guide

[← Back to Guides](./)

This guide explains how to write effective doc comments in the Syn project, following Rust-inspired documentation patterns adapted for Go.

## Overview

Good documentation is essential for maintainable code. Doc comments should explain **what** the code does, **why** it exists, and **how** to use it.

## Doc Comment Syntax

In Go, doc comments are written as regular comments that appear immediately before the item they document.

```go
// This is a regular comment.

// This is a doc comment that documents the function below.
func DoSomething() {}
```

### Exported vs Unexported

Only **exported** (public) items require doc comments:

```go
// NewClient creates a new API client.
// This is exported (PascalCase) and must be documented.
func NewClient() *Client {}

// cleanup performs internal cleanup.
// This is unexported (camelCase) and documentation is optional.
func cleanup() error {}
```

## Doc Comment Structure

A well-structured doc comment consists of:

1. **Summary line** - One brief sentence explaining what the item does
2. **Description** - Extended explanation (optional, for complex items)
3. **Parameters** - Document inputs using "Param name:" format
4. **Returns** - Document return values using "Returns:" format
5. **Examples** - Usage examples (recommended for public APIs)

### Basic Example

```go
// Chat sends a prompt to the API and returns the response.
//
// The prompt is sent along with any configured context messages.
// If the request fails due to a transient error, it will be
// automatically retried with exponential backoff.
//
// Param ctx: Context for cancellation control
// Param prompt: The user prompt to send
// Param opts: Optional configuration for temperature, max tokens, etc.
// Returns: The assistant's response, or an error if the request fails
func (c *Client) Chat(ctx context.Context, prompt string, opts ChatOptions) (string, error) {
	// ...
}
```

### Simple Example (No Params)

```go
// ListModels fetches available models from the API.
//
// Returns a slice of available models, or an error if the request fails
func (c *Client) ListModels(ctx context.Context) ([]Model, error) {
	// ...
}
```

## Best Practices

### Summary Line

- Start with a capital letter
- End with a period
- Use the function/variable name as the subject
- Keep it to one line

```go
// ✅ GOOD
// NewClient creates a new API client with the given configuration.

// ❌ BAD
// creates a client  // lowercase, no period
// This function creates a new client  // redundant "This function"
```

### Description

- Add a blank line after the summary
- Explain behavior, edge cases, and important details
- Mention any side effects or requirements

```go
// ✅ GOOD
// Embed generates embeddings for the given texts.
//
// The texts are processed in a single batch request. If the model
// string is empty, the default embedding model from the config is used.
// Returns an error if no texts are provided.

// ❌ BAD
// Embed generates embeddings.
// The texts are processed.  // Too vague
```

### Parameters

- Use "Param name:" format for each parameter
- Explain the purpose, not just the type
- Document constraints (e.g., "must not be empty")

```go
// ✅ GOOD
// Param apiKey: The API key for authentication (must not be empty)
// Param baseURL: The base URL for API requests

// ❌ BAD
// Param apiKey: string  // Redundant type info
// Param s: The string  // Vague parameter name
```

### Returns

- Use "Returns:" to start the returns section
- For multiple return values, describe each one
- Mention conditions that cause errors

```go
// ✅ GOOD
// Returns: The created client, or an error if the API key is invalid

// Returns: The response string and usage stats, or an error if the request fails

// ❌ BAD
// Returns: error  // What about the success case?
```

### Examples

Include examples for complex or non-obvious APIs:

```go
// DefaultChatOptions returns sensible defaults for CLI usage.
//
// Example:
//
//	opts := DefaultChatOptions()
//	opts.Temperature = Float64Ptr(0.8)
//	response, err := client.Chat(ctx, "Hello", opts)
func DefaultChatOptions() ChatOptions {
	// ...
}
```

## Good vs Bad Examples

### Function Documentation

```go
// ✅ GOOD - Clear summary, explains behavior
// Vision analyzes an image with a prompt using a vision-capable model.
//
// The imageSource parameter can be either a URL (http/https) or
// a local file path. Local files are automatically encoded as base64.
//
// Param ctx: Context for cancellation control
// Param prompt: Text prompt describing what to analyze
// Param imageSource: URL or local file path to the image
// Param opts: Optional chat configuration
// Returns: The analysis result, or an error if the image cannot be read
func (c *Client) Vision(ctx context.Context, prompt string, imageSource string, opts ChatOptions) (string, error)

// ❌ BAD - Vague, missing parameter documentation
// Vision processes an image.
func (c *Client) Vision(ctx context.Context, prompt string, imageSource string, opts ChatOptions) (string, error)
```

### Type Documentation

```go
// ✅ GOOD - Explains purpose and field constraints
// RetryConfig configures retry behavior for transient failures.
//
// All fields are optional. Zero values will be replaced with defaults:
// - MaxAttempts defaults to 3
// - InitialBackoff defaults to 1s
// - MaxBackoff defaults to 30s
type RetryConfig struct {
	MaxAttempts    int           // Maximum number of retry attempts
	InitialBackoff time.Duration // Initial backoff duration
	MaxBackoff     time.Duration // Maximum backoff duration
}

// ❌ BAD - No explanation of defaults or behavior
type RetryConfig struct {
	MaxAttempts    int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
}
```

### Constant Documentation

```go
// ✅ GOOD - Explains the constant's purpose
// DefaultTimeout is the default timeout for API requests.
const DefaultTimeout = 60 * time.Second

// ❌ BAD - No explanation
const DefaultTimeout = 60 * time.Second
```

## Package Documentation

Each package should have a doc comment that explains its purpose:

```go
// Package app provides core client functionality for the Synthetic.new API.
//
// The Client type is the main entry point, with methods for chat completions,
// embeddings, model listing, and vision analysis.
package app
```

## Documentation Checklist

Before submitting code, verify:

- [ ] All exported types, functions, constants, and variables have doc comments
- [ ] Summary lines start with a capital and end with a period
- [ ] Parameters are documented with "Param name:" format
- [ ] Return values are documented with "Returns:" format
- [ ] Complex behaviors have extended descriptions
- [ ] Examples are included for non-obvious APIs
- [ ] Package comments explain the package's purpose

## Further Reading

- [Effective Go - Commentary](https://go.dev/doc/effective_go#commentary)
- [Standard Go Doc Comments](https://tip.golang.org/doc/comment)
