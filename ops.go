package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/psethwick/tuidoist/task"
	todoist "github.com/sachaos/todoist/lib"
)

// todo this is _not_ a good place/idea
// I think []{actionTaken, Task} ? and treat it like a stack
// long game is complete sync workflow where offline actions can be synced later
// which means serializing this to disk
var lastCompletedTask task.Task

func (m *mainModel) undoCompleteTask() tea.Cmd {
	m.taskList.AddItem(lastCompletedTask)
	m.statusBarModel.SetMessage("undo complete", lastCompletedTask.Title)
	args := map[string]interface{}{"id": lastCompletedTask.Item.ID}
	// todo undoop
	return m.DoOp(todoist.NewCommand("item_uncomplete", args))
}

// todo confirm
func (m *mainModel) deleteTask() tea.Cmd {
	t, err := m.taskList.RemoveCurrentItem()
	if err != nil {
		dbg(err)
		return nil
	}
	m.statusBarModel.SetMessage("deleted", t.Title)
	return m.DoOp(
		todoist.NewCommand("item_delete", map[string]interface{}{"id": t.Item.ID}),
	)
}
func (m *mainModel) addTask(content string) tea.Cmd {
	if content == "" {
		return func() tea.Msg { return nil }
	}
	i := todoist.Item{}
	i.Content = content
	i.Priority = 1

	i.ProjectID = m.projectId
	i.SectionID = m.sectionId

	t := task.New(m.store, i)
	m.statusBarModel.SetMessage("added", t.Title)
	t = m.taskList.AddItemBottom(t)
	m.state = viewTasks
	item := t.Item
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

	return m.DoOp(todoist.NewCommand("item_add", args))
}

func (m *mainModel) completeTask() tea.Cmd {
	t, err := m.taskList.GetCursorItem()
	if err != nil {
		dbg(err)
		return func() tea.Msg { return nil }
	}
	lastCompletedTask = task.Task(t)
	t.Completed = true
	m.statusBarModel.SetMessage("completed", t.Title)
	m.taskList.RemoveCurrentItem()
	return m.DoOp(todoist.NewCommand("item_close", map[string]interface{}{"id": t.Item.ID}))
}

func (m *mainModel) MoveItem(item *todoist.Item, p project) func() tea.Msg {
	args := map[string]interface{}{"id": item.ID}
	if p.section.ID != "" {
		args["section_id"] = p.section.ID
	} else {
		args["project_id"] = p.project.ID
	}
	return m.DoOp(todoist.NewCommand("item_move", args))
}

func (m *mainModel) AddProject(name string) tea.Cmd {
	param := map[string]interface{}{
		"name": name,
	}
	return m.DoOp(todoist.NewCommand("project_add", param))
}

func (m *mainModel) RenameProject(projectId string, newName string) tea.Cmd {
	args := map[string]interface{}{
		"id":   projectId,
		"name": newName,
	}
	m.state = viewTasks

	return m.DoOp(todoist.NewCommand("project_update", args))
}

func (m *mainModel) UpdateItem(i todoist.Item) func() tea.Msg {
	// todo use DoOp
	return func() tea.Msg {
		m.client.UpdateItem(m.ctx, i)
		return m.sync()
	}
}

func (m *mainModel) DoOp(cmd todoist.Command) tea.Cmd {
	return m.DoOps(todoist.Commands{cmd})
}

func (m *mainModel) DoOps(cmds todoist.Commands) tea.Cmd {
	*m.opQueue = append(*m.opQueue, cmds...)
	return func() tea.Msg {
		err := m.client.ExecCommands(m.ctx, cmds)
		if err == nil {
			m.opQueue = &todoist.Commands{}
		} else {
			dbg(err)
		}
		return m.sync()
	}
}
