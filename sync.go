package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/psethwick/tuidoist/client"
	"github.com/psethwick/tuidoist/status"
	todoist "github.com/sachaos/todoist/lib"
)

func remove[T todoist.IDCarrier](s []T, ID string) []T {
	for i, thing := range s {
		if thing.GetID() == ID {
			s[i] = s[len(s)-1]
			return s[:len(s)-1]
		}
	}
	return s
}

func replace[T todoist.IDCarrier](s []T, ID string, n T) []T {
	for i, item := range s {
		if item.GetID() == ID {
			s[i] = n
			return s
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
		case "item_uncomplete":
			// TODO
		case "item_delete":
			fallthrough
		case "item_close":
			id := args["id"].(string)
			m.local.Items = remove(m.local.Items, id)
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
			// TODO
		case "item_update":
			// TODO
			// m.local.Items = replace(m.local.Items, )
		case "filter_add":
			// TODO
		case "filter_update":
			// TODO
		}
	}
	m.local.ConstructItemTree()
	m.refresh()
}

func (m *mainModel) sync(cmds ...todoist.Command) tea.Cmd {
	m.statusBarModel.SetSyncStatus(status.Syncing)
	if len(cmds) > 0 {
		m.applyCmds(cmds) // only 'new' ones
		*m.cmdQueue = append(*m.cmdQueue, cmds...)
	}
	return func() tea.Msg {
		if len(*m.cmdQueue) > 0 {
			res, err := m.client.ExecCommands(m.ctx, *m.cmdQueue)
			// TODO check res.SyncStatus and roll back failed ??
			if err != nil {
				dbg(err)
				// the cache is the 'real' store and any unflushed commands
				err = client.WriteCache(m.client.Store, m.cmdQueue)
				dbg(err)
				return nil // no reason to sync if api calls aren't working
			}
			m.cmdQueue = &todoist.Commands{}
			if res.TempIdMapping != nil {
				for temp, actual := range res.TempIdMapping {
					if item := m.local.ItemMap[temp]; item != nil {
						item.ID = actual
						m.local.Items = replace(m.local.Items, temp, *item)
						continue
					}
					if item := m.local.ProjectMap[temp]; item != nil {
						item.ID = actual
						m.local.Projects = replace(m.local.Projects, temp, *item)
					}
				}
			}
		}
		incStore, err := m.client.IncrementalSync(m.ctx, m.client.Store.SyncToken)
		m.client.Store.ApplyIncrementalSync(incStore)
		m.local.ApplyIncrementalSync(incStore)
		m.refresh()
		m.sub <- struct{}{}
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
