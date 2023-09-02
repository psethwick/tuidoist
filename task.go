package main

import (
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	filt "github.com/psethwick/tuidoist/filter"
	"github.com/psethwick/tuidoist/todoist"
)

type tasksModel struct {
	title   string
	tasks   list.Model
	main    *mainModel
	refresh func()
	gMenu   bool
}

var mdUrlRegex = regexp.MustCompile(`\[([^\]]+)\]\((https?:\/\/[^\)]+)\)`)

func newTasksModel(m *mainModel) tasksModel {
	tasks := list.New([]list.Item{}, taskDelegate{}, 40, 30)
	tasks.SetShowTitle(false)
	tasks.SetShowHelp(false)
	tasks.SetShowPagination(false)
	tasks.SetShowStatusBar(false)
	tasks.DisableQuitKeybindings()
	refresh := func() {
		ts := []list.Item{}
		if len(m.client.Store.Projects) > 0 {
			p := m.client.Store.Projects[0]
			for _, i := range m.client.Store.Items {
				if i.ProjectID == p.ID {
					ts = append(ts, newTask(m, i))
				}
			}
			switch listSort {
			case defaultSort:
				sort.Sort(SortByChildOrder(ts))
			case nameSort:
				sort.Sort(SortByName(ts))
			}
			m.tasksModel.tasks.SetItems(ts)
		}
	}

	return tasksModel{
		tasks:   tasks,
		main:    m,
		refresh: refresh,
		title:   "Inbox",
	}
}

type task struct {
	item      todoist.Item
	title     string
	summary   string
	completed bool
	url       string
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

type SortByChildOrder ItemSort

func (a SortByChildOrder) Len() int      { return len(a) }
func (a SortByChildOrder) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a SortByChildOrder) Less(i, j int) bool {
	return a[i].(task).item.ChildOrder < a[j].(task).item.ChildOrder
}

type SortByName ItemSort

func (a SortByName) Len() int           { return len(a) }
func (a SortByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SortByName) Less(i, j int) bool { return a[i].(task).item.Content < a[j].(task).item.Content }

func (m *mainModel) setTasks(p *project) {
	items := []todoist.Item{}
	for _, i := range m.client.Store.Items {
		if i.ProjectID == p.ID {
			items = append(items, i)
		}
	}
	tasks := []list.Item{}
	for _, i := range items {
		tasks = append(tasks, newTask(m, i))
	}
	switch listSort {
	case defaultSort:
		sort.Sort(SortByChildOrder(tasks))
	case nameSort:
		sort.Sort(SortByName(tasks))
	}
	m.statusBarModel.SetTitle(p.Name)
	m.tasksModel.tasks.SetItems(tasks)
}

func (m *mainModel) setTasksFromFilter(title string, expr filt.Expression) {
	tasks := []list.Item{}
	projects := m.client.Store.Projects
	for _, i := range m.client.Store.Items {
		if res, _ := filt.Eval(expr, i, projects); res {
			tasks = append(tasks, newTask(m, i))
		}
	}
	m.statusBarModel.SetTitle(title)
	m.tasksModel.tasks.SetItems(tasks)
}

// todo confirm
func (m *mainModel) deleteTask() func() tea.Msg {
	t := m.tasksModel.tasks.SelectedItem().(task)
	m.tasksModel.tasks.RemoveItem(m.tasksModel.tasks.Index())
	return func() tea.Msg {
		err := m.client.DeleteItem(m.ctx, []string{t.item.ID})
		if err != nil {
			dbg("del err", err)
		}
		return m.sync()
	}
}

func (m *mainModel) completeTask() func() tea.Msg {
	t := m.tasksModel.tasks.SelectedItem().(task)
	t.completed = true
	m.tasksModel.tasks.SetItem(m.tasksModel.tasks.Index(), t)
	return func() tea.Msg {
		err := m.client.CloseItem(m.ctx, []string{t.item.ID})
		if err != nil {
			dbg("complete task err", err)
		}
		return m.sync()
	}
}

func (tm *tasksModel) GiveHeight(h int) {
	fh, _ := listStyle.GetFrameSize()
	tm.tasks.SetHeight(tm.main.height - fh - h)
}

func (tm *tasksModel) Focus() {
	tm.main.state = tasksState
	tm.GiveHeight(0)
}

func (tm *tasksModel) View() string {
	return listStyle.Render(tm.tasks.View())
}

func (tm *tasksModel) OpenUrl(url string) func() tea.Msg {
	return func() tea.Msg {
		// todo mac: open, win: ???
		openCmd := exec.Command("xdg-open", url)
		openCmd.Run()
		return nil
	}
}

func (tm *tasksModel) openInbox() tea.Cmd {
	if len(tm.main.client.Store.Projects) == 0 {
		return nil
	}
	prj := project(tm.main.client.Store.Projects[0])
	var cmd tea.Cmd
	refresh := func() {
		tm.main.setTasks(&prj)
	}
	tm.main.tasksModel.refresh = refresh
	refresh()
	tm.main.tasksModel.tasks.FilterInput.SetValue("")
	tm.main.tasksModel.title = "Inbox"
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

func (tm *tasksModel) setSort(ps projectSort) {
	if ps == listSort { // toggle off
		listSort = defaultSort
	} else {
		listSort = ps
	}
	tm.refresh()
}

func (tm *tasksModel) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	if tm.tasks.FilterState() != list.Filtering {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "v":
				t := tm.main.tasksModel.tasks.SelectedItem().(task)
				if t.url != "" {
					cmds = append(cmds, tm.main.tasksModel.OpenUrl(t.url))
				}
			case "g":
				tm.gMenu = true
				// case "u":
				// 	if tm.gMenu {
				// todo how to do upcoming
				// 	}
			case "c":
				cmds = append(cmds, tm.main.completeTask())
				tm.main.state = tasksState
			case "delete":
				// TODO confirmation
				cmds = append(cmds, tm.main.deleteTask())
			case "f":
				if tm.gMenu {
					cmds = append(cmds, tm.main.OpenFilters())
					tm.gMenu = false
				}
			case "i":
				if tm.gMenu {
					cmds = append(cmds, tm.openInbox())
					tm.gMenu = false
				}
			case "t":
				if tm.gMenu {
					cmds = append(cmds, tm.main.chooseModel.gotoFilter(filter{Name: "Today", Query: "today"}))
					tm.gMenu = false
				}
			case "p":
				if tm.gMenu {
					cmds = append(cmds, tm.main.OpenProjects(chooseProject))
					tm.gMenu = false
				} else {
					tm.setSort(prioritySort)
				}
			case "n":
				tm.setSort(nameSort)
			case "d":
				tm.setSort(dateSort)
			case "r":
				tm.setSort(assigneeSort)
			case "m":
				cmds = append(cmds, tm.main.OpenProjects(moveToProject))
			case "enter":
				t := tm.tasks.SelectedItem().(task)
				tm.main.taskMenuModel.project = tm.main.client.Store.FindProject(t.item.ProjectID)
				tm.main.taskMenuModel.item = t.item
				tm.main.taskMenuModel.content.SetValue(t.item.Content)
				tm.main.taskMenuModel.desc.SetValue(t.item.Description)
				tm.main.state = taskMenuState
			case "a":
				tm.GiveHeight(tm.main.newTaskModel.Height())
				tm.main.newTaskModel.content.Focus()
				tm.main.state = newTaskBottomState
			case "A":
				tm.GiveHeight(tm.main.newTaskModel.Height())
				tm.main.newTaskModel.content.Focus()
				tm.main.state = newTaskTopState
			default:
				tm.gMenu = false
			}
		}
	}
	tasks, cmd := tm.tasks.Update(msg)
	tm.tasks = tasks
	cmds = append(cmds, cmd)
	return tea.Batch(cmds...)
}
