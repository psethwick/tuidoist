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
			m.inputModel.onAccept = func(input string) tea.Cmd {
				return func() tea.Msg {
					m.client.AddProject(m.ctx, todoist.Project{Name: input})
					return m.sync()
				}
			}
			// todo maybe all the song and dance could be owned by input
			m.inputModel.content.Focus()
			m.state = viewAddProject
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
			prj := m.client.Store.ProjectMap[m.inputModel.projectID]
			if prj == nil {
				dbg("did not find project", m.inputModel.projectID)
				return nil
			}
			m.inputModel.onAccept = func(input string) tea.Cmd {
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
			// initial value, onAccept, maybe title or something params
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
