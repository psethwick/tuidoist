package main

import (
    "fmt"
    "time"
    "strings"

    "github.com/charmbracelet/bubbles/list"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/sachaos/todoist/lib"
)

type task todoist.Item

func (i task) indent() string {
    return strings.Repeat("  ", i.Indent)
}

func (i task) Title() string {
    prefix := "⚪ "
    if i.Indent > 0 {
        prefix = "╰ ⚪ "
    }
    return fmt.Sprintf("%s%s%s", i.indent(), prefix, i.Content)
}



func reformatDate(d string, from string, to string) string {
    t, err := time.Parse(from, d)
    if err != nil {
        dbg(err)
    }
    return t.Format(to)
}

// todo priority
// todo no reason to recalc this all the time
func (i task) Description() string {
    extras := ""
    if i.Due != nil {
        // heres where reminders is
        // todoist.Store.Reminders
        // ⏰ ??
        var df string
        if strings.Contains(i.Due.Date, "T") {
            df = reformatDate(i.Due.Date, "2006-01-02T15:04:05", "02 Jan 06 15:04")
        } else {
            // date takes up same amount of space as date + time
            df = reformatDate(i.Due.Date, "2006-01-02", "02 Jan 06      ")
        }
        extras += fmt.Sprint(" 🗓️ ", df)
        if i.Due.IsRecurring {
            extras += fmt.Sprint(" 🔁 ", i.Due.String)
        }
    }
    return fmt.Sprint(i.indent(), extras)
}

func (i task) FilterValue() string { return i.Content }

func (m *mainModel) deleteTask() func() tea.Msg {
    t := m.tasks.SelectedItem().(task)
    m.tasks.RemoveItem(m.tasks.Index())
    return func() tea.Msg {
        err := m.client.DeleteItem(m.ctx, []string{t.ID})
        if err != nil {
            dbg("del error", err)
        }
        return m.sync()
    }
}

func (m *mainModel) completeTask() func() tea.Msg {
    t := m.tasks.SelectedItem().(task)
    m.tasks.RemoveItem(m.tasks.Index())
    return func() tea.Msg {
        err := m.client.CloseItem(m.ctx, []string{t.ID})
        if err != nil {
            dbg("complete task err", err)
        }
        return m.sync()
    }
}

func (m *mainModel) addTask() func() tea.Msg {
    content := m.newTask.Value()
    m.newTask.SetValue("")
    if content == "" {
        return func() tea.Msg { return nil }
    }
    t := todoist.Item{}
    t.ProjectID = m.projectId
    t.Content = content
    // todo priority, description, labels
    m.tasks.InsertItem(len(m.client.Store.Items)+1, task(t))
    return func() tea.Msg {
        m.client.AddItem(m.ctx, t)
        return m.sync()
    }
}

func taskDelegate(m *mainModel) list.DefaultDelegate {
    d := list.NewDefaultDelegate()
    d.UpdateFunc = func(msg tea.Msg, l *list.Model) tea.Cmd {
        var cmds []tea.Cmd
        switch msg := msg.(type) {
        case tea.KeyMsg:
            switch msg.String() {
            case "p":
                cmds = append(cmds, tea.ClearScreen)
                m.state = projectView
            case "C":
                cmds = append(cmds, m.completeTask())
            case "D":
                cmds = append(cmds, m.deleteTask())
            case "n":
                m.newTask.Prompt = "> "
                m.newTask.Focus()
                cmds = append(cmds, tea.ClearScreen)
                m.state = newTaskView
        }
    }

        return tea.Batch(cmds...)
    }
    return d
}
