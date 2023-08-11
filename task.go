package main

import (
    "fmt"
    "time"
    "strings"

    "github.com/charmbracelet/bubbles/list"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/sachaos/todoist/lib"
)

type task struct {
    item todoist.Item
    title string
    desc string
}

func reformatDate(d string, from string, to string) string {
    t, err := time.Parse(from, d)
    if err != nil {
        dbg(err)
    }
    return t.Format(to)
}

func newTask(m *mainModel, item todoist.Item) task {
    indent := strings.Repeat(" ", len(todoist.SearchItemParents(m.client.Store, &item)))
    prefix := "⚪ "
    if indent != "" {
        // subtask indicator
        prefix = fmt.Sprint("╰ ", prefix)
    }
    title := fmt.Sprintf("%s%s%s", indent, prefix, item.Content)
    desc := ""
    if item.Due != nil {
        // heres where reminders is
        // todoist.Store.Reminders
        // ⏰ ??
        var df string
        if strings.Contains(item.Due.Date, "T") {
            df = reformatDate(item.Due.Date, "2006-01-02T15:04:05", "02 Jan 06 15:04")
        } else {
            // date takes up same amount of space as date + time
            df = reformatDate(item.Due.Date, "2006-01-02", "02 Jan 06      ")
        }
        desc += fmt.Sprint(" 🗓️ ", df)
        if item.Due.IsRecurring {
            desc += fmt.Sprint(" 🔁 ", item.Due.String)
        }
    }
    desc = fmt.Sprint(indent, desc)
    return task {
        item: item,
        title: title,
        desc: desc,
    }
}


func (t task) Title() string {
    return t.title
}

// todo priority
func (t task) Description() string {
    return t.desc
}

func (t task) FilterValue() string { return t.item.Content }

func (m *mainModel) deleteTask() func() tea.Msg {
    t := m.tasks.SelectedItem().(task)
    m.tasks.RemoveItem(m.tasks.Index())
    return func() tea.Msg {
        err := m.client.DeleteItem(m.ctx, []string{t.item.ID})
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
        err := m.client.CloseItem(m.ctx, []string{t.item.ID})
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
    m.tasks.InsertItem(len(m.client.Store.Items)+1, newTask(m, t))
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
