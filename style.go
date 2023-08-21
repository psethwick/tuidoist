package main

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// dialog
	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	helpStyle = lipgloss.NewStyle().
			Align(lipgloss.Center).
			Width(50).
			Foreground(subtle)
	dialogTitle    = lipgloss.NewStyle().Width(50).Align(lipgloss.Center)
	dialogBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
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

	underlineStyle     = lipgloss.NewStyle().Underline(false)
	listTitleBarStyle  = lipgloss.NewStyle().Padding(0, 0, 1, 2)
	strikeThroughStyle = lipgloss.NewStyle().Strikethrough(true)
)
