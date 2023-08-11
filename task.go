package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sachaos/todoist/lib"
)

type task todoist.Item

func (i task) indent() string {
	return strings.Repeat("  ", i.Indent)
}

func (i task) Title() string {
	prefix := "⚪ "
	if i.Indent > 0 {
		prefix = "╰ ⚪ "
	}
	return fmt.Sprintf("%s%s%s", i.indent(), prefix, i.Content)
}

// todo due date, alarm sigil
// todo priority
func (i task) Description() string {
	// ⏰ ??
    duestring := ""
    if i.Due != nil {
        log.Printf("%s", i.Due.Date)
        duestring += "Due: "
        // heres where reminders is
        // todoist.Store.Reminders
        duestring += i.Due.Date
    }
	return fmt.Sprintf("%s%s%s", i.indent(), i.GetDescription(), duestring)
}

func (i task) FilterValue() string { return i.Content }

func (m *mainModel) deleteTask() func() tea.Msg {
	t := m.tasks.SelectedItem().(task)
	m.tasks.RemoveItem(m.tasks.Index())
	return func() tea.Msg {
		err := m.client.DeleteItem(m.ctx, []string{t.ID})
		if err != nil {
			log.Printf("%s", err)
		}
		return m.sync()
	}
}

func (m *mainModel) completeTask() func() tea.Msg {
	t := m.tasks.SelectedItem().(task)
	m.tasks.RemoveItem(m.tasks.Index())
	return func() tea.Msg {
		err := m.client.CloseItem(m.ctx, []string{t.ID})
		if err != nil {
			log.Printf("%s", err)
		}
		return m.sync()
	}
}

func (m *mainModel) addTask() func() tea.Msg {
	content := m.newTask.Value()
	m.newTask.SetValue("")
	if content == "" {
		return func() tea.Msg { return nil }
	}
	t := todoist.Item{}
	t.ProjectID = m.projectId
	t.Content = content
	// todo priority, description, labels
	m.tasks.InsertItem(len(m.client.Store.Items)+1, task(t))
	return func() tea.Msg {
		m.client.AddItem(m.ctx, t)
		return m.sync()
	}
}

func taskDelegate(m *mainModel) list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	d.UpdateFunc = func(msg tea.Msg, l *list.Model) tea.Cmd {
		var cmds []tea.Cmd
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "p":
				cmds = append(cmds, tea.ClearScreen)
				m.state = projectView
			case "C":
				cmds = append(cmds, m.completeTask())
			case "D":
				cmds = append(cmds, m.deleteTask())
			case "n":
				m.newTask.Prompt = "> "
				m.newTask.Focus()
				cmds = append(cmds, tea.ClearScreen)
				m.state = newTaskView
			}
		}

		return tea.Batch(cmds...)
	}
	return d
}
