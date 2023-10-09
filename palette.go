package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	todoist "github.com/sachaos/todoist/lib"
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
					m.client.AddProject(m.ctx, todoist.Project{Name: input})
					return m.sync()
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
			// todo maybe input model is not the right place for task/projetc 'context'
			prj := m.client.Store.ProjectMap[m.projectId]
			if prj == nil {
				dbg("did not find project", m.projectId)
				return nil
			}
			onAccept := func(input string) tea.Cmd {
				param := map[string]interface{}{}
				param["id"] = prj.ID
				param["name"] = input
				m.state = viewTasks
				return func() tea.Msg {
					m.client.ExecCommands(m.ctx,
						todoist.Commands{
							todoist.NewCommand("project_update", param),
						},
					)
					return m.sync()
				}
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
