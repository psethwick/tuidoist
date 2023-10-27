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
			onAccept := func(input string) tea.Cmd {
				return func() tea.Msg {
					return m.AddProject(input)
				}
			}
			m.inputModel.GetOnce("", "", onAccept)
			return nil
		},
	},
	paletteCommand{
		"change due date",
		func(m *mainModel) tea.Cmd {
			dbg("todo")
			return nil
		},
	},
	paletteCommand{
		"rename project",
		func(m *mainModel) tea.Cmd {
			prj := m.local.ProjectMap[m.projectId]
			if prj == nil {
				dbg("did not find project", m.projectId)
				return nil
			}
			onAccept := func(input string) tea.Cmd {
				return m.RenameProject(prj.ID, input)
			}
			m.inputModel.GetOnce("", prj.Name, onAccept)
			return nil
		},
	},
	paletteCommand{
		"archive project",
		func(m *mainModel) tea.Cmd {
			dbg("todo")
			return nil
		},
	},
}

// add section
// rename section
// move section (to other project, it seems)
// archive section

// context task
// edit content
// edit desc
// re-prioritise

func (pc paletteCommand) String() string {
	return pc.name
}
