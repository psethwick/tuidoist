package status

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
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

func (sb *Model) SetSort(s string) {
	if s == "" {
		sb.bar.ThirdColumn = ""
	} else {
		sb.bar.ThirdColumn = fmt.Sprint("sort: ", s)
	}
}

func (sb *Model) SetTitle(t string) {
	sb.bar.SecondColumn = t
}

func (sb *Model) GetTitle() string {
	return sb.bar.SecondColumn
}

func (sb *Model) SetNumber(n int) {
	sb.bar.FourthColumn = fmt.Sprint(n)
}

func New() Model {
	sb := statusbar.New(
		statusbar.ColorConfig{
			Foreground: style.White,
			Background: style.Pink,
		},
		statusbar.ColorConfig{
			Foreground: style.White,
			Background: style.Gray,
		},
		statusbar.ColorConfig{
			Foreground: style.White,
			Background: style.Gray,
		},
		statusbar.ColorConfig{
			Foreground: style.White,
			Background: style.Pink,
		},
	)
	m := Model{
		bar: &sb,
	}
	sb.SetContent("Syncing", "Inbox", "", "0")

	return m
}
