package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	// todo I should make keys configurable if I wanna release it
	// "github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/erikgeiser/promptkit/selection"
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
	projects  *selection.Model[todoist.Project]
	tasks     list.Model
	newTask   textinput.Model
	projectId string
}

func initialModel() *mainModel {
	m := mainModel{}
	m.client = GetClient()
	m.ctx = context.Background()

	m.tasks = list.New([]list.Item{}, taskDelegate(&m), 40, 30)
	m.tasks.DisableQuitKeybindings()
	m.newTask = textinput.New()
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
	if m.projects == nil {
		m.projects = selection.NewModel(sel)

	} else {
		m.projects.Selection = sel
	}
	m.projects.Filter = func(filter string, choice *selection.Choice[todoist.Project]) bool {
		return strings.Contains(strings.ToLower(choice.Value.Name), strings.ToLower(filter))
	}
	m.projects.SelectedChoiceStyle = func(c *selection.Choice[todoist.Project]) string {
		return c.Value.Name
	}
	m.projects.UnselectedChoiceStyle = func(c *selection.Choice[todoist.Project]) string {
		return c.Value.Name
	}
	return m.projects.Init()
}

func (m *mainModel) sync() func() tea.Msg {
	return func() tea.Msg {
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
}

func (m *mainModel) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen, m.refreshFromStore(), m.sync())
}

func qQuits(m *mainModel, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return tea.Quit
		case "r":
			return m.sync()
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
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	switch m.state {
	case projectView:
		// todo make each view have its own damn Update
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				p, err := m.projects.Value()
				if err == nil {
					m.projectId = p.ID
					m.setTasks(&p)
					m.switchProject(&p)
				}
				m.state = tasksView
				return m, m.projects.Init()
			case "q":
				return m, tea.Quit
			}

			_, cmd := m.projects.Update(msg)
			cmds = append(cmds, cmd)
		}
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
