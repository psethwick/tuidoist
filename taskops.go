package main

import (
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	fltr "github.com/psethwick/tuidoist/filter"
	"github.com/psethwick/tuidoist/task"
	"github.com/psethwick/tuidoist/todoist"
)

func (m *mainModel) setTasksFromProject(p *project) {

	tasks := []task.Task{}
	if p.section.ID == "" {
		for _, i := range m.client.Store.Items {
			if i.ProjectID == p.project.ID {
				tasks = append(tasks, task.New(m.client.Store, i))
			}
		}
	} else { // project / section
		for _, i := range m.client.Store.Items {
			if i.ProjectID == p.project.ID && i.SectionID == p.section.ID {
				tasks = append(tasks, task.New(m.client.Store, i))
			}
		}
	}
	m.statusBarModel.SetTitle(p.Display())
	m.statusBarModel.SetNumber(len(tasks))
	m.taskList.ResetItems(tasks)
}

func (m *mainModel) setTasksFromFilter(title string, expr fltr.Expression) {
	tasks := []task.Task{}
	projects := m.client.Store.Projects
	for _, i := range m.client.Store.Items {
		if res, _ := fltr.Eval(expr, i, projects); res {
			tasks = append(tasks, task.New(m.client.Store, i))
		}
	}
	m.statusBarModel.SetTitle(title)
	m.statusBarModel.SetNumber(len(tasks))
	m.taskList.ResetItems(tasks)
}

// todo confirm
func (m *mainModel) deleteTask() func() tea.Msg {
	t, err := m.taskList.RemoveCurrentItem()
	if err != nil {
		dbg(err)
		return nil
	}
	m.statusBarModel.SetMessage("deleted", t.Title)
	return func() tea.Msg {
		err := m.client.DeleteItem(m.ctx, []string{t.Item.ID})
		if err != nil {
			dbg("del err", err)
		}
		return m.sync()
	}
}

// todo this is _not_ a good place/idea
// I think []{actionTaken, Task} ? and treat it like a stack
// long game is complete sync workflow where offline actions can be synced later
// which means serializing this to disk
var lastCompletedTask task.Task

func (m *mainModel) undoCompleteTask() func() tea.Msg {
	m.taskList.AddItem(lastCompletedTask)
	m.statusBarModel.SetMessage("undo complete", lastCompletedTask.Title)
	return func() tea.Msg {
		err := m.client.UncompleteItem(m.ctx, lastCompletedTask.Item)
		if err != nil {
			dbg("uncomplete task err", err)
		}
		return m.sync()
	}
}

func (m *mainModel) completeTask() func() tea.Msg {
	t, err := m.taskList.GetCursorItem()
	if err != nil {
		dbg(err)
		return func() tea.Msg { return nil }
	}
	lastCompletedTask = task.Task(t)
	t.Completed = true
	m.statusBarModel.SetMessage("completed", t.Title)
	m.taskList.UpdateCurrentTask(t)
	return func() tea.Msg {
		err := m.client.CloseItem(m.ctx, []string{t.Item.ID})
		if err != nil {
			dbg("complete task err", err)
		}

		return m.sync()
	}
}

func (tm *mainModel) OpenUrl(url string) func() tea.Msg {
	return func() tea.Msg {
		// todo mac: open, win: ???
		openCmd := exec.Command("xdg-open", url)
		openCmd.Run()
		return nil
	}
}

func (m *mainModel) openInbox() tea.Cmd {
	if len(m.client.Store.Projects) == 0 {
		return nil
	}
	prj := project{m.client.Store.Projects[0], todoist.Section{}}
	var cmd tea.Cmd
	m.refresh = func() {
		m.setTasksFromProject(&prj)
	}
	m.refresh()
	return cmd
}
