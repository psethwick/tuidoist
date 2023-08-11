package main

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sachaos/todoist/lib"
)

type project todoist.Project

func (p project) Title() string {
	return p.Name
}
func (p project) Description() string { return p.Name }
func (p project) FilterValue() string { return p.Name }

func (m *mainModel) setTasks(p *project) {
	tasks := []list.Item{}
	for _, i := range m.client.Store.Items {
		if i.ProjectID == p.ID {
			tasks = append(tasks, newTask(m, i))
		}
	}
	m.tasks.SetItems(tasks)
}

func (m *mainModel) switchProject(p *project) {
	m.tasks.Title = p.Name
	m.projectId = p.ID
	m.state = tasksView
}

func projectDelegate(m *mainModel) list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	d.ShowDescription = false
	d.UpdateFunc = func(msg tea.Msg, l *list.Model) tea.Cmd {
		var cmds []tea.Cmd
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				cmds = append(cmds, tea.ClearScreen)
				m.state = tasksView
			case "enter":
				if p, ok := l.SelectedItem().(project); ok {
					m.setTasks(&p)
                    m.switchProject(&p)
				}
			}
		}
		return tea.Batch(cmds...)
	}
	return d
}
