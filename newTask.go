package main

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	todoist "github.com/sachaos/todoist/lib"
)


type newTaskModel struct {
	content textinput.Model
	main    *mainModel
}

func newNewTaskModel(m *mainModel) newTaskModel {
	return newTaskModel{
		textinput.New(),
		m,
	}
}

func (ntm *newTaskModel) Height() int {
	return 8 // height of textinput + dialog
}

func (ntm *newTaskModel) addTask() func() tea.Msg {
	content := ntm.content.Value()
	ntm.content.SetValue("")
	if content == "" {
		return func() tea.Msg { return nil }
	}
	t := todoist.Item{}
	t.ProjectID = ntm.main.projectId
	t.Content = content
	t.Priority = 1
	ntm.main.tasksModel.tasks.InsertItem(len(ntm.main.client.Store.Items)+1, newTask(ntm.main, t))
	return func() tea.Msg {
		// todo separate quick add?
		ntm.main.client.AddItem(ntm.main.ctx, t)
		return ntm.main.sync()
	}
}

func (ntm *newTaskModel) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			cmds = append(cmds, ntm.addTask())
		case "esc":
			ntm.content.SetValue("")
			ntm.main.tasksModel.Focus()
		}
	}
	input, cmd := ntm.content.Update(msg)
	ntm.content = input
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

func (ntm *newTaskModel) View() string {
	title := dialogTitle.Render("Add Task")
	help := helpStyle.Render("esc cancels           enter accepts")
	ui := lipgloss.JoinVertical(lipgloss.Left, title, ntm.content.View(), "", help)

	dialog := lipgloss.Place(ntm.main.size.Width, 5,
		lipgloss.Left, lipgloss.Left,
		dialogBoxStyle.Render(ui),
	)

	return dialog
}
