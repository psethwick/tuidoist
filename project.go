package main

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/sachaos/todoist/lib"
)

func (m *mainModel) setTasks(p *todoist.Project) {
	tasks := []list.Item{}
	for _, i := range m.client.Store.Items {
		if i.ProjectID == p.ID {
			tasks = append(tasks, newTask(m, i))
		}
	}
	m.tasks.SetItems(tasks)
}

func (m *mainModel) switchProject(p *todoist.Project) {
	m.tasks.Title = p.Name
	m.projectId = p.ID
	m.state = tasksView
}
