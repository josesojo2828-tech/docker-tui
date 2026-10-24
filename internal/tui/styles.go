package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Theme colors
	primaryColor   = lipgloss.Color("#3b82f6") // Blue
	secondaryColor = lipgloss.Color("#8b5cf6") // Purple
	successColor   = lipgloss.Color("#10b981") // Emerald Green
	warningColor   = lipgloss.Color("#f59e0b") // Amber
	dangerColor    = lipgloss.Color("#ef4444") // Red
	mutedColor     = lipgloss.Color("#64748b") // Slate
	darkBgColor    = lipgloss.Color("#0f172a") // Slate 900
	highlightColor = lipgloss.Color("#38bdf8") // Sky Blue

	// Styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ffffff")).
			Background(primaryColor).
			Padding(0, 1)

	headerBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#334155")).
			Padding(0, 1).
			MarginBottom(1)

	tabActiveStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ffffff")).
			Background(primaryColor).
			Padding(0, 2)

	tabInactiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#94a3b8")).
				Padding(0, 2)

	statBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#334155")).
			Padding(0, 1).
			Align(lipgloss.Center)

	statLabelStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Bold(false)

	statValueStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(highlightColor)

	statusMsgStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	statusErrStyle = lipgloss.NewStyle().
			Foreground(dangerColor).
			Bold(true)

	helpBarStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			MarginTop(1)

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(highlightColor).
			Bold(true)
)
