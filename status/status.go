package status

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mistakenelf/teacup/statusbar"
)

type SyncStatus uint

const (
	Synced SyncStatus = iota
	Syncing
	Error
)

type Model struct {
	// SyncStatus SyncStatus
	// Title      string
	// MetaData   string
	bar *statusbar.Model
}

func (sb *Model) Update(msg tea.Msg) tea.Cmd {
	newBar, cmd := sb.bar.Update(msg)
	sb.bar = &newBar
	return cmd
}

func (sb *Model) View() string {
	return sb.bar.View()
}

func (sb *Model) SetSyncStatus(ss SyncStatus) {
	switch ss {
	case Synced:
		sb.bar.FirstColumn = "Up to date"
	case Syncing:
		sb.bar.FirstColumn = "Syncing"
	case Error:
		sb.bar.FirstColumn = "Error"
	}
}

func (sb *Model) SetTitle(t string) {
	sb.bar.SecondColumn = t
}

func New() Model {
	sb := statusbar.New(
		// todo move to style.go
		statusbar.ColorConfig{
			Foreground: lipgloss.AdaptiveColor{Dark: "#ffffff", Light: "#ffffff"},
			Background: lipgloss.AdaptiveColor{Light: "#F25D94", Dark: "#F25D94"},
		},
		statusbar.ColorConfig{
			Foreground: lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#ffffff"},
			Background: lipgloss.AdaptiveColor{Light: "#3c3836", Dark: "#3c3836"},
		},
		statusbar.ColorConfig{
			Foreground: lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#ffffff"},
			Background: lipgloss.AdaptiveColor{Light: "#3c3836", Dark: "#3c3836"},
		},

		// statusbar.ColorConfig{
		// 	Foreground: lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#ffffff"},
		// 	Background: lipgloss.AdaptiveColor{Light: "#A550DF", Dark: "#A550DF"},
		// },
		statusbar.ColorConfig{
			Foreground: lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#ffffff"},
			Background: lipgloss.AdaptiveColor{Light: "#A550DF", Dark: "#A550DF"},
		},
		// statusbar.ColorConfig{
		// 	Foreground: lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#ffffff"},
		// 	Background: lipgloss.AdaptiveColor{Light: "#6124DF", Dark: "#6124DF"},
		// },
	)
	m := Model{
		bar: &sb,
	}
	sb.SetContent("Syncing", "Inbox", "", "0")

	return m
}
