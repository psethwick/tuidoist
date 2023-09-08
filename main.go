package main

import (
	"context"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/psethwick/tuidoist/todoist"

	"github.com/psethwick/tuidoist/client"
	"github.com/psethwick/tuidoist/status"
	"github.com/psethwick/tuidoist/tasklist"
)

type viewState uint

const (
	tasksState viewState = iota
	chooseState
	newTaskTopState
	newTaskBottomState
	taskMenuState
)

type mainModel struct {
	client         *todoist.Client
	height         int
	width          int
	state          viewState
	ctx            context.Context
	chooseModel    chooseModel
	taskList       tasklist.TaskList
	newTaskModel   newTaskModel
	taskMenuModel  taskMenuModel
	statusBarModel status.Model
	refresh        func()
	gMenu          bool
}

func initialModel() *mainModel {
	m := mainModel{}
	m.client = client.GetClient(dbg)
	m.ctx = context.Background()
	m.taskList = tasklist.New(ChildOrderLess, dbg)
	m.chooseModel = newChooseModel(&m)
	m.newTaskModel = newNewTaskModel(&m)
	m.taskMenuModel = newTaskMenuModel(&m)
	m.statusBarModel = status.New()
	return &m
}

func ChildOrderLess(a fmt.Stringer, b fmt.Stringer) bool {
	return a.(Task).Item.ChildOrder < b.(Task).Item.ChildOrder
}

func NameLess(a fmt.Stringer, b fmt.Stringer) bool {
	return a.(Task).Item.Content < b.(Task).Item.Content
}

func (m *mainModel) refreshFromStore() tea.Cmd {
	m.refresh()
	return nil
}

func (m *mainModel) sync() tea.Msg {
	m.statusBarModel.SetSyncStatus(status.Syncing)
	err := m.client.Sync(m.ctx)
	if err != nil {
		dbg("Synced", err)
		m.statusBarModel.SetSyncStatus(status.Error)
		return nil
	}
	err = client.WriteCache(m.client.Store)
	if err != nil {
		dbg(err)
		m.statusBarModel.SetSyncStatus(status.Error)
		return nil
	}
	m.refreshFromStore()
	m.statusBarModel.SetSyncStatus(status.Synced)

	return nil
}

func (m *mainModel) Init() tea.Cmd {
	m.refresh = func() {
		if len(m.client.Store.Projects) > 0 {
			p := project{m.client.Store.Projects[0], todoist.Section{}}
			m.setTasksFromProject(&p)
		}
	}
	m.refreshFromStore()
	return m.sync
}

func (m *mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height - 1 // statusbar
		m.width = msg.Width
		// for children to get the size they can actually have
		m.taskList.SetHeight(msg.Height - 1)
		m.taskList.SetWidth(msg.Width)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+w":
			fallthrough
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	switch m.state {
	case chooseState:
		cmds = append(cmds, m.chooseModel.Update(msg))
	case tasksState:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "j":
				m.taskList.MoveCursor(1)
			case "k":
				m.taskList.MoveCursor(-1)
			case "v":
				str, err := m.taskList.GetCursorItem()
				if err != nil {
					dbg(err)
				}
				t, ok := str.(Task)
				if ok {
					if t.Url != "" {
						cmds = append(cmds, m.OpenUrl(t.Url))
					}

				}
			case "G":
				m.taskList.Bottom()
			case "g":
				if m.gMenu {
					m.taskList.Top()
					m.gMenu = false

				} else {
					m.gMenu = true
				}
				// case "u":
				// 	if tm.gMenu {
				// todo how to do upcoming
				// 	}
			case "c":
				cmds = append(cmds, m.completeTask())
				m.state = tasksState
			case "delete":
				// TODO confirmation
				cmds = append(cmds, m.deleteTask())
			case "f":
				if m.gMenu {
					cmds = append(cmds, m.OpenFilters())
					m.gMenu = false
				}
			case "i":
				if m.gMenu {
					// TODO
					cmds = append(cmds, m.openInbox())
					m.gMenu = false
				}
			case "t":
				if m.gMenu {
					cmds = append(cmds, m.chooseModel.gotoFilter(filter{Name: "Today", Query: "today | overdue"}))
					m.gMenu = false
				}
			case "p":
				if m.gMenu {
					cmds = append(cmds, m.OpenProjects(chooseProject))
					m.gMenu = false
				} else {
					// TODO
					// tm.setSort(prioritySort)
				}
			case "n":
				// TODO
				//tm.setSort(nameSort)
			case "d":
				// tm.setSort(dateSort)
			case "r":
				// tm.setSort(assigneeSort)
			case "m":
				cmds = append(cmds, m.OpenProjects(moveToProject))
			case "enter":
				// t := tm.tasks.SelectedItem().(task)
				// tm.main.taskMenuModel.project = tm.main.client.Store.FindProject(t.item.ProjectID)
				// tm.main.taskMenuModel.item = t.item
				// tm.main.taskMenuModel.content.SetValue(t.item.Content)
				// tm.main.taskMenuModel.desc.SetValue(t.item.Description)
				// tm.main.state = taskMenuState
			case "a":
				m.taskList.SetHeight(m.height - m.newTaskModel.Height())
				m.taskList.Bottom()
				m.newTaskModel.content.Focus()
				m.state = newTaskBottomState
			case "A":
				m.taskList.SetHeight(m.height - m.newTaskModel.Height())
				m.taskList.Top()
				m.newTaskModel.content.Focus()
				m.state = newTaskTopState
			default:
				m.gMenu = false
			}
		}
	case newTaskTopState:
		fallthrough
	case newTaskBottomState:
		cmds = append(cmds, m.newTaskModel.Update(msg))
	case taskMenuState:
		cmds = append(cmds, m.taskMenuModel.Update(msg))
	}
	cmds = append(cmds, m.statusBarModel.Update(msg))
	return m, tea.Batch(cmds...)
}

func (m *mainModel) View() string {
	var s string
	switch m.state {
	case chooseState:
		s = m.chooseModel.View()
	case tasksState:
		s = m.taskList.View()
	case taskMenuState:
		s = m.taskMenuModel.View()
	case newTaskBottomState:
		s = lipgloss.JoinVertical(
			lipgloss.Left,
			m.taskList.View(),
			m.newTaskModel.View(),
		)
	case newTaskTopState:
		s = lipgloss.JoinVertical(
			lipgloss.Left,
			m.newTaskModel.View(),
			m.taskList.View(),
		)
	}
	return lipgloss.JoinVertical(lipgloss.Left, m.statusBarModel.View(), s)
}

func dbg(a ...any) {
	if len(os.Getenv("DEBUG")) > 0 {
		log.Println(a...)
	}
}

func main() {
	if len(os.Getenv("DEBUG")) > 0 {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		defer f.Close()
	}
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
