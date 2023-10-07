package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/erikgeiser/promptkit/selection"
	"github.com/psethwick/tuidoist/style"

	todoist "github.com/sachaos/todoist/lib"
	filt "github.com/sachaos/todoist/lib/filter"
)

type choosePurpose uint

const (
	chooseProject choosePurpose = iota
	moveToProject
	chooseFilter
	palette
)

type project struct {
	project todoist.Project
	section todoist.Section
}

func (p project) Display() string {
	if p.section.ID == "" {
		return p.project.Name
	}
	return fmt.Sprintf("%s/%s", p.project.Name, p.section.Name)
}

type chooseModel struct {
	chooser  *selection.Model[selectable]
	main     *mainModel
	purpose  choosePurpose
	title    string
	oldTitle string
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
{{- print "\n"}}
{{ if .IsFiltered }}
  {{- print "   " .FilterInput }}
{{ end }}
    {{- print "\n"}}

{{- range  $i, $choice := .Choices }}
  {{- if IsScrollUpHintPosition $i }}
    {{- "⇡ " -}}
  {{- else if IsScrollDownHintPosition $i -}}
    {{- "⇣ " -}}
  {{- else -}}
    {{- "  " -}}
  {{- end -}}

  {{- if eq $.SelectedIndex $i }}
   {{- print (Foreground "32" (Bold "  ▸ ")) (Selected $choice) "\n\n" }}
  {{- else }}
    {{- print "    " (Unselected $choice) "\n\n" }}
  {{- end }}
{{- end}}`
)

type selectable interface {
	Display() string
}

func (pm *chooseModel) initChooser(p []selectable, prompt string, purpose choosePurpose) tea.Cmd {
	sel := selection.New("", p)
	sm := selection.NewModel(sel)
	pm.oldTitle = pm.main.statusBarModel.GetTitle()
	pm.main.statusBarModel.SetTitle(prompt)
	sm.Template = customTemplate
	// we're double spacing + some room for the prompt
	sm.PageSize = pm.main.height/2 - 3
	// todo
	// sm.FilterInputTextStyle        lipgloss.Style
	// sm.FilterInputPlaceholderStyle lipgloss.Style
	// sm.FilterInputCursorStyle      lipgloss.Style
	sm.Filter = func(filter string, choice *selection.Choice[selectable]) bool {
		// todo fuzzier matching would be cool
		// https://github.com/charmbracelet/bubbles/blob/master/list/list.go#L87
		// short answer sahilm/fuzzy
		return strings.Contains(strings.ToLower(choice.Value.Display()), strings.ToLower(filter))
	}
	sm.SelectedChoiceStyle = func(c *selection.Choice[selectable]) string {
		return style.SelectedTitle.Render(c.Value.Display())
	}
	sm.UnselectedChoiceStyle = func(c *selection.Choice[selectable]) string {
		return style.NormalTitle.Render(c.Value.Display())
	}
	pm.chooser = sm
	pm.purpose = purpose
	return sm.Init()
}

func (pm *chooseModel) View() string {
	return pm.chooser.View()
}

func (m *mainModel) MoveItem(item *todoist.Item, p project) func() tea.Msg {
	return func() tea.Msg {
		args := map[string]interface{}{"id": item.ID}
		if p.section.ID != "" {
			args["section_id"] = p.section.ID
		} else {
			args["project_id"] = p.project.ID
		}
		err := m.client.ExecCommands(
			m.ctx,
			todoist.Commands{todoist.NewCommand("item_move", args)},
		)
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
	m.state = chooseState
	return m.chooseModel.initChooser(fls, "Choose Filter", chooseFilter)
}
func (m *mainModel) OpenPalette() tea.Cmd {
	// fls := make([]selectable, len(m.client.Store.Filters))
	// for i, f := range m.client.Store.Filters {
	// 	fls[i] = filter(f)
	// }
	// m.state = chooseState
	// return m.chooseModel.initChooser(PaletteCommands, "Choose Filter", chooseFilter)
	return nil
}

func (m *mainModel) OpenProjects(purpose choosePurpose) tea.Cmd {
	p := m.client.Store.Projects
	sections := m.client.Store.Sections
	var projs []selectable
	for _, prj := range p {
		var projectSections []todoist.Section
		for _, s := range sections {
			if s.ProjectID == prj.ID {
				projectSections = append(projectSections, s)
			}
		}
		// always add the 'whole' project even if there's sections
		projs = append(projs, project{prj, todoist.Section{}})
		for _, s := range projectSections {
			projs = append(projs, project{prj, s})
		}
	}
	var prompt string
	if purpose == chooseProject {
		prompt = "Choose Project"
	} else {
		prompt = "Move to Project"
	}
	m.state = chooseState
	return m.chooseModel.initChooser(projs, prompt, purpose)
}

func (pm *chooseModel) handleChooseProject() tea.Cmd {
	p, err := pm.chooser.Value()
	prj, ok := p.(project)
	if !ok {
		return nil
	}
	var cmds []tea.Cmd
	if err == nil {
		switch pm.purpose {
		case chooseProject:
			pm.main.refresh = func() {
				pm.main.setTasksFromProject(&prj)
			}
			pm.main.newTaskModel.projectID = prj.project.ID
			pm.main.newTaskModel.sectionId = prj.section.ID
			pm.main.refresh()
			pm.main.switchProject(&prj)
		case moveToProject:
			task, err := pm.main.taskList.RemoveCurrentItem()
			if err != nil {
				dbg(err)
			}
			if ok {
				cmds = append(cmds, pm.main.MoveItem(&task.Item, prj))
				pm.main.statusBarModel.SetTitle(pm.oldTitle)
			}
		}
	}
	pm.main.state = tasksState
	return tea.Batch(cmds...)
}

func (pm *chooseModel) gotoFilter(f filter) tea.Cmd {
	exprs := filt.Filter(f.Query)
	titles := strings.Split(f.Query, ",")

	if len(titles) == 1 {
		titles[0] = f.Name
	}
	var fts []filterTitle
	for i, ex := range exprs {
		dbg(i, fmt.Sprintf("%+v", ex))
		if _, ok := ex.(filt.ErrorExpr); ok {
			dbg("error", ex, f.Query)
			pm.main.statusBarModel.SetTitle(pm.oldTitle)
			return nil
		}
		fts = append(fts, filterTitle{titles[i], ex})
	}

	pm.main.refresh = func() {
		pm.main.setTasksFromFilter(fts)
	}
	pm.main.refresh()
	return nil
}

func (pm *chooseModel) handleChooseFilter() tea.Cmd {
	f, err := pm.chooser.Value()
	if err != nil {
		dbg(err)
		return nil
	}
	flt := f.(filter)
	pm.main.state = tasksState
	return pm.gotoFilter(flt)
}

func (cm *chooseModel) handleChooseCommand() tea.Cmd {
	// c, err := cm.chooser.Value()
	return nil
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
			case palette:
				cmds = append(cmds, pm.handleChooseCommand())
			}
			return tea.Batch(cmds...)
		case "esc":
			pm.main.state = tasksState
			pm.main.statusBarModel.SetTitle(pm.oldTitle)
			return nil
		}
	}
	_, cmd := pm.chooser.Update(msg)
	cmds = append(cmds, cmd)
	return tea.Batch(cmds...)
}

func newChooseModel(m *mainModel) chooseModel {
	return chooseModel{main: m}
}

func (m *mainModel) switchProject(p *project) {
	m.state = chooseState
}
