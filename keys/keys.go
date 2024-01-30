package keys

import (
	"github.com/charmbracelet/bubbles/key"
)

type GlobalKeyMap struct {
	Quit key.Binding
}

var GlobalKeys = GlobalKeyMap{
	Quit: key.NewBinding(key.WithKeys("ctrl+c", "ctrl+w")),
}

type TaskListKeyMap struct {
	AddTask        key.Binding
	Bottom         key.Binding
	Complete       key.Binding
	Delete         key.Binding
	Down           key.Binding
	GMenu          key.Binding
	Help           key.Binding
	Left           key.Binding
	LowerPriority  key.Binding
	MoveToProject  key.Binding
	OpenPalette    key.Binding
	PageDown       key.Binding
	PageHalfDown   key.Binding
	PageHalfUp     key.Binding
	PageUp         key.Binding
	Quit           key.Binding
	RaisePriority  key.Binding
	Reschedule     key.Binding
	Right          key.Binding
	Select         key.Binding
	MoveUp         key.Binding
	MoveDown       key.Binding
	SubtaskDemote  key.Binding
	SubtaskPromote key.Binding
	Cancel         key.Binding
	Up             key.Binding
	VisitLinks     key.Binding
}

func (k TaskListKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.AddTask, k.Complete, k.GMenu, k.OpenPalette, k.Help, k.Quit}
}

func (k TaskListKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			k.Up,
			k.Down,
			k.Left,
			k.Right,
			k.PageUp,
			k.PageDown,
			k.PageHalfUp,
			k.PageHalfDown,
			k.Bottom,
		},
		{
			k.Complete,
			k.RaisePriority,
			k.LowerPriority,
			k.Reschedule,
			k.Delete,
			k.SubtaskDemote,
			k.SubtaskPromote,
			k.VisitLinks,
			k.MoveToProject,
			k.MoveUp,
			k.MoveDown,
		},
		{
			k.GMenu,
			k.AddTask,
			k.Select,
			k.OpenPalette,
			k.Cancel,
			k.Quit,
		},
	}
}

var TaskListKeys = TaskListKeyMap{
	AddTask:        key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add task")),
	Bottom:         key.NewBinding(key.WithKeys("G"), key.WithHelp("G", "goto bottom")),
	Complete:       key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "complete")),
	Delete:         key.NewBinding(key.WithKeys("del"), key.WithHelp("del", "delete")),
	Down:           key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("↓/j", "down")),
	GMenu:          key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "submenu")),
	Help:           key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	Left:           key.NewBinding(key.WithKeys("h", "left"), key.WithHelp("←/h", "left")),
	LowerPriority:  key.NewBinding(key.WithKeys("-"), key.WithHelp("-", "lower priority")),
	MoveToProject:  key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "move to project")),
	MoveUp:         key.NewBinding(key.WithKeys("K"), key.WithHelp("K", "move up")),
	MoveDown:       key.NewBinding(key.WithKeys("J"), key.WithHelp("J", "move down")),
	OpenPalette:    key.NewBinding(key.WithKeys("ctrl+p"), key.WithHelp("^p", "command palette")),
	PageDown:       key.NewBinding(key.WithKeys("ctrl+f"), key.WithHelp("^f", "page down")),
	PageHalfDown:   key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("^d", "half page down")),
	PageHalfUp:     key.NewBinding(key.WithKeys("ctrl+u"), key.WithHelp("^u", "half page up")),
	PageUp:         key.NewBinding(key.WithKeys("ctrl+b"), key.WithHelp("^b", "page up")),
	Quit:           key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
	RaisePriority:  key.NewBinding(key.WithKeys("+"), key.WithHelp("+", "raise priority")),
	Reschedule:     key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "reschedule")),
	Right:          key.NewBinding(key.WithKeys("l", "right"), key.WithHelp("→/l", "right")),
	Select:         key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "toggle select")),
	SubtaskDemote:  key.NewBinding(key.WithKeys("<"), key.WithHelp("<", "demote subtask")),
	SubtaskPromote: key.NewBinding(key.WithKeys(">"), key.WithHelp(">", "promote subtask")),
	Cancel:         key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
	Up:             key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("↑/k", "up")),
	VisitLinks:     key.NewBinding(key.WithKeys("v"), key.WithHelp("v", "visit url")),
}

type gKeyMap struct {
	Exit    key.Binding
	Filter  key.Binding
	Help    key.Binding
	Inbox   key.Binding
	Now     key.Binding
	Project key.Binding
	Today   key.Binding
	Top     key.Binding
	Cancel  key.Binding
}

func (k gKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Inbox, k.Project, k.Help, k.Exit}
}

func (k gKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Inbox, k.Today, k.Project, k.Filter, k.Now, k.Top},
		{k.Cancel, k.Exit},
	}
}

var GMenuKeys = gKeyMap{
	Exit:    key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "exit submenu")),
	Filter:  key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "filter")),
	Help:    key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	Inbox:   key.NewBinding(key.WithKeys("i"), key.WithHelp("i", "inbox")),
	Now:     key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "now")),
	Project: key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "project")),
	Today:   key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "today")),
	Top:     key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "top")),
	Cancel:  key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
}

var InputKeys = InputKeyMap{
	Accept: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "accept")),
	Cancel: key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
}

type InputKeyMap struct {
	Accept key.Binding
	Cancel key.Binding
}

func (k InputKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Accept, k.Cancel}
}

func (k InputKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Accept, k.Cancel}}
}
