package main

import (
	"context"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	todoist "github.com/sachaos/todoist/lib"

	"github.com/psethwick/tuidoist/client"
	"github.com/psethwick/tuidoist/status"
	"github.com/psethwick/tuidoist/tasklist"
)

type viewState uint

const (
	viewTasks viewState = iota
	viewChooser
	viewNewTaskTop
	viewNewTaskBottom
	viewTaskMenu
	viewAddProject
)

type mainModel struct {
	client         *todoist.Client
	height         int
	width          int
	state          viewState
	ctx            context.Context
	chooseModel    chooseModel
	taskList       tasklist.TaskList
	inputModel     inputModel
	taskMenuModel  taskMenuModel
	statusBarModel status.Model
	refresh        func()
	gMenu          bool
	sub            chan struct{}
}

type syncedMsg struct{}

func waitForSync(sub chan struct{}) tea.Cmd {
	return func() tea.Msg {
		return syncedMsg(<-sub)
	}
}

func initialModel() *mainModel {
	m := mainModel{}
	m.client = client.GetClient(dbg)
	m.ctx = context.Background()
	m.chooseModel = newChooseModel(&m)
	m.inputModel = newInputModel(&m)
	m.taskMenuModel = newTaskMenuModel(&m)
	m.statusBarModel = status.New()
	m.taskList = tasklist.New(func(t string) { m.statusBarModel.SetTitle(t) }, dbg)
	m.sub = make(chan struct{})
	return &m
}

func (m *mainModel) sync() tea.Msg {
	m.statusBarModel.SetSyncStatus(status.Syncing)
	m.sub <- struct{}{}
	err := m.client.Sync(m.ctx)
	if err != nil {
		dbg(err)
		m.statusBarModel.SetSyncStatus(status.Error)
		m.sub <- struct{}{}
		return nil
	}
	err = client.WriteCache(m.client.Store)
	if err != nil {
		dbg(err)
		m.statusBarModel.SetSyncStatus(status.Error)
		m.sub <- struct{}{}
		return nil
	}
	m.refresh()
	m.statusBarModel.SetSyncStatus(status.Synced)
	m.sub <- struct{}{}
	return nil
}

func (m *mainModel) Init() tea.Cmd {
	m.refresh = func() {
		for _, tp := range m.client.Store.Projects {
			// todo default view prefs
			if tp.Name == "Inbox" {
				p := project{tp, todoist.Section{}}
				m.setTasksFromProject(&p)
				break
			}
		}
	}
	m.refresh()
	return tea.Batch(m.sync, waitForSync(m.sub))
}

// undo
// add -> delete
// complete -> uncomplete
// delete -> re-add? I will need the whole task...

func (m *mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		// for children to get the size they can actually have
		m.taskList.SetHeight(msg.Height - 1)
		m.taskList.SetWidth(msg.Width)
	case syncedMsg:
		return m, waitForSync(m.sub)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+w":
			fallthrough
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	switch m.state {
	case viewChooser:
		cmds = append(cmds, m.chooseModel.Update(msg))
	case viewTasks:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "q":
				return m, tea.Quit
			case "j":
				m.taskList.MoveCursor(1)
			case "k":
				m.taskList.MoveCursor(-1)
			case "h":
				m.taskList.PrevList()
			case "l":
				m.taskList.NextList()
			case "v":
				t, _ := m.taskList.GetCursorItem()
				if t.Url != "" {
					cmds = append(cmds, m.OpenUrl(t.Url))
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
				m.state = viewTasks
			case "delete":
				cmds = append(cmds, m.deleteTask())
			case "f":
				if m.gMenu {
					cmds = append(cmds, m.OpenFilters())
					m.gMenu = false
				}
			case "i":
				if m.gMenu {
					cmds = append(cmds, m.openInbox())
					m.gMenu = false
				}
			case "t":
				if m.gMenu {
					cmds = append(cmds, m.chooseModel.gotoFilter(filter{Name: "Today", Query: "today | overdue"}))
					m.gMenu = false
				}
			case "ctrl+p":
				dbg("ctrl+p")
				cmds = append(cmds, m.OpenPalette(paletteProject, paletteTask))
			case "p":
				if m.gMenu {
					cmds = append(cmds, m.OpenProjects(chooseProject))
					m.gMenu = false
				} else {
					m.statusBarModel.SetSort(m.taskList.Sort(tasklist.PrioritySort))
				}
			case "n":
				m.statusBarModel.SetSort(m.taskList.Sort(tasklist.NameSort))
			case "d":
				m.statusBarModel.SetSort(m.taskList.Sort(tasklist.DateSort))
			case "r":
				m.statusBarModel.SetSort(m.taskList.Sort(tasklist.AssigneeSort))
			case "m":
				cmds = append(cmds, m.OpenProjects(moveToProject))
			case "ctrl+u":
				m.taskList.HalfPageUp()
			case "ctrl+d":
				m.taskList.HalfPageDown()
			case "ctrl+f":
				m.taskList.WholePageDown()
			case "ctrl+b":
				m.taskList.WholePageUp()

			case "ctrl+z":
				fallthrough
			case "z":
				cmds = append(cmds, m.undoCompleteTask())
			case "enter":
				t, err := m.taskList.GetCursorItem()
				if err != nil {
					dbg(err)
				} else {
					m.taskMenuModel.project = m.client.Store.FindProject(t.Item.ProjectID)
					m.taskMenuModel.item = t.Item
					m.taskMenuModel.content.SetValue(t.Item.Content)
					m.taskMenuModel.desc.SetValue(t.Item.Description)
					m.state = viewTaskMenu
				}
			case "a":
				m.taskList.SetHeight(m.height - m.inputModel.Height() - 1)
				m.taskList.Bottom()
				m.inputModel.content.Focus()
				m.inputModel.purpose = inputAddTask
				m.state = viewNewTaskBottom
			case "A":
				m.taskList.SetHeight(m.height - m.inputModel.Height() - 1)
				m.taskList.Top()
				m.inputModel.content.Focus()
				m.inputModel.purpose = inputAddTask
				m.state = viewNewTaskTop
			default:
				m.gMenu = false
			}
		}
	case viewAddProject:
		fallthrough
	case viewNewTaskTop:
		fallthrough
	case viewNewTaskBottom:
		cmds = append(cmds, m.inputModel.Update(msg))
	case viewTaskMenu:
		cmds = append(cmds, m.taskMenuModel.Update(msg))
	}
	cmds = append(cmds, m.statusBarModel.Update(msg))
	return m, tea.Batch(cmds...)
}

func (m *mainModel) View() string {
	var s string
	switch m.state {
	case viewChooser:
		s = m.chooseModel.View()
	case viewTasks:
		s = m.taskList.View()
	case viewTaskMenu:
		s = m.taskMenuModel.View()
	case viewAddProject:
		fallthrough
	case viewNewTaskBottom:
		s = lipgloss.JoinVertical(
			lipgloss.Left,
			m.taskList.View(),
			m.inputModel.View(),
		)
	case viewNewTaskTop:
		s = lipgloss.JoinVertical(
			lipgloss.Left,
			m.inputModel.View(),
			m.taskList.View(),
		)
	}
	return lipgloss.Place(
		m.width, m.height, lipgloss.Left, lipgloss.Top,
		lipgloss.JoinVertical(lipgloss.Left, m.statusBarModel.View(), s),
	)
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
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
