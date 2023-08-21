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
)

type viewState uint

const (
	tasksState viewState = iota
	chooseState
	newTaskState
)

type mainModel struct {
	client       *todoist.Client
	size         tea.WindowSizeMsg
	state        viewState
	ctx          context.Context
	chooseModel  chooseModel
	tasksModel   tasksModel
	newTaskModel newTaskModel
}

func initialModel() *mainModel {
	m := mainModel{}
	m.client = client.GetClient(dbg)
	m.ctx = context.Background()
	m.tasksModel = newTasksModel(&m)
	m.chooseModel = newChooseModel(&m)
	m.newTaskModel = newNewTaskModel(&m)
	return &m
}

func (m *mainModel) refreshFromStore() tea.Cmd {
	m.tasksModel.refresh()
	return nil
}

func (m *mainModel) sync() tea.Msg {
	err := m.client.Sync(m.ctx)
	if err != nil {
		dbg("Synced", err)
	}
	err = client.WriteCache(m.client.Store)
	if err != nil {
		dbg(err)
	}
	m.refreshFromStore()
	return nil
}

func (m *mainModel) Init() tea.Cmd {
	return tea.Batch(m.refreshFromStore(), m.sync)
}

func (m *mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := listStyle.GetFrameSize()
		m.size = msg
		m.tasksModel.tasks.SetSize(msg.Width-h, msg.Height-v)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	switch m.state {
	case chooseState:
		cmd = m.chooseModel.Update(msg)
	case tasksState:
		cmd = m.tasksModel.Update(msg)
	case newTaskState:
		cmd = m.newTaskModel.Update(msg)
	}
	return m, cmd
}

func (m *mainModel) View() string {
	var s string
	switch m.state {
	case chooseState:
		s = m.chooseModel.View()
	case tasksState:
		s = m.tasksModel.View()
	case newTaskState:
		s = lipgloss.JoinVertical(
			lipgloss.Left,
			m.tasksModel.View(),
			m.newTaskModel.View(),
		)
	}
	return s
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
