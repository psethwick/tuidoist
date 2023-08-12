package main

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/sachaos/todoist/lib"
	"github.com/erikgeiser/promptkit/selection"
)

type projectsModel struct {
	projects  *selection.Model[todoist.Project]
}

func (m *mainModel) setTasks(p *todoist.Project) {
	tasks := []list.Item{}
	for _, i := range m.client.Store.Items {
		if i.ProjectID == p.ID {
			tasks = append(tasks, newTask(m, i))
		}
	}
	m.tasksModel.tasks.SetItems(tasks)
}

/// set/switch are separate so we can sync + update in the background
func (m *mainModel) switchProject(p *todoist.Project) {
	m.tasksModel.tasks.Title = p.Name
	m.projectId = p.ID
	m.state = projectState
}
