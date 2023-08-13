package main

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/erikgeiser/promptkit/selection"
	"github.com/sachaos/todoist/lib"
)

type projectPurpose uint

const (
	chooseProject projectPurpose = iota
	moveToProject
)

type projectsModel struct {
	projects *selection.Model[todoist.Project]
	main     *mainModel
	purpose  projectPurpose
}

func (pm *projectsModel) initSelect(p []todoist.Project) {
	sel := selection.New("Choose Project:", p)
	sm := selection.NewModel(sel)
	sm.Filter = func(filter string, choice *selection.Choice[todoist.Project]) bool {
		// todo fuzzier matching would be cool
		return strings.Contains(strings.ToLower(choice.Value.Name), strings.ToLower(filter))
	}
	sm.SelectedChoiceStyle = func(c *selection.Choice[todoist.Project]) string {
		return c.Value.Name
	}
	sm.UnselectedChoiceStyle = func(c *selection.Choice[todoist.Project]) string {
		return c.Value.Name
	}
	pm.projects = sm
	sm.Init()
}

func (pm *projectsModel) View() string {
	return listStyle.Render(pm.projects.View())
}

func (pm *projectsModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			p, err := pm.projects.Value()
			if err == nil {
				switch pm.purpose {
				case chooseProject:
					pm.main.projectId = p.ID
					pm.main.setTasks(&p)
					pm.main.switchProject(&p)
				case moveToProject:
					task := pm.main.tasksModel.tasks.SelectedItem().(task)
					err := pm.main.client.MoveItem(pm.main.ctx, &task.item, p.ID)
					dbg(err)
					err = pm.main.client.Sync(pm.main.ctx)
					dbg(err)
				}
			}
			pm.main.state = tasksState
			return pm.projects.Init()
		case "esc":
			pm.main.state = tasksState
			return nil
		}
	}
	_, cmd := pm.projects.Update(msg)
	return cmd
}

func newProjectsModel(m *mainModel) projectsModel {
	pm := projectsModel{}
	pm.main = m
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
