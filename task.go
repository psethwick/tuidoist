package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sachaos/todoist/lib"
)

type task struct {
	item      todoist.Item
	title     string
	desc      string
	completed bool
}

func reformatDate(d string, from string, to string) string {
	t, err := time.Parse(from, d)
	if err != nil {
		dbg(err)
	}
	return t.Format(to)
}

// todo labels
// heres where reminders is
// todoist.Store.Reminders
// ⏰ ??
func newTask(m *mainModel, item todoist.Item) task {
	indent := strings.Repeat(" ", len(todoist.SearchItemParents(m.client.Store, &item)))
	var checkbox string
	switch item.Priority {
	// mirror the priority colors from the webapp
	case 1:
		checkbox = " ⚪ "
	case 2:
		checkbox = " 🔵 "
	case 3:
		checkbox = " 🟡 "
	case 4:
		checkbox = " 🔴 "
	}
	if indent != "" {
		// subtask indicator
		checkbox = fmt.Sprint("╰", checkbox)
	}
	title := fmt.Sprint(indent, checkbox, item.Content)
	desc := ""
	if item.Due != nil {
		desc += " 🗓️ "
		var fd string
		if strings.Contains(item.Due.Date, "T") {
			fd = reformatDate(item.Due.Date, "2006-01-02T15:04:05", "02 Jan 06 15:04")
		} else {
			fd = reformatDate(item.Due.Date, "2006-01-02", "02 Jan 06")
		}
		if item.Due.IsRecurring {
			desc += " 🔁"
		}
		desc += fd
	}
	desc = fmt.Sprint(indent, desc)
	return task{
		item:  item,
		title: title,
		desc:  desc,
	}
}

func (t task) Title() string {
	if t.completed {
		return strikeThroughStyle.Render(t.title)
	}
	return t.title
}

func (t task) Description() string {
	return t.desc
}

func (t task) FilterValue() string { return t.item.Content }

// todo confirm
func (m *mainModel) deleteTask() func() tea.Msg {
	t := m.tasks.SelectedItem().(task)
	m.tasks.RemoveItem(m.tasks.Index())
	return func() tea.Msg {
		err := m.client.DeleteItem(m.ctx, []string{t.item.ID})
		if err != nil {
			dbg("del err", err)
		}
		return m.sync()
	}
}

func (m *mainModel) completeTask() func() tea.Msg {
	t := m.tasks.SelectedItem().(task)
	t.completed = true
	m.tasks.SetItem(m.tasks.Index(), t)
	return func() tea.Msg {
		err := m.client.CloseItem(m.ctx, []string{t.item.ID})
		if err != nil {
			dbg("complete task err", err)
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
    t.Priority = 1
	m.tasks.InsertItem(len(m.client.Store.Items)+1, newTask(m, t))
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
