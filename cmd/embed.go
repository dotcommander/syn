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

var embedCmd = &cobra.Command{ //nolint:gochecknoglobals // cobra command registration
	Use:   "embed [text...]",
	Short: "Generate text embeddings",
	Long: `Generate vector embeddings for text using nomic-embed-text-v1.5.

Examples:
  syn embed "Hello world"
  syn embed "Text 1" "Text 2" "Text 3"
  echo "Text" | syn embed
  syn embed --json "Hello world"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var texts []string

		// Check for stdin
		if hasStdinData() {
			data, err := readStdin()
			if err != nil {
				return fmt.Errorf("failed to read stdin: %w", err)
			}
			texts = append(texts, data)
		}

		// Add args
		texts = append(texts, args...)

		if len(texts) == 0 {
			return fmt.Errorf("no text provided (use args or stdin)")
		}

		return runEmbed(texts)
	},
}

func init() { //nolint:gochecknoinits // cobra command registration
	rootCmd.AddCommand(embedCmd)
}

func runEmbed(texts []string) error {
	client := newClient()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	model := viper.GetString("api.embedding_model")
	resp, err := client.Embed(ctx, texts, model)
	if err != nil {
		return fmt.Errorf("embedding failed: %w", err)
	}

	if viper.GetBool("json") {
		// Output raw JSON
		data, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	} else {
		// Human-readable output
		fmt.Println()
		fmt.Println(theme.Section.Render(fmt.Sprintf("Embeddings (%d)", len(resp.Data))))
		fmt.Println(theme.Divider.Render(strings.Repeat("-", 50)))
		fmt.Println()

		fmt.Printf("  %s %s\n",
			theme.Info.Render("Model:"),
			theme.Description.Render(resp.Model))
		fmt.Printf("  %s %d\n",
			theme.Info.Render("Total tokens:"),
			resp.Usage.TotalTokens)
		fmt.Println()

		for i, emb := range resp.Data {
			fmt.Printf("  %s %s\n",
				theme.Command.Render(fmt.Sprintf("Text %d:", i+1)),
				theme.Dim.Render(truncateString(texts[i], 60)))
			fmt.Printf("    %s %d\n",
				theme.Dim.Render("Dimensions:"),
				len(emb.Embedding))

			// Show first 5 values
			if len(emb.Embedding) > 0 {
				preview := emb.Embedding[:min(5, len(emb.Embedding))]
				fmt.Printf("    %s %v\n",
					theme.Dim.Render("First 5:"),
					preview)
			}
			fmt.Println()
		}
	}

	return nil
}

