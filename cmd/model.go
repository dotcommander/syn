package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var modelCmd = &cobra.Command{
	Use:   "model",
	Short: "Model management",
	Long:  `List and manage Synthetic.new models.`,
}

var modelListCmd = &cobra.Command{
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

		fmt.Println()
		fmt.Println(theme.Section.Render(fmt.Sprintf("Available Models (%d)", len(models))))
		fmt.Println(theme.Divider.Render(strings.Repeat("-", 50)))
		fmt.Println()

		for _, m := range models {
			fmt.Printf("  %s\n", theme.Command.Render(m.ID))
			if m.OwnedBy != "" {
				fmt.Printf("    %s %s\n",
					theme.Dim.Render("Owner:"),
					theme.Description.Render(m.OwnedBy))
			}
			if m.Created > 0 {
				created := time.Unix(m.Created, 0)
				fmt.Printf("    %s %s\n",
					theme.Dim.Render("Created:"),
					theme.Description.Render(created.Format("2006-01-02")))
			}
			fmt.Println()
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(modelCmd)
	modelCmd.AddCommand(modelListCmd)
}
