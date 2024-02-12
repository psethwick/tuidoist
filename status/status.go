package status

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mistakenelf/teacup/statusbar"
	"github.com/psethwick/tuidoist/style"
)

type SyncStatus uint

const (
	Synced SyncStatus = iota
	Syncing
	Error
)

type Model struct {
	bar   *statusbar.Model
	sort  string
	count int
}

func (sb *Model) Update(msg tea.Msg) tea.Cmd {
	newBar, cmd := sb.bar.Update(msg)
	sb.bar = &newBar
	return cmd
}

func (sb *Model) View() string {
	return sb.bar.View()
}

// todo eject teacup/statusbar, it is quite simple
func (sb *Model) SetSyncStatus(ss SyncStatus) {
	switch ss {
	case Synced:
		sb.bar.FirstColumn = "✅"
		sb.bar.FirstColumnColors = statusbar.ColorConfig{
			Foreground: style.White,
			Background: style.Pink,
		}
	case Syncing:
		sb.bar.FirstColumn = "🔁"
		sb.bar.FirstColumnColors = statusbar.ColorConfig{
			Foreground: style.White,
			Background: style.Yellow,
		}
	case Error:
		sb.bar.FirstColumn = "❌"
		sb.bar.FirstColumnColors = statusbar.ColorConfig{
			Foreground: style.White,
			Background: style.Red,
		}
	}
}

func (sb *Model) SetSort(s string) {
	if s == "" {
		sb.sort = ""
	} else {
		sb.sort = fmt.Sprint("by ", s)
	}
	sb.bar.FourthColumn = fmt.Sprint(sb.count, sb.sort)
}

func (sb *Model) SetTitle(t string) {
	sb.bar.SecondColumn = t
}

func (sb *Model) GetTitle() string {
	return sb.bar.SecondColumn
}

func ellipsis(s string) string {
	maxLen := 30
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen < 3 {
		maxLen = 3
	}
	return string(runes[0:maxLen-1]) + "…"
}

func (sb *Model) SetMessage(m ...any) {
	sb.bar.ThirdColumn = ellipsis(fmt.Sprint(m...))
}

func (sb *Model) SetNumber(n int) {
	sb.count = n
	sb.bar.FourthColumn = fmt.Sprint(sb.count, sb.sort)
}

func New() Model {
	sb := statusbar.New(
		statusbar.ColorConfig{
			Foreground: style.White,
			Background: lipgloss.AdaptiveColor{},
		},
		statusbar.ColorConfig{
			Foreground: style.White,
			Background: lipgloss.AdaptiveColor{},
		},
		statusbar.ColorConfig{
			Foreground: style.White,
			Background: lipgloss.AdaptiveColor{},
		},
		statusbar.ColorConfig{
			Foreground: style.White,
			Background: lipgloss.AdaptiveColor{},
		},
	)
	m := Model{
		bar: &sb,
	}
	sb.SetContent("Syncing", "Inbox", "", "0")

	return m
}
