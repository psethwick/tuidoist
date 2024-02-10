package main

import (
	"context"

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

func removeWithChildren(s *todoist.Store, itemID string) {
	removeIds := []string{itemID}
	item := s.ItemMap[itemID]

	childItem := item.ChildItem
	for childItem != nil {
		removeIds = append([]string{childItem.ID}, removeIds...)
		brotherItem := childItem.BrotherItem
		for brotherItem != nil {
			removeIds = append([]string{brotherItem.ID}, removeIds...)
			brotherItem = brotherItem.BrotherItem
		}
		childItem = childItem.ChildItem
	}

	for _, ri := range removeIds {
		s.Items = remove(s.Items, ri)
	}
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
		case "item_reorder":
			items := args["items"].([]map[string]interface{})
			for _, item := range items {
				id := item["id"].(string)
				order := item["child_order"].(int)
				if _, ok := m.local.ItemMap[id]; ok {
					m.local.ItemMap[id].ChildOrder = order
				}
			}
		case "item_uncomplete":
			// TODO
		case "item_delete":
			fallthrough
		case "item_close":
			id := args["id"].(string)
			removeWithChildren(m.local, id)
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
			if parentId, ok := args["parent_id"].(string); ok {
				for i, item := range m.local.Items {
					if item.ID == id {
						if parentId == "" {
							item.ParentID = nil
						} else {
							item.ParentID = &parentId
						}
						m.local.Items[i] = item
						break
					}
				}
			}
		case "item_update":
			ID := args["id"].(string)
			for _, item := range m.local.Items {
				if item.ID == ID {

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

					m.local.Items = replace(m.local.Items, item.ID, item)
				}
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
	// damned if we do, damned if we don't
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
			// todo here are there other temp mappings to worry about?
			if res.TempIdMapping != nil {
				for temp, actual := range res.TempIdMapping {
					for _, item := range m.local.Items {
						if item.ID == temp {
							item.ID = actual
							m.local.Items = replace(m.local.Items, temp, item)
							continue
						}
					}
					for _, item := range m.local.Projects {
						if item.ID == temp {
							item.ID = actual
							m.local.Projects = replace(m.local.Projects, temp, item)
							continue
						}
					}
				}
			}
		}
		if len(m.client.Store.Projects) == 0 {
			err := m.client.Sync(context.Background())
			if err != nil {
				dbg(err)
			}
			err = client.WriteCache(m.client.Store, m.cmdQueue)
			if err != nil {
				dbg(err)
			}
			err = client.ReadCache(m.client.Store, m.local, m.cmdQueue)
			if err != nil {
				dbg(err)
			}
		} else {
			syncToken := m.client.Store.SyncToken
			incStore, err := m.client.IncrementalSync(m.ctx, syncToken)
			if err != nil {
				dbg(err)
				m.statusBarModel.SetSyncStatus(status.Error)
				return nil
			}
			m.client.Store.ApplyIncrementalSync(incStore)
			m.local.ApplyIncrementalSync(incStore)
			err = client.WriteCache(m.client.Store, m.cmdQueue)
			if err != nil {
				dbg(err)
				m.statusBarModel.SetSyncStatus(status.Error)
				m.sub <- struct{}{}
				return nil
			}
		}
		m.statusBarModel.SetSyncStatus(status.Synced)
		m.sub <- struct{}{}
		m.refresh()
		return nil
	}
}
