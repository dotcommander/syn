package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/dotcommander/syn/internal/app"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search the web",
	Long: `Search the web using Synthetic's /v2/search endpoint.

Note: This API is under development and may have breaking changes.
Zero-data-retention: queries are not stored.

Examples:
  syn search "golang error handling"
  syn search --json "react hooks"
  syn search -i "claude docs"    # Interactive mode
  echo "python async" | syn search`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var query string

		// Check for stdin
		if hasStdinData() {
			data, err := readStdin()
			if err != nil {
				return fmt.Errorf("failed to read stdin: %w", err)
			}
			query = data
		}

		// Args override stdin
		if len(args) > 0 {
			query = strings.Join(args, " ")
		}

		if query == "" {
			return fmt.Errorf("no search query provided (use args or stdin)")
		}

		interactive, _ := cmd.Flags().GetBool("interactive")
		return runSearch(query, interactive)
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().BoolP("interactive", "i", false, "Enable interactive result selection")
}

func runSearch(query string, interactive bool) error {
	client := newClient()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := client.Search(ctx, query)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	// JSON mode - skip interactive
	if viper.GetBool("json") {
		data, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	// Print results
	printSearchResults(resp)

	// Interactive mode
	if interactive && len(resp.Results) > 0 {
		return interactiveSelection(resp.Results)
	}

	return nil
}

func printSearchResults(resp *app.SearchResponse) {
	fmt.Println()
	fmt.Println(theme.Section.Render(fmt.Sprintf("Search Results (%d)", len(resp.Results))))
	fmt.Println(theme.Divider.Render(strings.Repeat("-", 60)))
	fmt.Println()

	if len(resp.Results) == 0 {
		fmt.Println(theme.Dim.Render("  No results found."))
		fmt.Println()
		return
	}

	for i, result := range resp.Results {
		fmt.Printf("  %s %s\n",
			theme.Command.Render(fmt.Sprintf("%d.", i+1)),
			theme.Info.Render(result.Title))

		fmt.Printf("     %s\n", theme.Dim.Render(result.URL))

		if result.Snippet != "" {
			snippet := truncateString(result.Snippet, 100)
			fmt.Printf("     %s\n", theme.Description.Render(snippet))
		}

		if result.Published != "" {
			fmt.Printf("     %s %s\n",
				theme.Dim.Render("Published:"),
				theme.Dim.Render(result.Published))
		}

		fmt.Println()
	}
}

func interactiveSelection(results []app.SearchResult) error {
	reader := bufio.NewReader(os.Stdin)

	for {
		// Prompt styled with theme
		fmt.Printf("%s ", theme.UserPrompt.Render(
			fmt.Sprintf("Select result (1-%d, q to quit):", len(results))))

		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		input = strings.TrimSpace(input)

		// Quit command
		if input == "q" || input == "quit" {
			fmt.Println(theme.Dim.Render("Exiting."))
			return nil
		}

		// Parse selection
		selection, err := strconv.Atoi(input)
		if err != nil || selection < 1 || selection > len(results) {
			fmt.Println(theme.ErrorText.Render(
				fmt.Sprintf("Invalid selection. Enter 1-%d or 'q'.", len(results))))
			continue
		}

		// Open browser
		result := results[selection-1]
		fmt.Println(theme.SuccessText.Render(fmt.Sprintf("Opening: %s", result.URL)))

		if err := openBrowser(result.URL); err != nil {
			return fmt.Errorf("failed to open browser: %w", err)
		}

		// Continue prompting for more selections
		fmt.Println()
	}
}

// openBrowser opens a URL in the default browser (cross-platform).
func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin": // macOS
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}
