package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type paletteContext uint

const (
	paletteProject paletteContext = iota
	paletteSection
	paletteTask
)

type paletteCommand struct {
	name    string
	command func(*mainModel) tea.Cmd
}

var PaletteCommands = []fmt.Stringer{
	paletteCommand{
		"add project",
		func(m *mainModel) tea.Cmd {
			m.inputModel.GetOnce("", "", func(input string) tea.Cmd {
				return m.AddProject(input)
			})
			return nil
		},
	},
	paletteCommand{
		"archive project",
		func(m *mainModel) tea.Cmd {
			return m.ArchiveProject()
		},
	},
	paletteCommand{
		"rename section",
		func(m *mainModel) tea.Cmd {
			if sct := m.local.SectionMap[m.sectionId]; sct != nil {
				m.inputModel.GetOnce("", "", func(input string) tea.Cmd {
					return m.RenameSection(sct.ID, input)
				})
			}
			return nil
		},
	},
	paletteCommand{
		"archive section",
		func(m *mainModel) tea.Cmd {
			return m.ArchiveSection()
		},
	},
	paletteCommand{
		"add section",
		func(m *mainModel) tea.Cmd {
			m.inputModel.GetOnce("", "", func(input string) tea.Cmd {
				return m.AddSection(input)
			})
			return nil
		},
	},
	paletteCommand{
		"rename project",
		func(m *mainModel) tea.Cmd {
			if prj := m.local.ProjectMap[m.projectId]; prj != nil {
				m.inputModel.GetOnce("", prj.Name, func(input string) tea.Cmd {
					return m.RenameProject(prj.ID, input)
				})
			}

			return nil
		},
	},
}

// move section (to other project, it seems)

// context task
// edit content
// edit desc
// re-prioritise

func (pc paletteCommand) String() string {
	return pc.name
}
