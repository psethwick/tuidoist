package main

import tea "github.com/charmbracelet/bubbletea"

// not all choices should be available in all contexts

var projectCommands = []string{
	"add project",
	"rename project",
	"archive project",
}

var taskCommands = []string{
	"change due date",
}

var PaletteCommands = append(projectCommands, taskCommands...)

var PaletteMap = map[string]func(*mainModel) tea.Cmd{
	"add project": func(m *mainModel) tea.Cmd {
		dbg("todo adding project")
		return nil
	},
	"rename project": func(m *mainModel) tea.Cmd {
		dbg("todo rename project")
		return nil
	},
	"archive project": func(m *mainModel) tea.Cmd {
		dbg("todo archive project")
		return nil
	},
}

// add project
// rename project
// archive project?

// add section
// rename section
// move section (to other project, it seems)
// archive section

// context task
// edit content
// edit desc
// re-prioritise
// add/remove/change due date
