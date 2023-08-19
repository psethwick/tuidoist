package main

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sachaos/todoist/lib"
    filt "github.com/psethwick/tuidoist/filter"
)

type tasksModel struct {
	tasks      list.Model
	main       *mainModel
	mdUrlRegex *regexp.Regexp
}

var underlineStyle = lipgloss.NewStyle().Underline(false)

func newTasksModel(m *mainModel) tasksModel {
	tasks := list.New([]list.Item{}, list.NewDefaultDelegate(), 40, 30)
	re := regexp.MustCompile(`\[([^\]]+)\]\((https?:\/\/[^\)]+)\)`)

	tasks.Styles.TitleBar = lipgloss.NewStyle().Padding(0, 0, 1, 2)

	tasks.Styles.Title = lipgloss.NewStyle().
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("230")).
		Padding(0, 1)

	tasks.DisableQuitKeybindings()

	return tasksModel{
		tasks,
		m,
		re,
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
		// subtask indicator
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
	// guess I could just loop until there are none?
	// also doesn't handle bare urls
	urlMatch := m.tasksModel.mdUrlRegex.FindStringSubmatch(item.Content)
	url := ""
	if len(urlMatch) > 0 {
		content = underlineStyle.Render(strings.Replace(content, urlMatch[0], urlMatch[1], 1))
		content += "🔗"
		url = urlMatch[2]
	}
	title := fmt.Sprint(indent, checkbox, content, labels)
	summary = fmt.Sprint(indent, "   ", summary)
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

func (m *mainModel) setTasks(p *todoist.Project) {
	tasks := []list.Item{}
	for _, i := range m.client.Store.Items {
		if i.ProjectID == p.ID {
			tasks = append(tasks, newTask(m, i))
		}
	}
	m.tasksModel.tasks.SetItems(tasks)
}

func (m *mainModel) setTasksFromFilter(expr filt.Expression) {
	tasks := []list.Item{}
    projects := m.client.Store.Projects
	for _, i := range m.client.Store.Items {
        if res, _ := filt.Eval(expr, i, projects); res {
			tasks = append(tasks, newTask(m, i))
		}
	}
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
	tm.tasks.SetHeight(tm.main.size.Height - fh - h)
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
		openCmd := exec.Command("xdg-open", url)
		openCmd.Run()
		return nil
	}
}

func (tm *tasksModel) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "F":
            tm.main.OpenFilters()
		case "p":
            tm.main.OpenProjects(chooseProject, "asdfasdf", "Switch Project")
		case "m":
            tm.main.OpenProjects(moveToProject, "asdfasdf", "Move to Project")
		case "v":
			t := tm.tasks.SelectedItem().(task)
			if t.url != "" {
				cmds = append(cmds, tm.OpenUrl(t.url))
			}
		case "c":
			cmds = append(cmds, tm.main.completeTask())
		case "d":
			cmds = append(cmds, tm.main.deleteTask())
		case "a":
			tm.GiveHeight(tm.main.newTaskModel.Height())
			tm.main.newTaskModel.content.Prompt = "> "
			tm.main.newTaskModel.content.Focus()
			tm.main.state = newTaskState
		}
	}
	tasks, cmd := tm.tasks.Update(msg)
	tm.tasks = tasks
	cmds = append(cmds, cmd)
	return tea.Batch(cmds...)
}
