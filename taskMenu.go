package main

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	todoist "github.com/sachaos/todoist/lib"
)

type taskMenuModel struct {
	content textinput.Model
	desc    textinput.Model
	item    todoist.Item
	project *todoist.Project
	focus   taskMenuFocus
	main    *mainModel
}

type taskMenuFocus uint

const (
	focusBox taskMenuFocus = iota
	focusContent
	focusDesc
)

func newTaskMenuModel(m *mainModel) taskMenuModel {
	return taskMenuModel{
		textinput.New(),
		textinput.New(),
		todoist.Item{},
		nil,
		focusBox,
		m,
	}
}

func (m *mainModel) UpdateItem(i todoist.Item) func() tea.Msg {
	return func() tea.Msg {
		m.client.UpdateItem(m.ctx, i)
		return m.sync()
	}
}

func (tm *taskMenuModel) updateFocus() {
	switch tm.focus {
	case focusBox:
		tm.content.Blur()
		tm.desc.Blur()
	case focusContent:
		tm.content.Focus()
		tm.desc.Blur()
	case focusDesc:
		tm.content.Blur()
		tm.desc.Focus()
	}
}

func (tm *taskMenuModel) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "v":
			t := tm.main.tasksModel.tasks.SelectedItem().(task)
			if t.url != "" {
				cmds = append(cmds, tm.main.tasksModel.OpenUrl(t.url))
			}
		case "m":
			cmds = append(cmds, tm.main.OpenProjects(moveToProject))
		case "c":
			cmds = append(cmds, tm.main.completeTask())
			// todo tick, wait then exit
			tm.main.state = tasksState
		case "d":
			// TODO confirmation
			cmds = append(cmds, tm.main.deleteTask())
		case "enter":
			tm.item.Content = tm.content.Value()
			tm.item.Description = tm.desc.Value()
			cmds = append(cmds, tm.main.UpdateItem(tm.item))
			tm.main.state = tasksState
		case "tab":
			tm.focus = (tm.focus + 1) % 3
			tm.updateFocus()
		case "shift+tab":
			tm.focus = (tm.focus - 1) % 3
			tm.updateFocus()
		case "esc":
			tm.main.state = tasksState

		}
	}
	input, cmd := tm.content.Update(msg)
	tm.content = input
	cmds = append(cmds, cmd)
	input, cmd = tm.desc.Update(msg)
	tm.desc = input
	cmds = append(cmds, cmd)
	return tea.Batch(cmds...)
}

func (tm *taskMenuModel) View() string {
	title := dialogTitle.Render("Task")
	help := helpStyle.Render("(e)dit (c)complete (m)ove (d)elete")
	ui := lipgloss.JoinVertical(lipgloss.Left, title, tm.content.View(), tm.desc.View(), tm.project.Name, help)

	dialog := lipgloss.Place(tm.main.size.Width, tm.main.size.Height,
		lipgloss.Left, lipgloss.Center,
		dialogBoxStyle.Render(ui),
	)

	return dialog
}
