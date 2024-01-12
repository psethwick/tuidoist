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
			projectId := m.local.User.InboxProjectID
			if pID, ok := args["project_id"].(string); ok {
				projectId = pID
			}
			sectionID := ""

			if args["section_id"] != nil {
				sectionID = args["section_id"].(string)
			}
			item := todoist.Item{
				// Due:         due := todoist.Due{}
				ChildOrder: 999, // hax??
				Priority:   args["priority"].(int),
				HaveSectionID: todoist.HaveSectionID{
					SectionID: sectionID,
				},
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
			if sectionId, ok := args["section_id"].(string); ok {
				for i, item := range m.local.Items {
					if item.ID == id {
						item.SectionID = sectionId
						m.local.Items[i] = item
						break
					}
				}
			}
			if projectId, ok := args["project_id"].(string); ok {
				for i, item := range m.local.Items {
					if item.ID == id {
						item.ProjectID = projectId
						m.local.Items[i] = item
						break
					}
				}
			}
		case "item_update":
			ID := args["id"].(string)
			if item, ok := m.local.ItemMap[ID]; ok {
				if content, ok := args["content"].(string); ok {
					item.Content = content
				}
				if desc, ok := args["description"].(string); ok {
					item.Description = desc
				}
				if labelNames, ok := args["labels"].([]string); ok {
					item.LabelNames = labelNames
				}
				if priority, ok := args["priority"].(int); ok {
					item.Priority = priority
				}
				if due, ok := args["due"].(todoist.Due); ok {
					item.Due = &due
				}

				m.local.Items = replace(m.local.Items, item.ID, *item)
			}
		case "project_add":
			project := todoist.Project{
				HaveID: todoist.HaveID{ID: op.TempID},
				Name:   args["name"].(string),
			}
			m.local.Projects = append(m.local.Projects, project)
		case "project_update":
			project := todoist.Project{
				HaveID: todoist.HaveID{ID: op.TempID},
				Name:   args["name"].(string),
			}
			m.local.Projects = replace(m.local.Projects, project.ID, project)
		case "section_add":
			section := todoist.Section{
				HaveID:        todoist.HaveID{ID: op.TempID},
				Name:          args["name"].(string),
				HaveProjectID: todoist.HaveProjectID{ProjectID: args["project_id"].(string)},
			}
			m.local.Sections = append(m.local.Sections, section)
		case "section_update":
			section := todoist.Section{
				HaveID:        todoist.HaveID{ID: op.TempID},
				Name:          args["name"].(string),
				HaveProjectID: todoist.HaveProjectID{ProjectID: args["project_id"].(string)},
			}
			m.local.Sections = replace(m.local.Sections, section.ID, section)
		case "section_archive":
			fallthrough
		case "section_delete":
			m.local.Sections = remove(m.local.Sections, args["id"].(string))
		case "project_archive":
			fallthrough
		case "project_delete":
			m.local.Projects = remove(m.local.Projects, args["id"].(string))
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
		dbg("adding", cmds[0])
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
		if err != nil {
			dbg(err)
			m.statusBarModel.SetSyncStatus(status.Error)
			return nil
		}
		m.client.Store.ApplyIncrementalSync(incStore)
		m.local.ApplyIncrementalSync(incStore)
		// if m.projectId == "CHANGEME" && m.local.User.InboxProjectID != "" {
		// 	dbg("CHANGEME, doing inbox")
		// 	m.openInbox()
		// } else {
		// }
		err = client.WriteCache(m.client.Store, m.cmdQueue)
		if err != nil {
			dbg(err)
			m.statusBarModel.SetSyncStatus(status.Error)
			m.sub <- struct{}{}
			return nil
		}
		m.statusBarModel.SetSyncStatus(status.Synced)
		m.sub <- struct{}{}
		dbg("refreshing")
		m.refresh()
		return nil
	}
}
