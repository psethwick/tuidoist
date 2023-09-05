package main

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	filt "github.com/psethwick/tuidoist/filter"
	"github.com/psethwick/tuidoist/todoist"
)

var mdUrlRegex = regexp.MustCompile(`\[([^\]]+)\]\((https?:\/\/[^\)]+)\)`)

type task struct {
	item      todoist.Item
	title     string
	summary   string
	completed bool
	url       string
}

func (t task) String() string {
	s := fmt.Sprintf("%s\n%s", t.title, t.summary)
	if t.completed {
		return strikeThroughStyle.Render(s)
	}
	return s
}

func reformatDate(d string, from string, to string) string {
	// slicing d because _sometimes_ there's timezone info on the date
	// ain't nobody got time for that
	t, err := time.Parse(from, d[:len(from)])
	if err != nil {
		dbg(err)
	}
	return t.Format(to)
}

// heres where reminders is
// todoist.Store.Reminders
// ⏰ ??
// todo overdue should be red somewhere
// today, maybe also highlighted?
func newTask(m *mainModel, item todoist.Item) task {
	indent := strings.Repeat(" ", len(todoist.SearchItemParents(m.client.Store, &item)))
	var checkbox string
	switch item.Priority {
	case 1:
		checkbox = " ⚪ "
	case 2:
		checkbox = " 🔵 "
	case 3:
		checkbox = " 🟡 "
	case 4:
		checkbox = " 🔴 "
	}

	if indent != "" {
		checkbox = fmt.Sprint("╰", checkbox)
	}
	labels := ""
	for _, l := range item.LabelNames {
		labels += fmt.Sprint(" 🏷️ ", l)
	}
	summary := ""
	if item.Due != nil {
		summary += " 🗓️ "
		var fd string
		if strings.Contains(item.Due.Date, "T") {
			fd = reformatDate(item.Due.Date, "2006-01-02T15:04:05", "02 Jan 06 15:04")
		} else {
			fd = reformatDate(item.Due.Date, "2006-01-02", "02 Jan 06")
		}
		if item.Due.IsRecurring {
			fd += " 🔁"
		}
		summary += fd
	}

	content := item.Content
	// todo this only handles one url
	// also doesn't handle bare urls
	urlMatch := mdUrlRegex.FindStringSubmatch(item.Content)
	url := ""
	if len(urlMatch) > 0 {
		content = underlineStyle.Render(strings.Replace(content, urlMatch[0], urlMatch[1], 1))
		content += " 🔗"
		url = urlMatch[2]
	}
	title := fmt.Sprint(indent, checkbox, content, labels)
	summary = fmt.Sprintf("%s\n%s", fmt.Sprint(indent, "   ", summary), fmt.Sprint(indent, item.Description))

	return task{
		item:    item,
		title:   title,
		summary: summary,
		url:     url,
	}
}

func (t task) Title() string {
	if t.completed {
		return strikeThroughStyle.Render(t.title)
	}
	return t.title
}

func (t task) Description() string {
	return t.summary
}

func (t task) FilterValue() string { return t.item.Content }

type ItemSort []list.Item

func (m *mainModel) setTasks(p *project) {
	items := []todoist.Item{}
	for _, i := range m.client.Store.Items {
		if i.ProjectID == p.ID {
			items = append(items, i)
		}
	}
	tasks := []fmt.Stringer{}
	for _, i := range items {
		tasks = append(tasks, newTask(m, i))
	}
	switch listSort {
	case defaultSort:
		m.taskList.List.LessFunc = ChildOrderLess
	case nameSort:
		m.taskList.List.LessFunc = NameLess
	}
	m.statusBarModel.SetTitle(p.Name)
	m.taskList.List.ResetItems(tasks...)
}

func (m *mainModel) setTasksFromFilter(title string, expr filt.Expression) {
	tasks := []fmt.Stringer{}
	projects := m.client.Store.Projects
	for _, i := range m.client.Store.Items {
		if res, _ := filt.Eval(expr, i, projects); res {
			tasks = append(tasks, newTask(m, i))
		}
	}
	m.statusBarModel.SetTitle(title)
	m.taskList.List.ResetItems(tasks...)
}

// todo confirm
func (m *mainModel) deleteTask() func() tea.Msg {
	str, err := m.taskList.RemoveCurrentItem()
	dbg("AFTER", str)
	if err != nil {
		dbg(err)
		return nil
	}
	t := str.(task)
	return func() tea.Msg {
		err := m.client.DeleteItem(m.ctx, []string{t.item.ID})
		if err != nil {
			dbg("del err", err)
		}
		return m.sync()
	}
}

func updateTask(t task) func(fmt.Stringer) (fmt.Stringer, error) {
	return func(fmt.Stringer) (fmt.Stringer, error) {
		return t, nil
	}
}

func (m *mainModel) completeTask() func() tea.Msg {
	idx, err := m.taskList.List.GetCursorIndex()
	if err != nil {
		dbg(err)
		return func() tea.Msg { return nil }
	}
	t, err := m.taskList.List.GetItem(idx)
	tsk := t.(task)
	if err != nil {
		dbg(err)
		return func() tea.Msg { return nil }
	}
	tsk.completed = true
	m.taskList.List.UpdateItem(idx, updateTask(tsk))
	return func() tea.Msg {
		err := m.client.CloseItem(m.ctx, []string{tsk.item.ID})
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
	prj := project(m.client.Store.Projects[0])
	var cmd tea.Cmd
	refresh := func() {
		m.setTasks(&prj)
	}
	m.refresh = refresh
	refresh()
	return cmd
}

/*
Navigate
Open label… G then L

# Add new task to the top of list ⇧ A

z ctrl-z undo
*/
type projectSort uint

const (
	defaultSort projectSort = iota
	prioritySort
	nameSort
	dateSort
	assigneeSort
)

var listSort projectSort = defaultSort
