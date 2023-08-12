package main

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type newTaskModel struct {
	input textinput.Model
	main  *mainModel
}

func newNewTaskModel(m *mainModel) newTaskModel {
	return newTaskModel{
		textinput.New(),
		m,
	}
}

func (ntm *newTaskModel) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			cmds = append(cmds, ntm.main.addTask())
		case "esc":
			ntm.input.SetValue("")
			ntm.main.state = tasksState
		}
	}
	input, cmd := ntm.input.Update(msg)
	ntm.input = input
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

func (ntm *newTaskModel) View() string {
	return ntm.input.View()
}
