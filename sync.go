package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/psethwick/tuidoist/client"
	"github.com/psethwick/tuidoist/status"
	todoist "github.com/sachaos/todoist/lib"
)

func (m *mainModel) applyCmds(cmds []todoist.Command) {
	var dirty bool
	for _, op := range cmds {
		args := op.Args.(map[string]interface{})
		switch op.Type {
		case "item_add":
			var projectId string
			if args["project_id"] == nil {
				projectId = m.local.User.InboxProjectID
			} else {
				projectId = args["project_id"].(string)
			}
			m.local.Items = append(m.local.Items, todoist.Item{
				BaseItem: todoist.BaseItem{
					HaveProjectID: todoist.HaveProjectID{
						ProjectID: projectId,
					},
					Content: args["content"].(string),
					HaveID:  todoist.HaveID{ID: op.TempID},
				}})
			dirty = true
		case "item_delete":
		case "item_uncomplete":
		case "item_close":
		case "item_move":
		case "project_add":
		case "project_update":
		case "item_update":
		}
	}
	if dirty {
		m.local.ConstructItemTree()
	}
	m.refresh()
}
func (m *mainModel) sync(cmds ...todoist.Command) tea.Cmd {
	m.statusBarModel.SetSyncStatus(status.Syncing)
	m.applyCmds(cmds) // only 'new' ones
	m.refresh()
	*m.cmdQueue = append(*m.cmdQueue, cmds...)
	return func() tea.Msg {
		err := m.client.ExecCommands(m.ctx, *m.cmdQueue)
		if err != nil {
			dbg(err)
			// the cache is the 'real' store and any unflushed commands
			err = client.WriteCache(m.client.Store, m.cmdQueue)
			dbg(err)
			return nil // no reason to sync if api calls aren't working
		}
		m.cmdQueue = &todoist.Commands{}

		// TODO write an incremental sync implementation
		err = m.client.Sync(m.ctx)
		if err != nil {
			dbg(err)
			m.statusBarModel.SetSyncStatus(status.Error)
			m.sub <- struct{}{}
			return nil
		}
		err = client.WriteCache(m.client.Store, m.cmdQueue)
		if err != nil {
			dbg(err)
			m.statusBarModel.SetSyncStatus(status.Error)
			m.sub <- struct{}{}
			return nil
		}
		// err = client.LoadCache(m.client.Store, m.cmdQueue)
		m.refresh()
		m.statusBarModel.SetSyncStatus(status.Synced)
		m.sub <- struct{}{}
		return nil
	}
}
