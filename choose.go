package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	title    string
	oldTitle string
}

type filter struct {
	todoist.Filter
}

func (f filter) String() string {
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
	// we're double spacing + some room for the prompt
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
	dialogBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(1, 0).
		BorderTop(true).
		BorderLeft(true).
		BorderRight(true).
		BorderBottom(true)
	return dialogBoxStyle.Render(pm.chooser.View())
}

func (m *mainModel) OpenFilters() tea.Cmd {
	fls := make([]fmt.Stringer, len(m.local.Filters))
	for i, f := range m.local.Filters {
		fls[i] = filter{f}
	}
	return m.chooseModel.initChooser(fls, "Choose Filter", chooseFilter)
}

func (m *mainModel) OpenPalette() tea.Cmd {
	return m.chooseModel.initChooser(PaletteCommands, "Choose Command", choosePalette)
}

func (m *mainModel) OpenProjects(purpose choosePurpose) tea.Cmd {
	p := m.local.Projects
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
		cm.main.setTasksFromFilter(fts)
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
	switch v.(type) {
	case filter:
		return cm.gotoFilter(v.(filter))
	case paletteCommand:
		return v.(paletteCommand).command(cm.main)
	case project:
		prj := v.(project)
		var cmds []tea.Cmd
		if err == nil {
			switch cm.purpose {
			case chooseProject:
				cm.main.refresh = func() {
					cm.main.setTasksFromProject(&prj)
				}
				cm.main.refresh()
				cm.main.projectId = prj.project.ID
				cm.main.sectionId = prj.section.ID
				cm.main.switchProject(&prj)
			case moveToProject:
				task, err := cm.main.taskList.RemoveCurrentItem()
				if err != nil {
					dbg(err)
				}
				cmds = append(cmds, cm.main.MoveItem(&task.Item, prj))
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
		switch msg.String() {
		case "enter":
			cmds = append(cmds, pm.handleChoose())
			return tea.Batch(cmds...)
		case "esc":
			pm.main.state = viewTasks
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
	m.state = viewChooser
}
