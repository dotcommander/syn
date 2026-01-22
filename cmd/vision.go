package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/vampire/syn/internal/app"
)

var visionCmd = &cobra.Command{
	Use:   "vision <image> [prompt]",
	Short: "Analyze images with AI",
	Long: `Analyze images using Qwen3-VL vision model.

Examples:
  syn vision photo.jpg "What's in this image?"
  syn vision https://example.com/image.png "Describe this"
  syn vision screenshot.png  # Uses default prompt

Supported formats: JPEG, PNG, GIF, WebP
Accepts URLs or local file paths.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		imageSource := args[0]

		prompt := "What do you see in this image? Please describe it in detail."
		if len(args) > 1 {
			prompt = strings.Join(args[1:], " ")
		}

		return runVision(imageSource, prompt)
	},
}

func init() {
	rootCmd.AddCommand(visionCmd)
}

func runVision(imageSource, prompt string) error {
	client := newClient()
	opts := app.DefaultChatOptions()

	if m := viper.GetString("model"); m != "" {
		opts.Model = m
	}

	if viper.GetBool("verbose") {
		fmt.Fprintf(os.Stderr, "Image: %s\n", imageSource)
		fmt.Fprintf(os.Stderr, "Prompt: %s\n", prompt)
		fmt.Fprintf(os.Stderr, "Model: %s\n", app.ResolveModel("qwen"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	response, err := client.Vision(ctx, prompt, imageSource, opts)
	if err != nil {
		return fmt.Errorf("vision failed: %w", err)
	}

	fmt.Println(response)
	return nil
}
