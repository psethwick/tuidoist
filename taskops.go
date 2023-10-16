package main

import (
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"

	todoist "github.com/sachaos/todoist/lib"
	fltr "github.com/sachaos/todoist/lib/filter"

	"github.com/psethwick/tuidoist/task"
	"github.com/psethwick/tuidoist/tasklist"
)

func (m *mainModel) setTasksFromProject(p *project) {
	lists := []tasklist.List{}
	tasks := []task.Task{}
	var selectedList int
	for _, i := range m.store.Items {
		// no section tasks should be first
		if i.ProjectID == p.project.ID && i.SectionID == "" {
			tasks = append(tasks, task.New(m.store, i))
		}
	}
	if len(tasks) > 0 {
		lists = append(lists, tasklist.List{Title: p.project.Name, Tasks: tasks})
	}

	for _, s := range m.store.Sections {
		tasks = []task.Task{}
		if s.ProjectID == p.project.ID {
			for _, item := range m.store.Items {
				if item.ProjectID == p.project.ID && s.ID == item.SectionID {
					tasks = append(tasks, task.New(m.store, item))
				}
			}
			lists = append(lists, tasklist.List{
				Title: fmt.Sprintf("%s/%s", p.project.Name, s.Name),
				Tasks: tasks,
			})
			if s.ID == p.section.ID {
				selectedList = len(lists) - 1
			}
		}
	}
	if len(lists) == 0 {
		lists = append(lists, tasklist.List{Title: p.project.Name})
	}
	m.taskList.ResetItems(lists, selectedList)
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
		for _, i := range m.store.Items {
			if res, _ := fltr.Eval(l.Expr, &i, m.store); res {
				tasks = append(tasks, task.New(m.store, i))
			}
		}
		tls = append(tls, tasklist.List{Title: l.Title, Tasks: tasks})
	}
	// m.statusBarModel.SetTitle()
	// m.statusBarModel.SetNumber(len(tasks))
	m.taskList.ResetItems(tls, 0)
}

func (tm *mainModel) OpenUrl(url string) func() tea.Msg {
	return func() tea.Msg {
		// todo mac: open, win: ???
		openCmd := exec.Command("xdg-open", url)
		openCmd.Run()
		return nil
	}
}

func (m *mainModel) openInbox() {
	for _, tp := range m.store.Projects {
		if tp.ID == m.store.User.InboxProjectID {
			p := project{tp, todoist.Section{}}
			m.refresh = func() {
				m.setTasksFromProject(&p)
			}
			break
		}
	}
	m.refresh()
}
