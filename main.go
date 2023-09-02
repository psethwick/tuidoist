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
	tasksModel     tasksModel
	newTaskModel   newTaskModel
	taskMenuModel  taskMenuModel
	statusBarModel status.Model
}

func initialModel() *mainModel {
	m := mainModel{}
	m.client = client.GetClient(dbg)
	m.ctx = context.Background()
	m.tasksModel = newTasksModel(&m)
	m.chooseModel = newChooseModel(&m)
	m.newTaskModel = newNewTaskModel(&m)
	m.taskMenuModel = newTaskMenuModel(&m)
	m.statusBarModel = status.New()
	return &m
}

func (m *mainModel) refreshFromStore() tea.Cmd {
	m.tasksModel.refresh()
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
	return tea.Batch(m.refreshFromStore(), m.sync)
}

func (m *mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := listStyle.GetFrameSize()
		// h2, v2 := tuiStyle.GetFrameSize()
		m.height = msg.Height - 1 // statusbar
		m.width = msg.Width
		m.tasksModel.tasks.SetSize(msg.Width-h, msg.Height-v-1)
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
		cmds = append(cmds, m.tasksModel.Update(msg))
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
		s = m.tasksModel.View()
	case taskMenuState:
		s = m.taskMenuModel.View()
	case newTaskBottomState:
		s = lipgloss.JoinVertical(
			lipgloss.Left,
			m.tasksModel.View(),
			m.newTaskModel.View(),
		)
	case newTaskTopState:
		s = lipgloss.JoinVertical(
			lipgloss.Left,
			m.newTaskModel.View(),
			m.tasksModel.View(),
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
