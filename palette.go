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

var noContextCommands = []fmt.Stringer{
	paletteCommand{
		"add project",
		func(m *mainModel) tea.Cmd {
			m.inputModel.purpose = inputAddProject
			m.inputModel.content.Focus()
			m.state = viewAddProject
			return nil
		},
	},
}

// add section
// rename section
// move section (to other project, it seems)
// archive section
var projectCommands = []fmt.Stringer{
	paletteCommand{
		"rename project",
		func(m *mainModel) tea.Cmd {
			prj := m.client.Store.ProjectMap[m.inputModel.projectID]
			if prj == nil {
				dbg("did not find project", m.inputModel.projectID)
				return nil
			}
			m.inputModel.purpose = inputEditProject
			m.inputModel.content.SetValue(prj.Name)
			m.inputModel.content.Focus()
			m.state = viewAddProject
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

// context task
// edit content
// edit desc
// re-prioritise
var taskCommands = []fmt.Stringer{
	paletteCommand{
		"change due date",
		func(m *mainModel) tea.Cmd {
			dbg("todo")
			return nil
		},
	},
}

func (pc paletteCommand) String() string {
	return pc.name
}

func PaletteCommands(contexts ...paletteContext) []fmt.Stringer {
	commands := noContextCommands
	for _, ctx := range contexts {
		switch ctx {
		case paletteProject:
			commands = append(commands, projectCommands...)
		case paletteTask:
			commands = append(commands, taskCommands...)
		}
	}
	return commands
}
