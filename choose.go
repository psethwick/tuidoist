package main

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/erikgeiser/promptkit/selection"
	filt "github.com/psethwick/tuidoist/filter"
	todoist "github.com/sachaos/todoist/lib"
)

type choosePurpose uint

const (
	chooseProject choosePurpose = iota
	moveToProject
	chooseFilter
)

type project todoist.Project

func (p project) Display() string {
	return p.Name
}

type chooseModel struct {
	chooser *selection.Model[selectable]
	main    *mainModel
	purpose choosePurpose
	title   string
}

type filter struct {
	Color     string
	ID        string
	IsDeleted bool
	ItemOrder int
	Name      string
	Query     string
}

func (f filter) Display() string {
	return f.Name
}

const (
	customTemplate = `
{{- if .Prompt -}}
  {{ Bold .Prompt }}
{{ end -}}
{{ if .IsFiltered }}
  {{- print .FilterPrompt " " .FilterInput }}
{{ end }}

{{- range  $i, $choice := .Choices }}
  {{- if IsScrollUpHintPosition $i }}
    {{- "⇡ " -}}
  {{- else if IsScrollDownHintPosition $i -}}
    {{- "⇣ " -}}
  {{- else -}}
    {{- "  " -}}
  {{- end -}}

  {{- if eq $.SelectedIndex $i }}
   {{- print (Foreground "32" (Bold "▸ ")) (Selected $choice) "\n" }}
  {{- else }}
    {{- print "  " (Unselected $choice) "\n" }}
  {{- end }}
{{- end}}`

	resultTemplate = `
	{{- print .Prompt " " (Final .FinalChoice) "\n" -}}
	`
)

type selectable interface {
	Display() string
}

func (pm *chooseModel) initChooser(p []selectable, prompt string) tea.Cmd {
	sel := selection.New(prompt, p)
	sm := selection.NewModel(sel)
	sm.Template = customTemplate

	// todo
	// sm.FilterInputTextStyle        lipgloss.Style
	// sm.FilterInputPlaceholderStyle lipgloss.Style
	// sm.FilterInputCursorStyle      lipgloss.Style
	sm.ResultTemplate = resultTemplate
	sm.Filter = func(filter string, choice *selection.Choice[selectable]) bool {
		// todo fuzzier matching would be cool
		return strings.Contains(strings.ToLower(choice.Value.Display()), strings.ToLower(filter))
	}
	sm.SelectedChoiceStyle = func(c *selection.Choice[selectable]) string {
		return selectedTitle.Render(c.Value.Display())
	}
	sm.UnselectedChoiceStyle = func(c *selection.Choice[selectable]) string {
		return normalTitle.Render(c.Value.Display())
	}
	pm.chooser = sm
	return sm.Init()
}

func (pm *chooseModel) View() string {
	return pm.chooser.View()
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

func (m *mainModel) OpenFilters() tea.Cmd {
	fls := make([]selectable, len(m.client.Store.Filters))
	for i, f := range m.client.Store.Filters {
		fls[i] = filter(f)
	}
	m.chooseModel.purpose = chooseFilter
	m.state = chooseState
	return m.chooseModel.initChooser(fls, "Choose Filter")
}

func (m *mainModel) OpenProjects(purpose choosePurpose) tea.Cmd {
	p := m.client.Store.Projects
	projs := make([]selectable, len(p))
	for i, prj := range p {
		projs[i] = project(prj)
	}
	var prompt string
	if purpose == chooseProject {
		prompt = "Switch Project"
	} else {
		prompt = "Move to Project"
	}
	m.chooseModel.purpose = purpose
	m.state = chooseState
	return m.chooseModel.initChooser(projs, prompt)
}

func (pm *chooseModel) handleChooseProject() tea.Cmd {
	p, err := pm.chooser.Value()
	prj := p.(project)
	var cmd tea.Cmd
	if err == nil {
		switch pm.purpose {
		case chooseProject:
			refresh := func() {
				pm.main.setTasks(&prj)
			}
			pm.main.tasksModel.refresh = refresh
			pm.main.tasksModel.tasks.FilterInput.SetValue("")
			ProjectID = prj.ID
			refresh()
			pm.main.switchProject(&prj)
		case moveToProject:
			task := pm.main.tasksModel.tasks.SelectedItem().(task)
			pm.main.tasksModel.tasks.RemoveItem(pm.main.tasksModel.tasks.Index())
			cmd = pm.main.MoveItem(&task.item, prj.ID)
		}
	}
	return cmd
}

func (pm *chooseModel) gotoFilter(f filter) tea.Cmd {
	expr := filt.Filter(f.Query)
	refresh := func() {
		pm.main.setTasksFromFilter(expr)
	}
	pm.main.tasksModel.tasks.FilterInput.SetValue("")
	pm.main.tasksModel.refresh = refresh
	pm.main.tasksModel.tasks.Title = f.Name
	refresh()
	return nil
}

func (pm *chooseModel) handleChooseFilter() tea.Cmd {
	f, err := pm.chooser.Value()
	if err != nil {
		dbg(err)
		return nil
	}
	flt := f.(filter)
	ProjectID = ""
	return pm.gotoFilter(flt)
}

func (pm *chooseModel) Update(msg tea.Msg) tea.Cmd {
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
			cmds = append(cmds, pm.chooser.Init())
			return tea.Batch(cmds...)
		case "esc":
			pm.main.state = tasksState
			return nil
		}
	}
	_, cmd := pm.chooser.Update(msg)
	cmds = append(cmds, cmd)
	return tea.Batch(cmds...)
}

func newChooseModel(m *mainModel) chooseModel {
	pm := chooseModel{}
	pm.main = m
	return pm
}

func (m *mainModel) switchProject(p *project) {
	m.tasksModel.tasks.Title = p.Name
	m.state = chooseState
}
