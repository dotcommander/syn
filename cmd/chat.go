package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/dotcommander/syn/internal/app"
)

var chatCmd = &cobra.Command{ //nolint:gochecknoglobals // cobra command registration
	Use:   "chat",
	Short: "Start interactive chat session",
	Long: `Interactive REPL with conversation context.

Commands:
  /clear  - Clear conversation history
  /model  - Show current model
  /exit   - Exit chat session
  /help   - Show help`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInteractiveChat()
	},
}

func init() { //nolint:gochecknoinits // cobra command registration
	rootCmd.AddCommand(chatCmd)
}

// animateThinking displays an animated spinner while waiting for API response.
func animateThinking(w io.Writer, stop *atomic.Bool) {
	if w == nil {
		w = os.Stdout
	}
	spinnerStyle := theme.SpinnerStyle()
	i := 0
	for !stop.Load() {
		fmt.Fprintf(w, "\r%s %s", spinnerStyle.Render(SpinnerFrames[i%len(SpinnerFrames)]), theme.Dim.Render("Thinking..."))
		time.Sleep(80 * time.Millisecond)
		i++
	}
	fmt.Fprint(w, "\r\033[K") // Clear line
}

// inputResult holds the result of reading a line from stdin.
type inputResult struct {
	text string
	err  error
}

func runInteractiveChat() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	client := newClient()
	baseOpts := app.DefaultChatOptions()
	baseOpts.FilePath = viper.GetString("file")

	var conversationContext []app.Message
	maxContextMessages := 20

	printWelcomeBanner()

	scanner := bufio.NewScanner(os.Stdin)
	inputCh := make(chan inputResult, 1)
	go func() {
		for scanner.Scan() {
			inputCh <- inputResult{text: scanner.Text()}
		}
		inputCh <- inputResult{err: scanner.Err()}
	}()

	for {
		fmt.Print(theme.UserPrompt.Render("you> "))

		input, done := waitForInput(ctx, inputCh, scanner)
		if done {
			return nil
		}
		if input == "" {
			continue
		}

		if strings.HasPrefix(input, "/") {
			if handleChatCommand(input, &conversationContext) {
				continue
			}
		}

		opts := buildChatOpts(baseOpts, conversationContext)
		response, err := sendWithSpinner(ctx, client, input, opts)
		if err != nil {
			fmt.Println(theme.ErrorText.Render("Error: ") + theme.Dim.Render(err.Error()))
			fmt.Println()
			continue
		}

		conversationContext = appendExchange(conversationContext, input, response, maxContextMessages)

		fmt.Println()
		fmt.Printf("%s %s\n", theme.AssistantPrompt.Render("syn>"), response)
		fmt.Println()
	}
}

// waitForInput blocks until user input or context cancellation.
// Returns the trimmed input and whether the REPL should exit.
func waitForInput(ctx context.Context, inputCh <-chan inputResult, scanner *bufio.Scanner) (string, bool) {
	select {
	case <-ctx.Done():
		fmt.Println()
		fmt.Println(theme.Dim.Render("Goodbye!"))
		fmt.Println()
		return "", true
	case result := <-inputCh:
		if result.err != nil || (result.text == "" && scanner.Err() != nil) {
			fmt.Println()
			return "", true
		}
		return strings.TrimSpace(result.text), false
	}
}

func buildChatOpts(baseOpts app.ChatOptions, ctx []app.Message) app.ChatOptions {
	opts := baseOpts
	opts.Context = ctx
	if len(ctx) > 0 {
		opts.FilePath = ""
	}
	return opts
}

func sendWithSpinner(ctx context.Context, client *app.Client, input string, opts app.ChatOptions) (string, error) {
	var spinnerStop atomic.Bool
	go animateThinking(nil, &spinnerStop)

	response, _, err := client.Chat(ctx, input, opts)
	spinnerStop.Store(true)
	time.Sleep(100 * time.Millisecond) // Let spinner clear

	return response, err
}

func appendExchange(ctx []app.Message, input, response string, maxMessages int) []app.Message {
	ctx = append(ctx,
		app.Message{Role: "user", Content: input},
		app.Message{Role: "assistant", Content: response},
	)
	if len(ctx) > maxMessages {
		ctx = ctx[2:]
	}
	return ctx
}

func printWelcomeBanner() {
	fmt.Println()
	fmt.Println(theme.Title.Render(" SYN ") + " " + theme.Description.Render("Chat Session"))
	fmt.Println()
	fmt.Println(theme.Info.Render("  Model: ") + theme.Dim.Render(viper.GetString("api.model")))
	fmt.Println()
	fmt.Println(theme.HelpText.Render("  Commands: /help, /clear, /model, /exit"))
	fmt.Println(theme.Divider.Render(strings.Repeat("-", 50)))
	fmt.Println()
}

// handleChatCommand processes chat commands. Returns true if command was handled.
func handleChatCommand(input string, context *[]app.Message) bool {
	switch strings.ToLower(input) {
	case "/clear":
		*context = nil
		fmt.Print("\033[2J\033[H") // Clear screen
		printWelcomeBanner()
		return true

	case "/model":
		fmt.Println()
		fmt.Printf("  %s %s\n",
			theme.Info.Render("Current model:"),
			theme.Description.Render(viper.GetString("api.model")))
		fmt.Println()
		return true

	case "/exit", "/quit":
		fmt.Println()
		fmt.Println(theme.Dim.Render("Goodbye!"))
		fmt.Println()
		os.Exit(0)
		return true

	case "/help", "/?":
		printChatHelp()
		return true

	case "/context":
		printContextStyled(*context)
		return true

	default:
		if strings.HasPrefix(input, "/") {
			fmt.Println()
			fmt.Printf("  %s %s\n",
				theme.ErrorText.Render("Unknown command:"),
				theme.Dim.Render(input))
			fmt.Println(theme.HelpText.Render("  Type /help for available commands"))
			fmt.Println()
			return true
		}
		return false
	}
}

func printChatHelp() {
	fmt.Println()
	fmt.Println(theme.Section.Render("Chat Commands"))
	fmt.Println(theme.Divider.Render(strings.Repeat("-", 40)))

	commands := []struct {
		cmd  string
		desc string
	}{
		{"/help", "Show this help"},
		{"/clear", "Clear conversation and screen"},
		{"/model", "Show current model"},
		{"/context", "Show conversation context"},
		{"/exit", "Exit chat session"},
	}

	for _, c := range commands {
		fmt.Printf("  %s  %s\n",
			theme.Info.Render(fmt.Sprintf("%-12s", c.cmd)),
			theme.Dim.Render(c.desc))
	}
	fmt.Println()
}

func printContextStyled(ctx []app.Message) {
	fmt.Println()
	if len(ctx) == 0 {
		fmt.Println(theme.Dim.Render("  No context yet."))
		fmt.Println()
		return
	}

	fmt.Println(theme.Section.Render(fmt.Sprintf("Conversation Context (%d messages)", len(ctx))))
	fmt.Println(theme.Divider.Render(strings.Repeat("-", 40)))

	for _, msg := range ctx {
		var styledRole string
		if msg.Role == "user" {
			styledRole = theme.UserPrompt.Render("[You]")
		} else {
			styledRole = theme.AssistantPrompt.Render("[Syn]")
		}
		fmt.Printf("  %s %s\n",
			styledRole,
			theme.Dim.Render(truncateString(msg.Content, 50)))
	}
	fmt.Println()
}

func truncateString(s string, maxLen int) string {
	// Remove newlines for display
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
