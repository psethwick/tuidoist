package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
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
	viewHelp
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
	helpModel      help.Model
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
		s := <-sub
		return syncedMsg(s)
	}
}

func initialModel() *mainModel {
	m := mainModel{}
	m.client, m.local, m.cmdQueue = client.GetClient(dbg)
	m.ctx = context.Background()
	m.chooseModel = newChooseModel(&m)
	m.refresh = func() {}
	m.helpModel = help.New()
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
	return tea.Batch(waitForSync(m.sub), m.sync())
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
		if key.Matches(msg, GlobalKeys.Quit) {
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
				case key.Matches(msg, TaskListKeys.Unselect):
					m.taskList.Unselect()
				case key.Matches(msg, TaskListKeys.Select):
					m.taskList.Select()
				case key.Matches(msg, TaskListKeys.Quit):
					return m, tea.Quit
				case key.Matches(msg, TaskListKeys.Down):
					m.taskList.MoveCursor(1)
				case key.Matches(msg, TaskListKeys.Up):
					m.taskList.MoveCursor(-1)
				case key.Matches(msg, TaskListKeys.Left):
					m.resetRefresh(m.taskList.PrevList())
				case key.Matches(msg, TaskListKeys.Right):
					m.resetRefresh(m.taskList.NextList())
				case key.Matches(msg, TaskListKeys.VisitLinks):
					for _, t := range m.taskList.SelectedItems() {
						if t.Url != "" {
							cmds = append(cmds, m.OpenUrl(t.Url))
						}
					}
				case key.Matches(msg, TaskListKeys.Bottom):
					m.taskList.Bottom()
				case key.Matches(msg, TaskListKeys.GMenu):
					m.gMenu = true
				case key.Matches(msg, TaskListKeys.Complete):
					cmds = append(cmds, m.completeTasks())
					m.state = viewTasks
				case key.Matches(msg, TaskListKeys.Delete):
					cmds = append(cmds, m.deleteTasks())
				case key.Matches(msg, TaskListKeys.OpenPalette):
					cmds = append(cmds, m.OpenPalette())
				case key.Matches(msg, TaskListKeys.MoveToProject):
					cmds = append(cmds, m.OpenProjects(moveToProject))
				case key.Matches(msg, TaskListKeys.PageHalfUp):
					m.taskList.HalfPageUp()
				case key.Matches(msg, TaskListKeys.PageHalfDown):
					m.taskList.HalfPageDown()
				case key.Matches(msg, TaskListKeys.PageDown):
					m.taskList.WholePageDown()
				case key.Matches(msg, TaskListKeys.PageUp):
					m.taskList.WholePageUp()
				case key.Matches(msg, TaskListKeys.Reschedule):
					m.inputModel.GetOnce("reschedule >", "", m.rescheduleTasks)
					m.state = viewInput
				case key.Matches(msg, TaskListKeys.AddTask):
					m.taskList.Bottom()
					m.inputModel.GetRepeat("add >", "", m.addTask)
					m.state = viewInput
				case key.Matches(msg, TaskListKeys.RaisePriority):
					if item, err := m.taskList.GetCursorItem(); err == nil {
						item.Item.Priority = min(4, (item.Item.Priority + 1))
						cmds = append(cmds, m.UpdateItem(item.Item))
					}
				case key.Matches(msg, TaskListKeys.LowerPriority):
					if item, err := m.taskList.GetCursorItem(); err == nil {
						item.Item.Priority = max(1, (item.Item.Priority - 1))
						cmds = append(cmds, m.UpdateItem(item.Item))
					}
				case key.Matches(msg, TaskListKeys.SubtaskPromote):
					if item, err := m.taskList.GetCursorItem(); err == nil {
						if pitem, err := m.taskList.GetAboveItem(); err == nil {
							item.Item.ParentID = &pitem.Item.ID
							cmds = append(cmds, m.MoveItemsToNewParent(pitem.Item.ID))
						}
					}
				case key.Matches(msg, TaskListKeys.SubtaskDemote):
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
			} else { // m.gMenu true
				switch {
				case key.Matches(msg, GMenuKeys.Top):
					m.taskList.Top()
				case key.Matches(msg, GMenuKeys.Project):
					cmds = append(cmds, m.OpenProjects(chooseProject))
				case key.Matches(msg, GMenuKeys.Inbox):
					m.openInbox()
				case key.Matches(msg, GMenuKeys.Filter):
					cmds = append(cmds, m.OpenFilters())
				case key.Matches(msg, GMenuKeys.Project):
					cmds = append(cmds, m.OpenProjects(chooseProject))
				case key.Matches(msg, GMenuKeys.Today):
					cmds = append(cmds, m.chooseModel.gotoFilter(
						filter{todoist.Filter{Name: "Today", Query: "today | overdue"}}),
					)
				}
				m.gMenu = false // gmenu not sticky
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
	var bottom string
	if m.state == viewInput {
		bottom = m.inputModel.View()
	} else {
		bottom = m.helpModel.View(TaskListKeys)
	}
	base := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.Place(
			m.width, m.height-1, lipgloss.Left, lipgloss.Top,
			lipgloss.JoinVertical(
				lipgloss.Left, m.statusBarModel.View(), m.taskList.View(),
			),
		),
		bottom,
	)
	switch m.state {
	case viewChooser:
		return overlay.PlaceOverlay(10, 1, m.chooseModel.View(), base)
	}
	return base
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
