# config Package

Configuration management and default values for the syn application.

## Functions

### SetDefaults
```go
func SetDefaults()
```
Configures sensible defaults for all application settings. Called automatically at application startup.

## Default Values

### API Defaults

| Setting | Default Value |
|---------|---------------|
| `api.key` | syn_2afa3d6ae1d48878694a13cbbe35d76c |
| `api.base_url` | https://api.synthetic.new/openai/v1 |
| `api.anthropic_base_url` | https://api.synthetic.new/anthropic/v1 |
| `api.model` | hf:deepseek-ai/DeepSeek-V3.2 |
| `api.embedding_model` | hf:nomic-ai/nomic-embed-text-v1.5 |

### Retry Configuration

| Setting | Default Value |
|---------|---------------|
| `api.retry.max_attempts` | 3 |
| `api.retry.initial_backoff` | 1s |
| `api.retry.max_backoff` | 30s |

### Chat Defaults

| Setting | Default Value |
|---------|---------------|
| `chat.temperature` | 0.6 |
| `chat.max_tokens` | 8192 |
| `chat.top_p` | 0.9 |

## Usage

The config package uses Viper for configuration management. Defaults are set first, then overridden by:
1. Config file (`~/.config/syn/config.yaml`)
2. Environment variables (`SYN_*`)
3. Command-line flags

```go
import "github.com/vampire/syn/internal/config"

func main() {
    config.SetDefaults()
    // Application code...
}
```
