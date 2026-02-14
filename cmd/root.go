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

	"github.com/dotcommander/syn/internal/app"
)

var ( //nolint:gochecknoglobals // cobra flag bindings require package-level vars
	cfgFile    string
	verbose    bool
	filePath   string
	jsonOutput bool
	modelFlag  string
)

var rootCmd = &cobra.Command{ //nolint:gochecknoglobals // cobra root command
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
		fmt.Fprintf(os.Stderr, "\n%s %s\n\n",
			theme.ErrorText.Render("Error:"),
			theme.Description.Render(err.Error()))
		os.Exit(1)
	}
}

func init() { //nolint:gochecknoinits // cobra command registration
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true
	rootCmd.SetHelpFunc(styledHelp)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default $HOME/.config/syn/config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVarP(&filePath, "file", "f", "", "include file contents in prompt")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output in JSON format")
	rootCmd.PersistentFlags().StringVarP(&modelFlag, "model", "m", "", "model to use (aliases: kimi, qwen, coder, glm, gpt, r1, minimax, llama)")

	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	_ = viper.BindPFlag("file", rootCmd.PersistentFlags().Lookup("file"))
	_ = viper.BindPFlag("json", rootCmd.PersistentFlags().Lookup("json"))
	_ = viper.BindPFlag("model", rootCmd.PersistentFlags().Lookup("model"))
}

func styledHelp(cmd *cobra.Command, args []string) {
	// For subcommands, print their description and usage
	if cmd.Name() != "syn" {
		fmt.Println()
		fmt.Println(theme.Title.Render(" "+strings.ToUpper(cmd.Name())+" ") + " " + theme.Description.Render(cmd.Short))
		fmt.Println()
		if cmd.Long != "" {
			fmt.Println(theme.Description.Render(cmd.Long))
			fmt.Println()
		}
		fmt.Println(theme.Section.Render("Usage"))
		fmt.Println(theme.Divider.Render(strings.Repeat("-", 40)))
		fmt.Printf("  %s\n\n", theme.Example.Render(cmd.UseLine()))
		return
	}

	fmt.Println()
	fmt.Println(theme.Title.Render(" SYN ") + " " + theme.Description.Render("Chat with Synthetic.new models"))
	fmt.Println()

	// Examples section
	fmt.Println(theme.Section.Render("Examples"))
	fmt.Println(theme.Divider.Render(strings.Repeat("-", 50)))
	examples := []string{
		`syn "Explain quantum computing"`,
		`syn -f main.go "Review this code"`,
		`syn search "golang context"`,
		`syn eval --limit 1`,
		`syn embed "Hello world"`,
	}
	for _, ex := range examples {
		fmt.Printf("  %s\n", theme.Example.Render(ex))
	}
	fmt.Println()

	// Commands section
	fmt.Println(theme.Section.Render("Commands"))
	fmt.Println(theme.Divider.Render(strings.Repeat("-", 50)))
	commands := [][]string{
		{"chat", "Interactive chat session (REPL)"},
		{"search", "Search the web"},
		{"eval", "Evaluate key-insight extraction"},
		{"vision", "Analyze images with AI"},
		{"embed", "Generate text embeddings"},
		{"model", "Model management"},
	}
	for _, c := range commands {
		fmt.Printf("  %s  %s\n",
			theme.Command.Render(fmt.Sprintf("%-10s", c[0])),
			theme.Description.Render(c[1]))
	}
	fmt.Println()

	// Flags section
	fmt.Println(theme.Section.Render("Flags"))
	fmt.Println(theme.Divider.Render(strings.Repeat("-", 50)))
	flags := [][]string{
		{"-m, --model <name>", "Model (kimi, qwen, coder, r1, glm, gpt, ...)"},
		{"-f, --file <path>", "Include file contents in prompt"},
		{"--json", "Output as JSON"},
		{"-v, --verbose", "Show debug info"},
		{"-h, --help", "Show this help"},
	}
	for _, f := range flags {
		fmt.Printf("  %s  %s\n",
			theme.Flag.Render(fmt.Sprintf("%-18s", f[0])),
			theme.Description.Render(f[1]))
	}
	fmt.Println()

	fmt.Println(theme.Description.Render("Use \"syn <command> --help\" for command details"))
	fmt.Println()
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

	// Also accept SYNTHETIC_API_KEY
	_ = viper.BindEnv("api.key", "SYN_API_KEY", "SYNTHETIC_API_KEY")

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
	if m := viper.GetString("model"); m != "" {
		opts.Model = m
	}

	if viper.GetBool("verbose") {
		fmt.Fprintf(os.Stderr, "Prompt: %s\n", prompt)
		if opts.FilePath != "" {
			fmt.Fprintf(os.Stderr, "File: %s\n", opts.FilePath)
		}
		if opts.Model != "" {
			fmt.Fprintf(os.Stderr, "Model: %s\n", app.ResolveModel(opts.Model))
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	response, _, err := client.Chat(ctx, prompt, opts)
	if err != nil {
		return fmt.Errorf("failed to get response: %w", err)
	}

	if viper.GetBool("json") {
		output := map[string]any{
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
