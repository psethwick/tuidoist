package main

/*
Add Task
Add new task to the bottom of list A
Add new task to the top of list ⇧ A
Save a newly created task and create a new one below Enter
Save changes to an existing task and create a new task below ⇧ Enter
Save a new task or save changes to an existing one and create a new task above Ctrl Enter
*/

/*
Projects
Add section S
Share project ⇧ S
View as list/board ⇧ V
Sort by date D
Sort by priority P
Sort alphabetically N
Sort by assignee R
Open more project actions W
*/

import (
    "github.com/charmbracelet/bubbles/key"
)


/*
Navigate
Search /
Go to home G then H or H
Go to Inbox G then i
Go to Today G then T
Go to Upcoming G then U
Open project… G then P
Open label… G then L
*/
type globalKeyMap struct {
    Quit key.Binding
}

var globalKeys = globalKeyMap {
    Quit: key.NewBinding(key.WithKeys("ctrl+c", "ctrl+w")),
}

type gKeyMap struct {
    Home key.Binding
    Inbox key.Binding
    Today key.Binding
    Project key.Binding
    Filter key.Binding
}

var gKeys = gKeyMap {
    Home: key.NewBinding(key.WithKeys("h")),
    Inbox: key.NewBinding(key.WithKeys("i")),
    Today: key.NewBinding(key.WithKeys("t")),
    Project: key.NewBinding(key.WithKeys("p")),
    Filter: key.NewBinding(key.WithKeys("f")),
}

/*
General
Open task view Enter
Dismiss/Cancel Esc
Undo Z or Ctrl Z
Open command menu Ctrl K
Show keyboard shortcuts ?
Open/close sidebar menu M
*/

type taskListKeyMap struct {
    Open key.Binding
    Add key.Binding
    Undo key.Binding
}

var taskListKeys = taskListKeyMap {
    Open: key.NewBinding(key.WithKeys("enter")),
    Add: key.NewBinding(key.WithKeys("a")),
    Undo: key.NewBinding(key.WithKeys("z")),
}

/*
Edit Task
Edit task Ctrl E
Complete focused task E
Comment on task C
Set due date … T
Remove due date ⇧ T
Set priority … Y
Assign to… ⇧ R
Change labels L
Move to … V
Delete task permanently… ⇧ Delete
*/

