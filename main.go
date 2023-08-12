package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	// todo I should make keys configurable if I wanna release it
	// "github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/erikgeiser/promptkit/selection"
	todoist "github.com/sachaos/todoist/lib"
)

type viewState uint

const (
	tasksState viewState = iota
	projectState
	newTaskState
)

var (
	listStyle = lipgloss.NewStyle().
			Align(lipgloss.Left).
			BorderStyle(lipgloss.HiddenBorder())
	strikeThroughStyle = lipgloss.NewStyle().Strikethrough(true)
)

type mainModel struct {
	client     *todoist.Client
	state      viewState
	ctx        context.Context
	projects   projectsModel
	tasksModel tasksModel
	newTask    newTaskModel
	projectId  string
}

func initialModel() *mainModel {
	m := mainModel{}
	m.client = GetClient()
	m.ctx = context.Background()
	m.tasksModel = newTasksModel(&m)

	m.newTask = newNewTaskModel()
	return &m
}

func (m *mainModel) refreshFromStore() tea.Cmd {
	for i, p := range m.client.Store.Projects {
		if i == 0 && m.projectId == "" {
			m.setTasks(&p)
		} else if m.projectId == p.ID {
			m.setTasks(&p)
		}
	}
	sel := selection.New("Choose Project:", m.client.Store.Projects)
	if m.projects.projects == nil {
		m.projects.projects = selection.NewModel(sel)
	} else {
		m.projects.projects.Selection = sel
	}
	m.projects.projects.Filter = func(filter string, choice *selection.Choice[todoist.Project]) bool {
		return strings.Contains(strings.ToLower(choice.Value.Name), strings.ToLower(filter))
	}
	m.projects.projects.SelectedChoiceStyle = func(c *selection.Choice[todoist.Project]) string {
		return c.Value.Name
	}
	m.projects.projects.UnselectedChoiceStyle = func(c *selection.Choice[todoist.Project]) string {
		return c.Value.Name
	}
	return m.projects.projects.Init()
}

func (m *mainModel) sync() tea.Msg {
	dbg("Syncing")
	err := m.client.Sync(m.ctx)
	dbg("Synced", err)
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
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := listStyle.GetFrameSize()
		m.tasksModel.tasks.SetSize(msg.Width-h, msg.Height-v)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	switch m.state {
	case projectState:
		// todo make each view have its own damn Update
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				p, err := m.projects.projects.Value()
				if err == nil {
					m.projectId = p.ID
					m.setTasks(&p)
					m.switchProject(&p)
				}
				m.state = tasksState
				return m, m.projects.projects.Init()
			case "q":
				return m, tea.Quit
			}

			_, cmd := m.projects.projects.Update(msg)
			cmds = append(cmds, cmd)
		}
	case tasksState:
		cmds = append(cmds, qQuits(m, msg))
		m.tasksModel.tasks, cmd = m.tasksModel.tasks.Update(msg)
		cmds = append(cmds, cmd)
	case newTaskState:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				cmds = append(cmds, m.addTask())
			case "esc":
				m.newTask.input.SetValue("")
				m.state = tasksState
			}
		}
		m.newTask.input, cmd = m.newTask.input.Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func (m *mainModel) View() string {
	var s string
	switch m.state {
	case projectState:
		s += listStyle.Render(m.projects.projects.View())
	case tasksState:
		s += listStyle.Render(m.tasksModel.tasks.View())
	case newTaskState:
		s += m.newTask.input.View()
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
