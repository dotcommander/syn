package cmd

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/dotcommander/syn/internal/app"
)

var modelCmd = &cobra.Command{ //nolint:gochecknoglobals // cobra command registration
	Use:   "model",
	Short: "Model management",
	Long:  `List and manage Synthetic.new models.`,
}

var modelListCmd = &cobra.Command{ //nolint:gochecknoglobals // cobra command registration
	Use:   "list",
	Short: "List available models",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newClient()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		models, err := client.ListModels(ctx)
		if err != nil {
			return fmt.Errorf("failed to list models: %w", err)
		}

		// Build reverse alias lookup: full model ID → []aliases
		reverseAliases := buildReverseAliases()

		// Sort models alphabetically by ID
		sort.Slice(models, func(i, j int) bool {
			return models[i].ID < models[j].ID
		})

		fmt.Println()
		fmt.Println(theme.Section.Render(fmt.Sprintf("Models (%d)", len(models))))
		fmt.Println(theme.Divider.Render(strings.Repeat("-", 50)))
		fmt.Println()

		// Find longest model ID for alignment
		maxLen := 0
		for _, m := range models {
			if len(m.ID) > maxLen {
				maxLen = len(m.ID)
			}
		}

		for _, m := range models {
			// Build right-side tags
			var tags []string
			aliases := reverseAliases[m.ID]
			if len(aliases) > 0 {
				tags = append(tags, theme.Flag.Render(strings.Join(aliases, ", ")))
			}
			if visionModels[m.ID] {
				tags = append(tags, theme.Example.Render("vision"))
			}

			pad := maxLen - len(m.ID) + 2
			if len(tags) > 0 {
				fmt.Printf("  %s%s%s\n", theme.Command.Render(m.ID), strings.Repeat(" ", pad), strings.Join(tags, "  "))
			} else {
				fmt.Printf("  %s\n", theme.Dim.Render(m.ID))
			}
		}

		fmt.Println()
		return nil
	},
}

// visionModels lists model IDs known to support image inputs.
var visionModels = map[string]bool{ //nolint:gochecknoglobals // read-only lookup table
	"hf:moonshotai/Kimi-K2.5":       true,
	"hf:nvidia/Kimi-K2.5-NVFP4":     true,
}

// buildReverseAliases creates a map from full model ID to its short aliases.
func buildReverseAliases() map[string][]string {
	reverse := make(map[string][]string)
	for alias, fullID := range app.ModelAliases() {
		reverse[fullID] = append(reverse[fullID], alias)
	}
	// Sort aliases for deterministic output
	for id := range reverse {
		sort.Strings(reverse[id])
	}
	return reverse
}

func init() { //nolint:gochecknoinits // cobra command registration
	rootCmd.AddCommand(modelCmd)
	modelCmd.AddCommand(modelListCmd)
}
