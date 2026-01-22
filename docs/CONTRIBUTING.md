# Contributing to Syn

Thank you for your interest in contributing to Syn! This guide will help you get started.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Code Style](#code-style)
- [Submitting Issues](#submitting-issues)
- [Submitting Pull Requests](#submitting-pull-requests)
- [Testing](#testing)

## Code of Conduct

- Be respectful and inclusive
- Provide constructive feedback
- Focus on what is best for the community
- Show empathy towards other community members

## Getting Started

### Prerequisites

- Go 1.25.4 or later
- Git
- A Synthetic.new API key (for testing)

### Project Structure

```
syn/
├── cmd/          # CLI commands
├── internal/     # Internal application code
├── docs/         # Documentation
├── go.mod        # Go module definition
└── main.go       # Application entry point
```

## Development Setup

1. **Fork and clone the repository:**

   ```bash
   git clone https://github.com/YOUR_USERNAME/syn.git
   cd syn
   ```

2. **Install dependencies:**

   ```bash
   go mod download
   ```

3. **Build the project:**

   ```bash
   go build -o syn
   ```

4. **Run tests:**

   ```bash
   go test ./...
   ```

5. **Verify installation:**

   ```bash
   ./syn --help
   ```

## Code Style

### Go Conventions

- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Use `gofmt` for formatting (run `gofmt -w .`)
- Run `golint` and `go vet` before committing

### Naming Conventions

- **Packages**: lowercase, single word when possible
- **Exports**: PascalCase (e.g., `NewClient`)
- **Private**: camelCase (e.g., `makeRequest`)
- **Constants**: UPPER_SNAKE_CASE or PascalCase for exported
- **Interfaces**: `-er` suffix (e.g., `Reader`, `Writer`)

### Code Organization

- Keep functions focused and under 50 lines when possible
- Use table-driven tests for multiple test cases
- Document exported functions, types, and constants
- Handle errors explicitly; don't ignore them

### Example Format

```go
// Package app provides core client functionality for Synthetic.new API.
package app

import (
    "context"
    "fmt"
)

// NewClient creates and returns a new API client.
// It validates the API key and sets up default configuration.
func NewClient(apiKey string) (*Client, error) {
    if apiKey == "" {
        return nil, fmt.Errorf("API key is required")
    }
    return &Client{
        apiKey: apiKey,
        baseURL: "https://api.synthetic.new",
    }, nil
}
```

### Commit Messages

Follow conventional commit format:

```
type(scope): description

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Test additions or changes
- `refactor`: Code refactoring
- `chore`: Maintenance tasks

Example:
```
feat(chat): add streaming response support

Implement SSE streaming for chat completions to enable
real-time response generation.

Closes #123
```

## Submitting Issues

### Before Creating an Issue

1. Search existing issues to avoid duplicates
2. Check if the issue is already fixed in the latest version
3. Gather relevant information (error messages, logs, environment)

### Issue Template

```markdown
**Description**
A clear and concise description of the problem or feature request.

**Steps to Reproduce** (for bugs)
1. Run command: `syn chat "test"`
2. Observe error: ...

**Expected Behavior**
What should happen.

**Actual Behavior**
What actually happens.

**Environment**
- Go version: 1.25.4
- OS: macOS 15.2
- Syn version: v0.1.0

**Additional Context**
Logs, screenshots, or other relevant information.
```

## Submitting Pull Requests

### Workflow

1. **Fork the repository** and create your branch from `main`
2. **Make your changes** following code style guidelines
3. **Write/update tests** for your changes
4. **Ensure all tests pass:** `go test ./...`
5. **Update documentation** if needed
6. **Commit your changes** with clear messages
7. **Push to your fork** and submit a pull request

### Branch Naming

Use descriptive branch names:

- `feat/add-embeddings-command`
- `fix/memory-leak-in-client`
- `docs/update-api-reference`
- `refactor/cleanup-config-handling`

### Pull Request Template

```markdown
**Description**
Brief description of changes and motivation.

**Type of Change**
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

**Testing**
- [ ] Tests added/updated
- [ ] All tests pass: `go test ./...`
- [ ] Manual testing completed

**Checklist**
- [ ] Code follows style guidelines
- [ ] Documentation updated
- [ ] Commit messages follow convention
- [ ] No merge conflicts

**Related Issues**
Fixes #123
Related to #456
```

### Review Process

- Maintainers will review your PR within a few days
- Address feedback requests promptly
- Keep the PR focused and minimal
- Squash commits if requested before merging

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/app

# Run with verbose output
go test -v ./...
```

### Writing Tests

- Write tests for all new functionality
- Aim for >80% code coverage
- Use table-driven tests for multiple cases
- Test edge cases: empty inputs, nil values, errors

### Example Test

```go
func TestNewClient(t *testing.T) {
    tests := []struct {
        name    string
        apiKey  string
        wantErr bool
    }{
        {"valid key", "test-key", false},
        {"empty key", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            client, err := NewClient(tt.apiKey)
            if (err != nil) != tt.wantErr {
                t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !tt.wantErr && client == nil {
                t.Error("NewClient() returned nil client")
            }
        })
    }
}
```

## Questions?

- Open an issue for bugs or feature requests
- Check existing documentation in `docs/`
- Review example code in `docs/examples/`

Thank you for contributing to Syn!
