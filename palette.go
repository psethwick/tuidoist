package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type paletteCommand struct {
	name    string
	command func(*mainModel) tea.Cmd
}

var PaletteCommands = []fmt.Stringer{
	paletteCommand{
		"add project",
		func(m *mainModel) tea.Cmd {
			m.inputModel.GetOnce("add project", "", func(input string) tea.Cmd {
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
			for _, sct := range m.local.Sections {
				if sct.ID == m.sectionId {
					m.inputModel.GetOnce("rename section", "", func(input string) tea.Cmd {
						return m.RenameSection(sct, input)
					})
				}
			}
			return nil
		},
	},
	paletteCommand{
		"archive section",
		func(m *mainModel) tea.Cmd {
			if m.sectionId == "" {
				dbg("no section to archive")
				return nil
			}
			return m.ArchiveSection(m.sectionId)
		},
	},
	paletteCommand{
		"add section",
		func(m *mainModel) tea.Cmd {
			if m.projectId == "" {
				dbg("no project to add section to")
				return nil
			}
			m.inputModel.GetOnce("add section", "", func(input string) tea.Cmd {
				return m.AddSection(input, m.projectId)
			})
			return nil
		},
	},
	paletteCommand{
		"rename project",
		func(m *mainModel) tea.Cmd {
			for _, prj := range m.local.Projects {
				if prj.ID == m.projectId {
					m.inputModel.GetOnce("rename project", prj.Name, func(input string) tea.Cmd {
						return m.RenameProject(prj.ID, input)
					})
				}
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
