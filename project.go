package main

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/erikgeiser/promptkit/selection"
	filt "github.com/psethwick/tuidoist/filter"
	"github.com/sachaos/todoist/lib"
)

type projectPurpose uint

const (
	chooseProject projectPurpose = iota
	moveToProject
	chooseFilter
)

type projectsModel struct {
	projects *selection.Model[todoist.Project]
	filters  *selection.Model[filter]
	main     *mainModel
	purpose  projectPurpose
	title    string
}

type filter struct {
	Color     string
	ID        string
	IsDeleted bool
	ItemOrder int
	Name      string
	Query     string
}

func (pm *projectsModel) initSelectFilters(p []filter) {
	sel := selection.New("Select Filter", p)
	sm := selection.NewModel(sel)
	sm.PageSize = 10
	sm.Filter = func(filter string, choice *selection.Choice[filter]) bool {
		// todo fuzzier matching would be cool
		return strings.Contains(strings.ToLower(choice.Value.Name), strings.ToLower(filter))
	}
	sm.SelectedChoiceStyle = func(c *selection.Choice[filter]) string {
		return c.Value.Name
	}
	sm.UnselectedChoiceStyle = func(c *selection.Choice[filter]) string {
		return c.Value.Name
	}
	pm.filters = sm
	sm.Init()
}

func (pm *projectsModel) initSelectProjects(p []todoist.Project) {
	sel := selection.New("Switch Project", p)
	sm := selection.NewModel(sel)
	sm.PageSize = 10
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
	switch pm.purpose {
	case chooseProject:
		fallthrough
	case moveToProject:
		return listStyle.Render(pm.projects.View())
	case chooseFilter:
		return listStyle.Render(pm.filters.View())
	}
	return "aaaahhh"
}

func (m *mainModel) MoveItem(item *todoist.Item, projectId string) func() tea.Msg {
	return func() tea.Msg {
		err := m.client.MoveItem(m.ctx, item, projectId)
		if err != nil {
			dbg(err)
		}
		err = m.client.Sync(m.ctx)
		if err != nil {
			dbg(err)
		}
		return nil
	}
}

func (m *mainModel) OpenFilters() {
	m.projectsModel.purpose = chooseFilter
	m.state = projectState
}

func (m *mainModel) OpenProjects(purpose projectPurpose, title string, prompt string) {
	m.projectsModel.purpose = purpose
	m.projectsModel.title = title
	m.projectsModel.projects.Prompt = prompt
	m.state = projectState
}

func (pm *projectsModel) handleChooseProject() tea.Cmd {
	p, err := pm.projects.Value()
	var cmd tea.Cmd
	if err == nil {
		switch pm.purpose {
		case chooseProject:
			pm.main.projectId = p.ID
			pm.main.setTasks(&p)
			pm.main.switchProject(&p)
		case moveToProject:
			task := pm.main.tasksModel.tasks.SelectedItem().(task)
			pm.main.tasksModel.tasks.RemoveItem(pm.main.tasksModel.tasks.Index())
			cmd = pm.main.MoveItem(&task.item, p.ID)
		}
	}
	return cmd
}

func (pm *projectsModel) handleChooseFilter() tea.Cmd {
	f, err := pm.filters.Value()
	expr := filt.Filter(f.Query)
	if err != nil {
		dbg(err)
		return nil
	}
	pm.main.setTasksFromFilter(expr)

	return nil
}

func (pm *projectsModel) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			switch pm.purpose {
			case chooseProject:
				fallthrough
			case moveToProject:
				cmds = append(cmds, pm.handleChooseProject())
			case chooseFilter:
				cmds = append(cmds, pm.handleChooseFilter())
			}
			pm.main.state = tasksState
			return tea.Batch(pm.projects.Init(), pm.filters.Init())
		case "esc":
			pm.main.state = tasksState
			return nil
		}
	}
	_, cmd := pm.filters.Update(msg)
	cmds = append(cmds, cmd)
	_, cmd = pm.projects.Update(msg)
	cmds = append(cmds, cmd)
	return tea.Batch(cmds...)
}

func newProjectsModel(m *mainModel) projectsModel {
	pm := projectsModel{}
	pm.main = m
	var filters []filter
	for _, f := range m.client.Store.Filters {
		filters = append(filters, filter(f))
	}
	pm.initSelectFilters(filters)
	pm.initSelectProjects(m.client.Store.Projects)
	return pm
}

func (m *mainModel) switchProject(p *todoist.Project) {
	m.tasksModel.tasks.Title = p.Name
	m.projectId = p.ID
	m.state = projectState
}

func (m *mainModel) SetProjects(p []todoist.Project) tea.Cmd {
	m.projectsModel.initSelectProjects(p)
	return nil
}
