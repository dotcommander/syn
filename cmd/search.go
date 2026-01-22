package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

		return runSearch(query)
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}

func runSearch(query string) error {
	client := newClient()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := client.Search(ctx, query)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if viper.GetBool("json") {
		data, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	} else {
		fmt.Println()
		fmt.Println(theme.Section.Render(fmt.Sprintf("Search Results (%d)", len(resp.Results))))
		fmt.Println(theme.Divider.Render(strings.Repeat("-", 60)))
		fmt.Println()

		if len(resp.Results) == 0 {
			fmt.Println(theme.Dim.Render("  No results found."))
			fmt.Println()
			return nil
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

	return nil
}
