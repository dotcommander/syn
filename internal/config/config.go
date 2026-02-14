package config

import (
	"time"

	"github.com/spf13/viper"
)

// SetDefaults configures sensible defaults for the application.
// Note: API key must be provided via SYN_API_KEY env var or config file.
func SetDefaults() {
	// API defaults (key intentionally omitted - must be configured by user)
	viper.SetDefault("api.base_url", "https://api.synthetic.new/openai/v1")
	viper.SetDefault("api.anthropic_base_url", "https://api.synthetic.new/anthropic/v1")
	viper.SetDefault("api.model", "hf:deepseek-ai/DeepSeek-V3.2")
	viper.SetDefault("api.embedding_model", "hf:nomic-ai/nomic-embed-text-v1.5")

	// Retry configuration
	viper.SetDefault("api.retry.max_attempts", 3)
	viper.SetDefault("api.retry.initial_backoff", 1*time.Second)
	viper.SetDefault("api.retry.max_backoff", 30*time.Second)

	// Chat defaults
	viper.SetDefault("chat.temperature", 0.6)
	viper.SetDefault("chat.max_tokens", 8192)
	viper.SetDefault("chat.top_p", 0.9)
}
