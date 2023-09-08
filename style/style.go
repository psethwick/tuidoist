package style

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// tuiStyle = lipgloss.NewStyle().
	// 		Border(lipgloss.RoundedBorder()).
	// 		BorderForeground(lipgloss.Color("#874BFD")).
	// 		Margin(1).
	// 		BorderTop(true).
	// 		BorderLeft(true).
	// 		BorderRight(true).
	// 		BorderBottom(true)

	// dialog
	subtle = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	Help   = lipgloss.NewStyle().
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
	buttonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFF7DB")).
			Background(lipgloss.Color("#888B7E")).
			Padding(0, 3).
			MarginTop(1)

	activeButtonStyle = buttonStyle.Copy().
				Foreground(lipgloss.Color("#FFF7DB")).
				Background(lipgloss.Color("#F25D94")).
				MarginRight(2).
				Underline(true)

		// List
	listStyle = lipgloss.NewStyle().
			Align(lipgloss.Left).
			BorderStyle(lipgloss.HiddenBorder())
	listTitleStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("230")).
			Padding(0, 1)

		// from default delegate, keep here for now?
	NormalTitle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"}).
			Padding(0, 0, 0, 2)
	normalDesc = NormalTitle.Copy().
			Foreground(lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#777777"})
	SelectedTitle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"}).
			Foreground(lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"}).
			Padding(0, 0, 0, 1)

	Underline         = lipgloss.NewStyle().Underline(false)
	listTitleBarStyle = lipgloss.NewStyle().Padding(0, 0, 1, 2)
	StrikeThrough     = lipgloss.NewStyle().Strikethrough(true)
)
