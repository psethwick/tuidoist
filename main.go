package main

import (
	"context"
	"fmt"
	"log"
	"os"

	// todo I should make keys configurable if I wanna release it
	// "github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	todoist "github.com/sachaos/todoist/lib"
)

type viewState uint

const (
	tasksState viewState = iota
	projectState
	// filtersState
	newTaskState
)

var (
	listStyle = lipgloss.NewStyle().
			Align(lipgloss.Left).
			BorderStyle(lipgloss.HiddenBorder())
	strikeThroughStyle = lipgloss.NewStyle().Strikethrough(true)
	dialogBoxStyle     = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#874BFD")).
				Padding(1, 0).
				BorderTop(true).
				BorderLeft(true).
				BorderRight(true).
				BorderBottom(true)
	buttonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFF7DB")).
			Background(lipgloss.Color("#888B7E")).
			Padding(0, 3).
			MarginTop(1)

	activeButtonStyle = buttonStyle.Copy().
				Foreground(lipgloss.Color("#FFF7DB")).
				Background(lipgloss.Color("#F25D94")).
				MarginRight(2).
				Underline(true)
)

type mainModel struct {
	client        *todoist.Client
	size          tea.WindowSizeMsg
	state         viewState
	ctx           context.Context
	projectsModel projectsModel
	tasksModel    tasksModel
	newTaskModel  newTaskModel
	projectId     string
}

func initialModel() *mainModel {
	m := mainModel{}
	m.client = GetClient()
	m.ctx = context.Background()
	m.tasksModel = newTasksModel(&m)
	m.projectsModel = newProjectsModel(&m)
	m.newTaskModel = newNewTaskModel(&m)
	return &m
}

func (m *mainModel) refreshFromStore() tea.Cmd {
	for i, p := range m.client.Store.Projects {
		if i == 0 && m.projectId == "" {
			m.tasksModel.tasks.Title = p.Name
			m.setTasks(&p)
		} else if m.projectId == p.ID {
			m.setTasks(&p)
		}
	}
	return m.SetProjects(m.client.Store.Projects)
}

func (m *mainModel) sync() tea.Msg {
	err := m.client.Sync(m.ctx)
	if err != nil {
		dbg("Synced", err)
	}
	err = WriteCache(m.client.Store)
	if err != nil {
		dbg(err)
	}
	m.refreshFromStore()
	return nil
}

func (m *mainModel) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen, m.refreshFromStore(), m.sync)
}

func qQuits(m *mainModel, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return tea.Quit
		case "r":
			return m.sync
		}
	}
	return nil
}

func (m *mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
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
	case projectState:
		cmds = append(cmds, m.projectsModel.Update(msg))
	case tasksState:
		cmds = append(cmds, m.tasksModel.Update(msg))
	case newTaskState:
		cmds = append(cmds, m.newTaskModel.Update(msg))
	}
	return m, tea.Batch(cmds...)
}

func (m *mainModel) View() string {
	var s string
	switch m.state {
	case projectState:
		s = m.projectsModel.View()
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
