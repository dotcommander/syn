package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/dotcommander/syn/internal/app"
)

var visionCmd = &cobra.Command{ //nolint:gochecknoglobals // cobra command registration
	Use:   "vision [prompt]",
	Short: "Analyze images with AI",
	Long: `Analyze images using a vision-capable model.

Examples:
  syn vision -f photo.jpg "What's in this image?"
  syn vision -f https://example.com/image.png "Describe this"
  syn vision -f screenshot.png  # Uses default prompt

Supported formats: JPEG, PNG, GIF, WebP
Accepts URLs or local file paths via -f flag.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		imageSource := viper.GetString("file")
		if imageSource == "" {
			return fmt.Errorf("image required: use -f <image>")
		}

		prompt := "What do you see in this image? Please describe it in detail."
		if len(args) > 0 {
			prompt = strings.Join(args, " ")
		}

		return runVision(imageSource, prompt)
	},
}

func init() { //nolint:gochecknoinits // cobra command registration
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
		fmt.Fprintf(os.Stderr, "Model: %s\n", app.ResolveModel("kimi"))
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
