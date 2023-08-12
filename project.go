package main

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/erikgeiser/promptkit/selection"
	"github.com/sachaos/todoist/lib"
)

type projectsModel struct {
	projects *selection.Model[todoist.Project]
}

func (pm *projectsModel) initSelect(p []todoist.Project) {
	sel := selection.New("Choose Project:", p)
	sm := selection.NewModel(sel)
    pm.projects = sm
	sm.Filter = func(filter string, choice *selection.Choice[todoist.Project]) bool {
		return strings.Contains(strings.ToLower(choice.Value.Name), strings.ToLower(filter))
	}
	sm.SelectedChoiceStyle = func(c *selection.Choice[todoist.Project]) string {
		return c.Value.Name
	}
	sm.UnselectedChoiceStyle = func(c *selection.Choice[todoist.Project]) string {
		return c.Value.Name
	}
	sm.Init()
}

func newProjectsModel(m *mainModel) projectsModel {
    pm := projectsModel{}
    pm.initSelect(m.client.Store.Projects)
	return pm
}

func (m *mainModel) switchProject(p *todoist.Project) {
	m.tasksModel.tasks.Title = p.Name
	m.projectId = p.ID
	m.state = projectState
}

func (m *mainModel) SetProjects(p []todoist.Project) tea.Cmd {
    m.projectsModel.initSelect(p)
    return nil
}
