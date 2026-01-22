package cmd

import "github.com/charmbracelet/lipgloss"

// Theme holds all lipgloss styles for consistent UI across commands.
type Theme struct {
	// Colors
	Primary lipgloss.Color
	Success lipgloss.Color
	Error   lipgloss.Color
	Subtle  lipgloss.Color
	Accent  lipgloss.Color
	White   lipgloss.Color
	Gold    lipgloss.Color
	Dark    lipgloss.Color

	// Common styles
	Title       lipgloss.Style
	Section     lipgloss.Style
	Divider     lipgloss.Style
	Dim         lipgloss.Style
	Info        lipgloss.Style
	ErrorText   lipgloss.Style
	HelpText    lipgloss.Style
	Description lipgloss.Style

	// Chat-specific styles
	UserPrompt      lipgloss.Style
	AssistantPrompt lipgloss.Style

	// Command/flag styles (for help)
	Command lipgloss.Style
	Flag    lipgloss.Style
	Example lipgloss.Style
}

// DefaultTheme returns the default theme with Synthetic colors.
func DefaultTheme() *Theme {
	t := &Theme{
		// Colors - using cyan/teal for Synthetic branding
		Primary: lipgloss.Color("#00CED1"), // Dark Cyan
		Success: lipgloss.Color("#00FA9A"), // Medium Spring Green
		Error:   lipgloss.Color("#FF6B6B"),
		Subtle:  lipgloss.Color("#626262"),
		Accent:  lipgloss.Color("#00BFFF"), // Deep Sky Blue
		White:   lipgloss.Color("#FAFAFA"),
		Gold:    lipgloss.Color("#FFD700"),
		Dark:    lipgloss.Color("#444444"),
	}

	// Build styles from colors
	t.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(t.White).
		Background(t.Primary).
		Padding(0, 1)

	t.Section = lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Primary)

	t.Divider = lipgloss.NewStyle().
		Foreground(t.Dark)

	t.Dim = lipgloss.NewStyle().
		Foreground(t.Subtle)

	t.Info = lipgloss.NewStyle().
		Foreground(t.Accent)

	t.ErrorText = lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Error)

	t.HelpText = lipgloss.NewStyle().
		Italic(true).
		Foreground(t.Subtle)

	t.Description = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#AAAAAA"))

	// Chat styles
	t.UserPrompt = lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Success)

	t.AssistantPrompt = lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Primary)

	// Help styles
	t.Command = lipgloss.NewStyle().
		Foreground(t.Success)

	t.Flag = lipgloss.NewStyle().
		Foreground(t.Accent)

	t.Example = lipgloss.NewStyle().
		Italic(true).
		Foreground(t.Gold)

	return t
}

// SpinnerStyle returns a style for the thinking spinner.
func (t *Theme) SpinnerStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.Primary)
}

// theme is the default theme instance used by all commands.
var theme = DefaultTheme()

// SpinnerFrames contains the Braille animation frames for loading spinners.
var SpinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
