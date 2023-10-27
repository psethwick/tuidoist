package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/psethwick/tuidoist/client"
	"github.com/psethwick/tuidoist/status"
	todoist "github.com/sachaos/todoist/lib"
)

func removeItem(s []todoist.Item, itemId string) []todoist.Item {
	for i, item := range s {
		if item.ID == itemId {
			s[i] = s[len(s)-1]
			return s[:len(s)-1]
		}
	}
	return s
}

func (m *mainModel) applyCmds(cmds []todoist.Command) {
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
			item := todoist.Item{
				BaseItem: todoist.BaseItem{
					HaveProjectID: todoist.HaveProjectID{
						ProjectID: projectId,
					},
					Content: args["content"].(string),
					HaveID:  todoist.HaveID{ID: op.TempID},
				}}
			m.local.Items = append(m.local.Items, item)
			m.local.ItemMap[op.TempID] = &item
		case "item_uncomplete":
		case "item_delete":
			fallthrough
		case "item_close":
			id := args["id"].(string)
			m.local.Items = removeItem(m.local.Items, id)
			delete(m.local.ItemMap, id)
		case "item_move":
			id := args["id"].(string)
			if args["section_id"] != nil {
				sectionId := args["section_id"].(string)
				for i, item := range m.local.Items {
					if item.ID == id {
						item.SectionID = sectionId
						m.local.Items[i] = item
						break
					}
				}
			}
			if args["project_id"] != nil {
				projectId := args["project_id"].(string)
				for i, item := range m.local.Items {
					if item.ID == id {
						item.ProjectID = projectId
						m.local.Items[i] = item
						break
					}
				}
			}
		case "project_add":
			project := todoist.Project{
				HaveID: todoist.HaveID{ID: op.TempID},
				Name:   args["name"].(string),
			}
			m.local.Projects = append(m.local.Projects, project)
			m.local.ProjectMap[op.TempID] = &project
		case "project_update":
		case "item_update":
		}
	}
	m.refresh()
}

// BIG TODO need to update tempid if we get the mapping back

func (m *mainModel) sync(cmds ...todoist.Command) tea.Cmd {
	m.statusBarModel.SetSyncStatus(status.Syncing)
	m.applyCmds(cmds) // only 'new' ones
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
		m.statusBarModel.SetSyncStatus(status.Synced)
		m.sub <- struct{}{}
		return nil
	}
}
