package main

import (
	"context"
	"fmt"
	"log"
	"os"

	// todo I should make keys configurable if I wanna release it
	// "github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	todoist "github.com/sachaos/todoist/lib"
)

type sessionState uint

const (
	tasksView sessionState = iota
	projectView
	newTaskView
)

var (
	listStyle = lipgloss.NewStyle().
			Align(lipgloss.Left).
			BorderStyle(lipgloss.HiddenBorder())
	strikeThroughStyle = lipgloss.NewStyle().Strikethrough(true)
)

type mainModel struct {
	client    *todoist.Client
	state     sessionState
	ctx       context.Context
	projects  list.Model
	tasks     list.Model
	newTask   textinput.Model
	projectId string
}

func initialModel() *mainModel {
	m := mainModel{}
	m.client = GetClient()
	m.ctx = context.Background()
	m.projects.Title = "Projects"
	m.projects = list.New([]list.Item{}, projectDelegate(&m), 40, 30)

	m.tasks = list.New([]list.Item{}, taskDelegate(&m), 40, 30)
	m.tasks.DisableQuitKeybindings()
	m.newTask = textinput.New()
	return &m
}

func (m *mainModel) refreshFromStore() {
	var projects []list.Item
	for i, p := range m.client.Store.Projects {
		if i == 0 && m.projectId == "" {
			proj := project(p)
			m.setTasks(&proj)
		} else if m.projectId == p.ID {
			proj := project(p)
			m.setTasks(&proj)
		}
		projects = append(projects, project(p))
	}
	m.projects.SetItems(projects)
}

func (m *mainModel) sync() func() tea.Msg {
	return func() tea.Msg {
		dbg("Syncing")
		err := m.client.Sync(m.ctx)
		m.refreshFromStore()
		dbg("Synced", err)
		return nil
	}
}

func (m *mainModel) Init() tea.Cmd {
	m.refreshFromStore()
	return tea.Batch(tea.EnterAltScreen, m.sync())
}

func qQuits(m *mainModel, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return tea.Quit
		}
	}
	return nil
}

func (m *mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := listStyle.GetFrameSize()
		m.tasks.SetSize(msg.Width-h, msg.Height-v)
		m.projects.SetSize(msg.Width-h, msg.Height-v)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	switch m.state {
	case projectView:
		cmds = append(cmds, qQuits(m, msg))
		m.projects, cmd = m.projects.Update(msg)
		cmds = append(cmds, cmd)
	case tasksView:
		cmds = append(cmds, qQuits(m, msg))
		m.tasks, cmd = m.tasks.Update(msg)
		cmds = append(cmds, cmd)
	case newTaskView:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				cmds = append(cmds, m.addTask())
			case "esc":
				m.newTask.SetValue("")
				m.state = tasksView
			}
		}
		m.newTask, cmd = m.newTask.Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func (m *mainModel) View() string {
	var s string
	switch m.state {
	case projectView:
		s += listStyle.Render(m.projects.View())
	case tasksView:
		s += listStyle.Render(m.tasks.View())
	case newTaskView:
		s += m.newTask.View()
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
