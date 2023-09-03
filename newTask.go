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
	if ntm.main.state == newTaskBottomState {
		maxOrder := 0
		for _, t := range ntm.main.taskList.List.GetAllItems() {
			task := t.(task)
			maxOrder = max(maxOrder, task.item.ChildOrder)
		}
		t.ChildOrder = maxOrder + 1
		ntm.main.taskList.List.Bottom()
	} else {
		minOrder := 0
		for _, t := range ntm.main.taskList.List.GetAllItems() {
			task := t.(task)
			minOrder = min(minOrder, task.item.ChildOrder)
		}
		t.ChildOrder = minOrder - 1
		ntm.main.taskList.List.Top()
	}
	ntm.main.taskList.List.AddItems(newTask(ntm.main, t))
	ntm.main.taskList.List.Sort()
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
			ntm.main.state = tasksState
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
