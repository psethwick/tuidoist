package main

import (
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"

	todoist "github.com/sachaos/todoist/lib"
	fltr "github.com/sachaos/todoist/lib/filter"

	"github.com/psethwick/tuidoist/task"
	"github.com/psethwick/tuidoist/tasklist"
)

func (m *mainModel) setTasksFromProject(p *project) {

	lists := []tasklist.List{}
	if p.section.ID == "" {
		tasks := []task.Task{}
		for _, i := range m.client.Store.Items {
			if i.ProjectID == p.project.ID {
				tasks = append(tasks, task.New(m.client.Store, i))
			}
		}
		lists = append(lists, tasklist.List{Title: p.Display(), Tasks: tasks})
	} else {
		// TODO multiple lists if there are sections
		// include (no section) if tasks exist there
		tasks := []task.Task{}
		for _, i := range m.client.Store.Items {
			if i.ProjectID == p.project.ID && i.SectionID == p.section.ID {
				tasks = append(tasks, task.New(m.client.Store, i))
			}
		}
		lists = append(lists, tasklist.List{Title: p.Display(), Tasks: tasks})
	}
	m.taskList.ResetItems(lists)
	m.statusBarModel.SetTitle(m.taskList.Title())
	m.statusBarModel.SetNumber(m.taskList.Len())
}

type filterTitle struct {
	Title string
	Expr  fltr.Expression
}

func (m *mainModel) setTasksFromFilter(lists []filterTitle) {
	var tls []tasklist.List
	for _, l := range lists {
		tasks := []task.Task{}
		for _, i := range m.client.Store.Items {
			if res, _ := fltr.Eval(l.Expr, &i, m.client.Store); res {
				tasks = append(tasks, task.New(m.client.Store, i))
			}
		}
		tls = append(tls, tasklist.List{Title: l.Title, Tasks: tasks})
	}
	// m.statusBarModel.SetTitle(title)
	// m.statusBarModel.SetNumber(len(tasks))
	m.taskList.ResetItems(tls)
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
		// TODO
		// err := m.client.UncompleteItem(m.ctx, lastCompletedTask.Item)
		// if err != nil {
		// 	dbg("uncomplete task err", err)
		// }
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
