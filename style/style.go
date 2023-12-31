package style

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	subtle = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	White  = lipgloss.AdaptiveColor{Dark: "#ffffff", Light: "#ffffff"}
	Yellow = lipgloss.AdaptiveColor{Dark: "#daa520", Light: "#daa520"}
	Red    = lipgloss.AdaptiveColor{Dark: "#ff0000", Light: "#ff0000"}
	Pink   = lipgloss.AdaptiveColor{Light: "#F26D94", Dark: "#F25D94"}
	Gray   = lipgloss.AdaptiveColor{Light: "#3c3836", Dark: "#3c3836"}

	Help = lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(50).
		Foreground(subtle)
	DialogTitle = lipgloss.NewStyle().Width(50).Align(lipgloss.Center)
	DialogBox   = lipgloss.NewStyle().
			Padding(1, 0).
			BorderTop(true).
			BorderLeft(true).
			BorderRight(true).
			BorderBottom(true)

	NormalTitle = lipgloss.NewStyle().
			Foreground(Gray).
			Padding(0, 0, 0, 2)
	SelectedTitle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			Foreground(Pink).
			Padding(0, 0, 0, 1)

	Underline     = lipgloss.NewStyle().Underline(false)
	StrikeThrough = lipgloss.NewStyle().Strikethrough(true)
	Selected      = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, true)
)
