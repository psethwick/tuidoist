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
	"github.com/psethwick/tuidoist/input"
	"github.com/psethwick/tuidoist/overlay"
	"github.com/psethwick/tuidoist/status"
	"github.com/psethwick/tuidoist/tasklist"
)

type viewState uint

const (
	viewTasks viewState = iota
	viewChooser
	viewInput
	viewTaskMenu
)

type mainModel struct {
	client *todoist.Client
	// optimistic updates and offline actions applied
	local *todoist.Store

	height         int
	width          int
	state          viewState
	ctx            context.Context
	chooseModel    chooseModel
	taskList       tasklist.TaskList
	inputModel     input.InputModel
	taskMenuModel  taskMenuModel
	statusBarModel status.Model
	refresh        func()
	gMenu          bool
	sub            chan struct{}

	cmdQueue *todoist.Commands

	projectId string
	sectionId string
}

type syncedMsg struct{}

func waitForSync(sub chan struct{}) tea.Cmd {
	return func() tea.Msg {
		return syncedMsg(<-sub)
	}
}

func initialModel() *mainModel {
	m := mainModel{}
	m.client, m.local, m.cmdQueue = client.GetClient(dbg)
	m.ctx = context.Background()
	m.chooseModel = newChooseModel(&m)
	m.refresh = func() {}
	m.taskMenuModel = newTaskMenuModel(&m)
	m.statusBarModel = status.New()
	m.taskList = tasklist.New(func(t string) { m.statusBarModel.SetTitle(t) }, dbg)
	m.inputModel = input.New(func() { m.state = viewInput }, func() { m.state = viewTasks })
	m.applyCmds(*m.cmdQueue) // update the local store with unflushed commands
	m.sub = make(chan struct{})
	return &m
}

func (m *mainModel) Init() tea.Cmd {
	m.openInbox()
	return tea.Sequence(m.sync(), waitForSync(m.sub))
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
					m.openInbox()
					m.gMenu = false
				}
			case "t":
				if m.gMenu {
					cmds = append(cmds, m.chooseModel.gotoFilter(
						filter{todoist.Filter{Name: "Today", Query: "today | overdue"}}),
					)
					m.gMenu = false
				}
			case "ctrl+p":
				cmds = append(cmds, m.OpenPalette())
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
				// cmds = append(cmds, m.undoCompleteTask())

			// case "enter":
			// 	t, err := m.taskList.GetCursorItem()
			// 	if err != nil {
			// 		dbg(err)
			// 	} else {
			// 		m.taskMenuModel.project = m.store.FindProject(t.Item.ProjectID)
			// 		m.taskMenuModel.item = t.Item
			// 		m.taskMenuModel.content.SetValue(t.Item.Content)
			// 		m.taskMenuModel.desc.SetValue(t.Item.Description)
			// 		m.state = viewTaskMenu
			// 	}
			case "a":
				m.taskList.Bottom()
				m.inputModel.GetRepeat("add task", "", m.addTask)
				m.state = viewInput
			default:
				m.gMenu = false
			}
		}
	case viewInput:
		cmds = append(cmds, m.inputModel.Update(msg))
	case viewTaskMenu:
		cmds = append(cmds, m.taskMenuModel.Update(msg))
	}
	cmds = append(cmds, m.statusBarModel.Update(msg))
	return m, tea.Batch(cmds...)
}

func (m *mainModel) View() string {
	base := lipgloss.Place(
		m.width, m.height, lipgloss.Left, lipgloss.Top,
		lipgloss.JoinVertical(lipgloss.Left, m.statusBarModel.View(), m.taskList.View()),
	)
	s := ""
	switch m.state {
	// case viewTasks:
	// 	s = m.taskList.View()
	// case viewTaskMenu: // todo do I really need this? might still be useful tbf
	// 	s = m.taskMenuModel.View()
	case viewChooser:
		s = m.chooseModel.View()
	case viewInput:
		s = m.inputModel.View()
	}

	return overlay.PlaceOverlay(10, 1, s, base)
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
