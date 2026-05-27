package theme

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	colorPrimary = lipgloss.AdaptiveColor{Light: "#0066CC", Dark: "#66B2FF"}
	colorMuted   = lipgloss.AdaptiveColor{Light: "#999999", Dark: "#666666"}
	colorAccent  = lipgloss.AdaptiveColor{Light: "#00AA00", Dark: "#00FF00"}

	// Styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary)

	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			PaddingTop(1)

	SelectedStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)

	MutedStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	StatusBarStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Background(lipgloss.AdaptiveColor{Light: "#E0E0E0", Dark: "#1A1A2E"}).
			Foreground(colorMuted)

	BorderStyle = lipgloss.NormalBorder()

	TableBorder = lipgloss.NewStyle().
			BorderStyle(BorderStyle).
			BorderForeground(colorMuted).
			Padding(0, 1)

	ActionStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)

	SpinnerStyle = lipgloss.NewStyle().Foreground(colorPrimary)
)
