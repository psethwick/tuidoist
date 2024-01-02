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
	for _, i := range m.local.Items {
		// no section tasks should be first
		if i.ProjectID == p.project.ID && i.SectionID == "" {
			tasks = append(tasks, task.New(m.local, i))
		}
	}
	if len(tasks) > 0 {
		lists = append(
			lists,
			tasklist.List{
				Title: p.project.Name, Tasks: tasks, ListId: project{p.project, todoist.Section{}},
			})
	}

	for _, s := range m.local.Sections {
		tasks = []task.Task{}
		if s.ProjectID == p.project.ID {
			for _, item := range m.local.Items {
				if item.ProjectID == p.project.ID && s.ID == item.SectionID {
					tasks = append(tasks, task.New(m.local, item))
				}
			}
			lists = append(lists, tasklist.List{
				Title:  fmt.Sprintf("%s/%s", p.project.Name, s.Name),
				Tasks:  tasks,
				ListId: project{p.project, s},
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

type filterSelection struct {
	lists []filterTitle
	index int
}

func (m *mainModel) setTasksFromFilter(fs filterSelection) {
	// this can't be the right place to do this
	m.projectId = ""
	m.sectionId = ""
	var tls []tasklist.List
	for i, l := range fs.lists {
		tasks := []task.Task{}
		for _, i := range m.local.Items {
			if res, _ := fltr.Eval(l.Expr, &i, m.local); res {
				tasks = append(tasks, task.New(m.local, i))
			}
		}
		tls = append(
			tls,
			tasklist.List{Title: l.Title, Tasks: tasks, ListId: filterSelection{fs.lists, i}},
		)
	}
	m.taskList.ResetItems(tls, fs.index)
}

func (tm *mainModel) OpenUrl(url string) func() tea.Msg {
	return func() tea.Msg {
		// todo mac: open, win: ???
		openCmd := exec.Command("xdg-open", url)
		if err := openCmd.Run(); err != nil {
			dbg(err)
		}
		return nil
	}
}

func (m *mainModel) openInbox() {
	dbg("openInbox", len(m.local.Projects))

	for _, tp := range m.local.Projects {
		if tp.ID == m.local.User.InboxProjectID {
			p := project{tp, todoist.Section{}}
			m.refresh = func() {
				m.setTasksFromProject(&p)
			}
			m.refresh()
			return
		}
	}
	// possible to end up here if InboxProjectID isn't set for whatever reason
	// empty cache, mostly
	m.projectId = "CHANGEME"
	m.refresh = func() {
		// one empty list in the list is better than an empty list of lists
		dbg("empty list refresh")
		if m.local.User.InboxProjectID != "" {
			dbg("attempt fix loading project")
			m.openInbox()
			return
		}
		m.taskList.ResetItems([]tasklist.List{{Title: "Loading..."}}, 0)
	}
	m.refresh()
}
