package main

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	todoist "github.com/sachaos/todoist/lib"

	"github.com/psethwick/tuidoist/style"
	"github.com/psethwick/tuidoist/task"
)

type inputPurpose uint

const (
	inputAddTask inputPurpose = iota
	inputAddProject
	inputEditProject
)

type inputModel struct {
	content   textinput.Model
	main      *mainModel
	purpose   inputPurpose
	projectID string
	sectionId string
}

func newInputModel(m *mainModel) inputModel {
	ti := textinput.New()
	ti.Prompt = "   > "
	return inputModel{
		content: ti,
		main:    m,
	}
}

func (im *inputModel) Height() int {
	return 4 // height of textinput + dialog
}

func (im *inputModel) addTask() tea.Cmd {
	content := im.content.Value()
	im.content.SetValue("")
	if content == "" {
		return func() tea.Msg { return nil }
	}
	i := todoist.Item{}
	i.Content = content
	i.Priority = 1

	i.ProjectID = im.projectID
	i.SectionID = im.sectionId

	t := task.New(im.main.client.Store, i)
	im.main.statusBarModel.SetMessage("added", t.Title)
	if im.main.state == viewNewTaskTop {
		t = im.main.taskList.AddItemTop(t)
	} else {
		t = im.main.taskList.AddItemBottom(t)
	}
	return func() tea.Msg {
		item := t.Item
		param := map[string]interface{}{}
		if item.Content != "" {
			param["content"] = item.Content
		}
		if item.SectionID != "" {
			param["section_id"] = item.SectionID
		}
		if item.Description != "" {
			param["description"] = item.Description
		}
		if item.DateString != "" {
			param["date_string"] = item.DateString
		}
		if len(item.LabelNames) != 0 {
			param["labels"] = item.LabelNames
		}
		if item.Priority != 0 {
			param["priority"] = item.Priority
		}
		if item.ProjectID != "" {
			param["project_id"] = item.ProjectID
		}
		if item.Due != nil {
			param["due"] = item.Due
		}
		param["auto_reminder"] = item.AutoReminder

		im.main.client.ExecCommands(im.main.ctx,
			todoist.Commands{
				todoist.NewCommand("item_add", param),
			},
		)
		return im.main.sync()
	}
}

func (im *inputModel) addProject() tea.Cmd {
	return func() tea.Msg {
		im.main.client.AddProject(im.main.ctx, todoist.Project{Name: im.content.Value()})
		return im.main.sync()
	}
}

func (im *inputModel) editProject() tea.Cmd {
	param := map[string]interface{}{}
	param["id"] = im.projectID
	param["name"] = im.content.Value()
	return func() tea.Msg {
		im.main.client.ExecCommands(im.main.ctx,
			todoist.Commands{
				todoist.NewCommand("project_update", param),
			},
		)
		return im.main.sync()
	}
}

// Save a newly created task and create a new one below Enter
// Save changes to an existing task and create a new task below ⇧ Enter
// Save a new task or save changes to an existing one and create a new task above Ctrl Enter
func (im *inputModel) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			switch im.purpose {
			case inputEditProject:
				cmds = append(cmds, im.editProject())
				im.main.state = viewTasks
			case inputAddProject:
				cmds = append(cmds, im.addProject())
				im.main.state = viewTasks
			case inputAddTask:
				cmds = append(cmds, im.addTask())
			}
		case "esc":
			im.content.SetValue("")
			im.main.taskList.SetHeight(im.main.height - 1)
			im.content.Blur()
			im.projectID = ""
			im.sectionId = ""
			im.main.state = viewTasks
		}
	}
	input, cmd := im.content.Update(msg)
	im.content = input
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

func (im *inputModel) View() string {
	return lipgloss.Place(im.main.width, 3,
		lipgloss.Left, lipgloss.Left,
		style.DialogBox.Render(im.content.View()),
	)
}
