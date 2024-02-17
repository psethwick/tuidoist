package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	todoist "github.com/sachaos/todoist/lib"

	"github.com/psethwick/tuidoist/client"
	"github.com/psethwick/tuidoist/input"
	"github.com/psethwick/tuidoist/keys"
	"github.com/psethwick/tuidoist/overlay"
	"github.com/psethwick/tuidoist/status"
	"github.com/psethwick/tuidoist/style"
	"github.com/psethwick/tuidoist/tasklist"
)

type viewState uint

const (
	viewTasks viewState = iota
	viewChooser
	viewInput
	viewTaskMenu
	viewHelp
	needToken
)

func (m *mainModel) viewKeyMap() help.KeyMap {
	if m.gMenu {
		return keys.GMenuKeys
	}
	switch m.state {
	case viewTasks:
		return keys.TaskListKeys
	case viewChooser:
		fallthrough
	case viewInput:
		return keys.InputKeys
	case viewTaskMenu:
		return keys.TaskListKeys
	}
	panic("rip")
}

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
	helpModel      help.Model
	statusBarModel status.Model
	refresh        func()
	sub            chan struct{}
	gMenu          bool

	cmdQueue  *todoist.Commands
	syncMutex sync.Mutex

	projectId string
	sectionId string
}

type syncedMsg struct{}

func waitForSync(sub chan struct{}) tea.Cmd {
	return func() tea.Msg {
		s := <-sub
		return syncedMsg(s)
	}
}

func initialModel() *mainModel {
	m := mainModel{}
	m.ctx = context.Background()
	m.chooseModel = newChooseModel(&m, keys.InputKeys)
	m.refresh = func() {}
	m.helpModel = help.New()
	m.helpModel.Styles = help.Styles{
		ShortKey:       style.HelpKeyStyle,
		ShortDesc:      style.HelpDescStyle,
		ShortSeparator: style.HelpSepStyle,
		Ellipsis:       style.HelpSepStyle,
		FullKey:        style.HelpKeyStyle,
		FullDesc:       style.HelpDescStyle,
		FullSeparator:  style.HelpSepStyle,
	}
	m.taskMenuModel = newTaskMenuModel(&m)
	m.statusBarModel = status.New()
	m.taskList = tasklist.New(func(t string) { m.statusBarModel.SetTitle(t) }, dbg)
	m.inputModel = input.New(func() { m.state = viewInput }, func() { m.state = viewTasks }, keys.InputKeys)

	m.local = new(todoist.Store)
	m.cmdQueue = new(todoist.Commands)
	m.syncMutex = sync.Mutex{}

	m.sub = make(chan struct{})
	return &m
}

func (m *mainModel) Init() tea.Cmd {
	var store todoist.Store
	err := client.LoadCache(&store, m.local, m.cmdQueue)
	if err != nil {
		panic(err)
	}
	m.applyCmds(*m.cmdQueue) // update the local store with unflushed commands
	clnt, err := client.GetClient(dbg, &store)
	if clnt != nil {
		m.client = clnt
	}
	if err != nil {
		if err.Error() == "need token" {
			return m.requestApiToken()
		} else {
			panic(err)
		}
	}
	return tea.Batch(waitForSync(m.sub), m.sync(),
		func() tea.Msg {
			m.openInbox()
			return nil
		})
}

func (m *mainModel) resetRefresh(listId interface{}) {
	switch typed := listId.(type) {
	case filterSelection:
		m.refresh = func() {
			m.setTasksFromFilter(typed)
		}
	case project:
		m.projectId = typed.project.ID
		m.sectionId = typed.section.ID
		m.refresh = func() {
			m.setTasksFromProject(&typed)
		}
	}
}

func (m *mainModel) requestApiToken() tea.Cmd {
	m.inputModel.GetOnce(
		"API token from Todoist Integrations",
		"",
		func(s string) tea.Cmd {
			err := client.SetToken(s)
			if err != nil {
				dbg(err)
			}
			var store todoist.Store
			clnt, err := client.GetClient(dbg, &store)
			if err != nil {
				dbg(err)
			}
			if clnt != nil {
				m.client = clnt
			}
			return tea.Sequence(tea.Batch(waitForSync(m.sub), m.sync()), func() tea.Msg {
				m.openInbox()
				return nil
			})
		})
	return nil
}

func (m *mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		// for children to get the size they can actually have
		m.taskList.SetHeight(msg.Height - 2)
		m.taskList.SetWidth(msg.Width)
	case syncedMsg:
		return m, waitForSync(m.sub)
	case tea.KeyMsg:
		if key.Matches(msg, keys.GlobalKeys.Quit) {
			return m, tea.Quit
		}
	}
	switch m.state {
	case viewChooser:
		cmds = append(cmds, m.chooseModel.Update(msg))
	case viewTasks:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if !m.gMenu {
				switch {
				case key.Matches(msg, keys.TaskListKeys.Help):
					m.helpModel.ShowAll = true
				case key.Matches(msg, keys.TaskListKeys.Cancel):
					m.helpModel.ShowAll = false
					m.taskList.Unselect()
				case key.Matches(msg, keys.TaskListKeys.Select):
					m.taskList.Select()
				case key.Matches(msg, keys.TaskListKeys.Quit):
					return m, tea.Quit
				case key.Matches(msg, keys.TaskListKeys.Down):
					m.taskList.MoveCursor(1)
				case key.Matches(msg, keys.TaskListKeys.Up):
					m.taskList.MoveCursor(-1)
				case key.Matches(msg, keys.TaskListKeys.Left):
					m.resetRefresh(m.taskList.PrevList())
				case key.Matches(msg, keys.TaskListKeys.Right):
					m.resetRefresh(m.taskList.NextList())
				case key.Matches(msg, keys.TaskListKeys.VisitLinks):
					for _, t := range m.taskList.SelectedItems() {
						if t.Url != "" {
							cmds = append(cmds, m.OpenUrl(t.Url))
						}
					}
				case key.Matches(msg, keys.TaskListKeys.Bottom):
					m.taskList.Bottom()
				case key.Matches(msg, keys.TaskListKeys.GMenu):
					m.gMenu = true
				case key.Matches(msg, keys.TaskListKeys.Complete):
					cmds = append(cmds, m.completeTasks())
					m.state = viewTasks
				case key.Matches(msg, keys.TaskListKeys.Delete):
					cmds = append(cmds, m.deleteTasks())
				case key.Matches(msg, keys.TaskListKeys.OpenPalette):
					cmds = append(cmds, m.OpenPalette())
				case key.Matches(msg, keys.TaskListKeys.MoveToProject):
					cmds = append(cmds, m.OpenProjects(moveToProject))
				case key.Matches(msg, keys.TaskListKeys.PageHalfUp):
					m.taskList.HalfPageUp()
				case key.Matches(msg, keys.TaskListKeys.PageHalfDown):
					m.taskList.HalfPageDown()
				case key.Matches(msg, keys.TaskListKeys.PageDown):
					m.taskList.WholePageDown()
				case key.Matches(msg, keys.TaskListKeys.PageUp):
					m.taskList.WholePageUp()
				case key.Matches(msg, keys.TaskListKeys.Reschedule):
					m.inputModel.GetOnce("reschedule >", "", m.rescheduleTasks)
					m.state = viewInput
				case key.Matches(msg, keys.TaskListKeys.MoveUp):
					cmds = append(cmds, m.ReorderTasks(m.taskList.SendToTop()))
				case key.Matches(msg, keys.TaskListKeys.MoveDown):
					cmds = append(cmds, m.ReorderTasks(m.taskList.SendToBottom()))
				case key.Matches(msg, keys.TaskListKeys.AddTask):
					m.inputModel.GetRepeat("add >", "", m.addTaskBottom)
					m.state = viewInput
					m.taskList.Bottom()
				case key.Matches(msg, keys.TaskListKeys.AddTaskTop):
					m.inputModel.GetRepeat("add >", "", m.addTaskTop)
					m.state = viewInput
					m.taskList.Top()
				case key.Matches(msg, keys.TaskListKeys.Priority1):
					cmds = append(cmds, m.setPriority(4))
				case key.Matches(msg, keys.TaskListKeys.Priority2):
					cmds = append(cmds, m.setPriority(3))
				case key.Matches(msg, keys.TaskListKeys.Priority3):
					cmds = append(cmds, m.setPriority(2))
				case key.Matches(msg, keys.TaskListKeys.Priority4):
					cmds = append(cmds, m.setPriority(1))
				case key.Matches(msg, keys.TaskListKeys.SubtaskPromote):
					if item, err := m.taskList.GetCursorItem(); err == nil {
						if pitem, err := m.taskList.GetAboveItem(); err == nil {
							item.Item.ParentID = &pitem.Item.ID
							cmds = append(cmds, m.MoveItemsToNewParent(pitem.Item.ID))
						}
					}
				case key.Matches(msg, keys.TaskListKeys.SubtaskDemote):
					if item, err := m.taskList.GetCursorItem(); err == nil {
						parents := todoist.SearchItemParents(m.local, &item.Item)
						newParentId := ""
						for i := 0; i < len(parents)-1; i++ {
							if i >= 0 && i <= len(parents) {
								dbg(i, parents[i].Content)
								newParentId = parents[i].ID
							}
						}
						cmds = append(cmds, m.MoveItemsToNewParent(newParentId))
					}
				}
			} else { // gMenu
				isHelp := key.Matches(msg, keys.GMenuKeys.Help)
				if !isHelp {
					m.helpModel.ShowAll = false
				}
				switch {
				case isHelp:
					m.helpModel.ShowAll = true
				case key.Matches(msg, keys.GMenuKeys.Top):
					m.taskList.Top()
				case key.Matches(msg, keys.GMenuKeys.Project):
					cmds = append(cmds, m.OpenProjects(chooseProject))
				case key.Matches(msg, keys.GMenuKeys.Inbox):
					m.openInbox()
				case key.Matches(msg, keys.GMenuKeys.Filter):
					cmds = append(cmds, m.OpenFilters())
				case key.Matches(msg, keys.GMenuKeys.Project):
					cmds = append(cmds, m.OpenProjects(chooseProject))
				case key.Matches(msg, keys.GMenuKeys.Today):
					cmds = append(cmds, m.chooseModel.gotoFilter(
						filter{todoist.Filter{Name: "Today", Query: "today | overdue"}}),
					)
				}
				if !m.helpModel.ShowAll {
					m.gMenu = false // gmenu not sticky
				}
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
	keyMap := m.viewKeyMap()
	base := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.Place(
			m.width, m.height-1, lipgloss.Left, lipgloss.Top,
			lipgloss.JoinVertical(
				lipgloss.Left, m.statusBarModel.View(), m.taskList.View(),
			),
		),
		m.helpModel.ShortHelpView(keyMap.ShortHelp()),
	)
	popup := ""
	switch {
	case m.state == viewInput:
		popup = m.inputModel.View()
	case m.helpModel.ShowAll:
		popup = m.helpModel.View(keyMap)
	case m.state == viewChooser:
		popup = m.chooseModel.View()
	}
	if popup != "" {
		return overlay.PlaceOverlay(style.DialogBoxStyle, popup, base)
	} else {
		return base
	}
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
	p := tea.NewProgram(initialModel()) //, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
