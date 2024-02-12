package style

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// todo maybe make these more semantic fg/bg etc
	White  = lipgloss.AdaptiveColor{Dark: "#ffffff", Light: "#ffffff"}
	Yellow = lipgloss.AdaptiveColor{Dark: "#daa520", Light: "#daa520"}
	Red    = lipgloss.AdaptiveColor{Dark: "#ff0000", Light: "#ff0000"}
	Pink   = lipgloss.AdaptiveColor{Light: "#F26D94", Dark: "#F25D94"}
	Gray   = lipgloss.AdaptiveColor{Light: "#8E6F8C", Dark: "#C0A7BF"}

	DialogTitle = lipgloss.NewStyle().Width(50).Align(lipgloss.Center)
	DialogBox   = lipgloss.NewStyle().
			Padding(1, 0).
			BorderTop(true).
			BorderLeft(true).
			BorderRight(true).
			BorderBottom(true)
	DialogBoxStyle = lipgloss.NewStyle().
			Width(70).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Pink).
			BorderTop(true).
			BorderLeft(true).
			BorderRight(true).
			BorderBottom(true)
	NormalTitle = lipgloss.NewStyle().
			Foreground(Gray)
	SelectedTitle = lipgloss.NewStyle().
			Foreground(Pink)

	Underline     = lipgloss.NewStyle().Underline(false)
	StrikeThrough = lipgloss.NewStyle().Strikethrough(true)
	Selected      = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, true)
)
