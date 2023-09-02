package main

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/psethwick/tuidoist/todoist"
)

type newTaskModel struct {
	content textinput.Model
	main    *mainModel
}

var ProjectID = ""

func newNewTaskModel(m *mainModel) newTaskModel {
	ti := textinput.New()
	ti.Prompt = "   > "
	return newTaskModel{
		ti,
		m,
	}
}

func (ntm *newTaskModel) Height() int {
	return 4 // height of textinput + dialog
}

func (ntm *newTaskModel) addTask() func() tea.Msg {
	content := ntm.content.Value()
	ntm.content.SetValue("")
	if content == "" {
		return func() tea.Msg { return nil }
	}
	t := todoist.Item{}
	t.Content = content
	t.Priority = 1
	if ProjectID != "" {
		t.ProjectID = ProjectID
	}
	var idx int
	if ntm.main.state == newTaskBottomState {
		idx = len(ntm.main.tasksModel.tasks.Items()) + 1
	} else {
		minOrder := 0
		for _, t := range ntm.main.tasksModel.tasks.Items() {
			task := t.(task)
			minOrder = min(minOrder, task.item.ChildOrder)
		}
		t.ChildOrder = minOrder - 1
		idx = 0
	}
	ntm.main.tasksModel.tasks.InsertItem(idx, newTask(ntm.main, t))
	ntm.main.tasksModel.tasks.Select(idx)
	return func() tea.Msg {
		// todo separate quick add?
		ntm.main.client.AddItem(ntm.main.ctx, t)
		return ntm.main.sync()
	}
}

// Save a newly created task and create a new one below Enter
// Save changes to an existing task and create a new task below ⇧ Enter
// Save a new task or save changes to an existing one and create a new task above Ctrl Enter
func (ntm *newTaskModel) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			cmds = append(cmds, ntm.addTask())
		case "esc":
			ntm.content.SetValue("")
			ntm.main.taskList.List.Height = ntm.main.height
			ntm.content.Blur()
		}
	}
	input, cmd := ntm.content.Update(msg)
	ntm.content = input
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

func (ntm *newTaskModel) View() string {
	dialog := lipgloss.Place(ntm.main.width, 3,
		lipgloss.Left, lipgloss.Left,
		dialogBoxStyle.Render(ntm.content.View()),
	)

	return dialog
}
