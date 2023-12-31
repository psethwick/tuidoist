package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/psethwick/tuidoist/task"
	todoist "github.com/sachaos/todoist/lib"
)

// todo this is _not_ a good place/idea
// I think []{actionTaken, Task} ? and treat it like a stack
// long game is complete sync workflow where offline actions can be synced later
// which means serializing this to disk
// var lastCompletedTask task.Task
//
// func (m *mainModel) undoCompleteTask() tea.Cmd {
// 	m.taskList.AddItem(lastCompletedTask)
// 	m.statusBarModel.SetMessage("undo complete", lastCompletedTask.Title)
// 	args := map[string]interface{}{"id": lastCompletedTask.Item.ID}
// 	// todo undoop
// 	return m.sync(todoist.NewCommand("item_uncomplete", args))
// }

// todo confirm
func (m *mainModel) deleteTasks() tea.Cmd {
	return m.bulkOps("deleted", func(t task.Task) todoist.Command {
		return todoist.NewCommand("item_delete", map[string]interface{}{"id": t.Item.ID})
	})
}

func (m *mainModel) completeTasks() tea.Cmd {
	return m.bulkOps("completed", func(t task.Task) todoist.Command {
		return todoist.NewCommand("item_close", map[string]interface{}{"id": t.Item.ID})
	})
}

func (m *mainModel) rescheduleTasks(newDate string) tea.Cmd {
	return m.bulkOps("rescheduled", func(t task.Task) todoist.Command {
		t.Item.Due = &todoist.Due{
			String: newDate,
		}
		return todoist.NewCommand("item_update", t.Item.UpdateParam())
	})
}

func (m *mainModel) MoveItems(p project) tea.Cmd {
	return m.bulkOps("moved", func(t task.Task) todoist.Command {
		args := map[string]interface{}{"id": t.Item.ID}
		if p.section.ID != "" {
			args["section_id"] = p.section.ID
		} else {
			args["project_id"] = p.project.ID
		}
		return todoist.NewCommand("item_move", args)
	})
}

func (m *mainModel) bulkOps(name string, builder func(task.Task) todoist.Command) tea.Cmd {
	var cmds []todoist.Command
	for _, t := range m.taskList.SelectedItems() {
		m.statusBarModel.SetMessage(name, t.Title)
		cmds = append(cmds, builder(t))
	}
	cmdLen := len(cmds)
	if cmdLen > 1 {
		m.statusBarModel.SetMessage(name, fmt.Sprint(cmdLen, "tasks"))
	}
	return m.sync(cmds...)
}

func (m *mainModel) addTask(content string) tea.Cmd {
	if content == "" {
		return func() tea.Msg { return nil }
	}
	item := todoist.Item{}
	item.Content = content
	item.Priority = 1

	item.ProjectID = m.projectId
	item.SectionID = m.sectionId

	m.statusBarModel.SetMessage("added", item.Content)
	args := map[string]interface{}{}
	if item.Content != "" {
		args["content"] = item.Content
	}
	if item.SectionID != "" {
		args["section_id"] = item.SectionID
	}
	if item.Description != "" {
		args["description"] = item.Description
	}
	if item.DateString != "" {
		args["date_string"] = item.DateString
	}
	if len(item.LabelNames) != 0 {
		args["labels"] = item.LabelNames
	}
	if item.Priority != 0 {
		args["priority"] = item.Priority
	}
	if item.ProjectID != "" {
		args["project_id"] = item.ProjectID
	}
	if item.Due != nil {
		args["due"] = item.Due
	}
	args["auto_reminder"] = item.AutoReminder

	return m.sync(todoist.NewCommand("item_add", args))
}

func (m *mainModel) AddProject(name string) tea.Cmd {
	param := map[string]interface{}{
		"name": name,
	}
	return m.sync(todoist.NewCommand("project_add", param))
}

func (m *mainModel) RenameProject(projectId string, newName string) tea.Cmd {
	args := map[string]interface{}{
		"id":   projectId,
		"name": newName,
	}
	return m.sync(todoist.NewCommand("project_update", args))
}

func (m *mainModel) UpdateItem(i todoist.Item) tea.Cmd {
	return m.sync(todoist.NewCommand("item_update", i.UpdateParam()))
}

func (m *mainModel) RenameSection(ID string, newName string) tea.Cmd {
	args := map[string]interface{}{
		"id":   ID,
		"name": newName,
	}
	return m.sync(todoist.NewCommand("section_update", args))
}

func (m *mainModel) AddSection(name string) tea.Cmd {
	// todo
	return nil
}

func (m *mainModel) ArchiveSection() tea.Cmd {
	// todo
	return nil
}

func (m *mainModel) ArchiveProject() tea.Cmd {
	// todo
	return nil
}
