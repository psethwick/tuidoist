package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sachaos/todoist/lib"
)

type tasksModel struct {
	tasks list.Model
	main  *mainModel
}

func newTasksModel(m *mainModel) tasksModel {
	tasks := list.New([]list.Item{}, taskDelegate(m), 40, 30)
	tasks.DisableQuitKeybindings()
	return tasksModel{
		tasks,
		m,
	}
}

type task struct {
	item      todoist.Item
	title     string
	summary   string
	completed bool
}

func reformatDate(d string, from string, to string) string {
	t, err := time.Parse(from, d)
	if err != nil {
		dbg(err)
	}
	return t.Format(to)
}

// heres where reminders is
// todoist.Store.Reminders
// ⏰ ??
// todo overdue should be red somewhere
// today, maybe also highlighted?
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
	labels := ""
	for _, l := range item.LabelNames {
		labels += fmt.Sprint(" 🏷️ ", l)
	}
	summary := ""
	if item.Due != nil {
		summary += " 🗓️ "
		var fd string
		if strings.Contains(item.Due.Date, "T") {
			fd = reformatDate(item.Due.Date, "2006-01-02T15:04:05", "02 Jan 06 15:04")
		} else {
			fd = reformatDate(item.Due.Date, "2006-01-02", "02 Jan 06")
		}
		if item.Due.IsRecurring {
			checkbox += "🔁 "
		}
		summary += fd
	}
	title := fmt.Sprint(indent, checkbox, item.Content, labels)
	summary = fmt.Sprint(indent, "   ", summary)
	return task{
		item:    item,
		title:   title,
		summary: summary,
	}
}

func (t task) Title() string {
	if t.completed {
		return strikeThroughStyle.Render(t.title)
	}
	return t.title
}

func (t task) Description() string {
	return t.summary
}

func (t task) FilterValue() string { return t.item.Content }

func (m *mainModel) setTasks(p *todoist.Project) {
	tasks := []list.Item{}
	for _, i := range m.client.Store.Items {
		if i.ProjectID == p.ID {
			tasks = append(tasks, newTask(m, i))
		}
	}
	m.tasksModel.tasks.SetItems(tasks)
}

// todo confirm
func (m *mainModel) deleteTask() func() tea.Msg {
	t := m.tasksModel.tasks.SelectedItem().(task)
	m.tasksModel.tasks.RemoveItem(m.tasksModel.tasks.Index())
	return func() tea.Msg {
		err := m.client.DeleteItem(m.ctx, []string{t.item.ID})
		if err != nil {
			dbg("del err", err)
		}
		return m.sync()
	}
}

func (m *mainModel) completeTask() func() tea.Msg {
	t := m.tasksModel.tasks.SelectedItem().(task)
	t.completed = true
	m.tasksModel.tasks.SetItem(m.tasksModel.tasks.Index(), t)
	return func() tea.Msg {
		err := m.client.CloseItem(m.ctx, []string{t.item.ID})
		if err != nil {
			dbg("complete task err", err)
		}
		return m.sync()
	}
}

func (tm *tasksModel) View() string {
	return listStyle.Render(tm.tasks.View())
}

func (tm *tasksModel) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			cmds = append(cmds, tea.Quit)
		case "r":
			cmds = append(cmds, tm.main.sync)
		}
	}
	tasks, cmd := tm.tasks.Update(msg)
	tm.tasks = tasks
	cmds = append(cmds, cmd)
	return tea.Batch(cmds...)
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
				m.projectsModel.purpose = chooseProject
				m.state = projectState
			case "v":
				cmds = append(cmds, tea.ClearScreen)
				m.projectsModel.purpose = moveToProject
				m.state = projectState
			case "C":
				cmds = append(cmds, m.completeTask())
			case "D":
				cmds = append(cmds, m.deleteTask())
			case "n":
				m.newTaskModel.content.Prompt = "> "
				m.newTaskModel.content.Focus()
				cmds = append(cmds, tea.ClearScreen)
				m.state = newTaskState
			}
		}

		return tea.Batch(cmds...)
	}
	return d
}
