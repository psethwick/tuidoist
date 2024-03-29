package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/erikgeiser/promptkit/selection"
	"github.com/psethwick/tuidoist/keys"
	"github.com/psethwick/tuidoist/style"

	todoist "github.com/sachaos/todoist/lib"
	filt "github.com/sachaos/todoist/lib/filter"
)

type choosePurpose uint

const (
	chooseProject choosePurpose = iota
	moveToProject
	chooseFilter
	choosePalette
)

type project struct {
	project todoist.Project
	section todoist.Section
}

func (p project) String() string {
	if p.section.ID == "" {
		return p.project.Name
	}
	return fmt.Sprintf("%s/%s", p.project.Name, p.section.Name)
}

type chooseModel struct {
	chooser  *selection.Model[fmt.Stringer]
	main     *mainModel
	purpose  choosePurpose
	oldTitle string
	keyMap   keys.InputKeyMap
}

type filter struct {
	todoist.Filter
}

func (f filter) String() string {
	return f.Name
}

const (
	customTemplate = `
{{ if .IsFiltered }}
  {{- print "   " .FilterInput }}
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
   {{- print (Foreground "32" (Bold "  ▸ ")) (Selected $choice) "\n" }}
  {{- else }}
    {{- print "    " (Unselected $choice) "\n" }}
  {{- end }}
{{- end}}`
)

func (pm *chooseModel) initChooser(p []fmt.Stringer, prompt string, purpose choosePurpose) tea.Cmd {
	sel := selection.New("", p)
	sm := selection.NewModel(sel)
	pm.oldTitle = pm.main.statusBarModel.GetTitle()
	pm.main.statusBarModel.SetTitle(prompt)
	sm.Template = customTemplate
	sm.PageSize = 20
	// todo
	// sm.FilterInputTextStyle        lipgloss.Style
	// sm.FilterInputPlaceholderStyle lipgloss.Style
	// sm.FilterInputCursorStyle      lipgloss.Style
	sm.Filter = func(filter string, choice *selection.Choice[fmt.Stringer]) bool {
		// todo fuzzier matching would be cool
		// https://github.com/charmbracelet/bubbles/blob/master/list/list.go#L87
		// short answer sahilm/fuzzy
		return strings.Contains(strings.ToLower(choice.Value.String()), strings.ToLower(filter))
	}
	sm.SelectedChoiceStyle = func(c *selection.Choice[fmt.Stringer]) string {
		return style.SelectedTitle.Render(c.Value.String())
	}
	sm.UnselectedChoiceStyle = func(c *selection.Choice[fmt.Stringer]) string {
		return style.NormalTitle.Render(c.Value.String())
	}
	pm.chooser = sm
	pm.purpose = purpose
	pm.main.state = viewChooser
	return sm.Init()
}

func (pm *chooseModel) View() string {
	return pm.chooser.View()
}

func (m *mainModel) OpenFilters() tea.Cmd {
	fls := make([]fmt.Stringer, len(m.local.Filters))
	for i, f := range m.local.Filters {
		fls[i] = filter{f}
	}
	if len(fls) == 0 {
		dbg("zero filters")
		return nil
	}
	return m.chooseModel.initChooser(fls, "Choose Filter", chooseFilter)
}

func (m *mainModel) OpenPalette() tea.Cmd {
	// if there's a task
	// if there's a project
	// if there's a section
	return m.chooseModel.initChooser(PaletteCommands, "Choose Command", choosePalette)
}

func (m *mainModel) OpenProjects(purpose choosePurpose) tea.Cmd {
	dbg("OpenProjects")
	p := m.local.Projects
	dbg(len(p))
	sections := m.local.Sections
	var projs []fmt.Stringer
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
	if len(projs) == 0 {
		// with zero choices chooser very rudely returns tea.Quit...
		dbg("zero projects")
		return nil
	}
	return m.chooseModel.initChooser(projs, prompt, purpose)
}

func (cm *chooseModel) gotoFilter(f filter) tea.Cmd {
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
			cm.main.statusBarModel.SetTitle(cm.oldTitle)
			return nil
		}
		fts = append(fts, filterTitle{titles[i], ex})
	}

	cm.main.refresh = func() {
		cm.main.setTasksFromFilter(filterSelection{fts, 0})
	}
	cm.main.refresh()
	cm.main.state = viewTasks
	return nil
}

func (cm *chooseModel) handleChoose() tea.Cmd {
	v, err := cm.chooser.Value()
	if err != nil {
		dbg(err)
		return nil
	}
	switch choosy := v.(type) {
	case filter:
		return cm.gotoFilter(choosy)
	case paletteCommand:
		return choosy.command(cm.main)
	case project:
		var cmds []tea.Cmd
		if err == nil {
			switch cm.purpose {
			case chooseProject:
				cm.main.refresh = func() {
					cm.main.setTasksFromProject(&choosy)
				}
				cm.main.refresh()
				cm.main.projectId = choosy.project.ID
				cm.main.sectionId = choosy.section.ID
			case moveToProject:
				cmds = append(cmds, cm.main.MoveItemsToProject(choosy))
				cm.main.statusBarModel.SetTitle(cm.oldTitle)
			}
		}
		cm.main.state = viewTasks
		return tea.Batch(cmds...)
	}
	return nil
}

func (pm *chooseModel) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, pm.keyMap.Accept):
			cmds = append(cmds, pm.handleChoose())
			return tea.Batch(cmds...)
		case key.Matches(msg, pm.keyMap.Cancel):
			pm.main.state = viewTasks
			pm.main.statusBarModel.SetTitle(pm.oldTitle)
			return nil
		}
	}
	_, cmd := pm.chooser.Update(msg)
	cmds = append(cmds, cmd)
	return tea.Batch(cmds...)
}

func newChooseModel(m *mainModel, km keys.InputKeyMap) chooseModel {
	return chooseModel{main: m, keyMap: km}
}
