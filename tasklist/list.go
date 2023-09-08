package tasklist

import (
	"fmt"

	"github.com/muesli/termenv"
	"github.com/psethwick/bubblelister"
	"github.com/psethwick/tuidoist/task"
)

type TaskList struct {
	list   bubblelister.Model
	logger func(...any)
	sort   TaskSort
}

type TaskSort uint

const (
	DefaultSort TaskSort = iota
	NameSort
	// PrioritySort
	// DateSort
	// AssigneeSort
)

type lessFunc = func(fmt.Stringer, fmt.Stringer) bool

var sortLessFunc = map[TaskSort]lessFunc{
	DefaultSort: childOrderLess,
	NameSort:    nameLess,

	// todo
	// PrioritySort: childOrderLess,
	// DateSort:     childOrderLess,
	// AssigneeSort: childOrderLess,
}

var sortDesc = map[TaskSort]string{
	DefaultSort: "",
	NameSort:    "Name",

	// todo
	// PrioritySort: childOrderLess,
	// DateSort:     childOrderLess,
	// AssigneeSort: childOrderLess,
}

func childOrderLess(a fmt.Stringer, b fmt.Stringer) bool {
	return a.(task.Task).Item.ChildOrder < b.(task.Task).Item.ChildOrder
}

func nameLess(a fmt.Stringer, b fmt.Stringer) bool {
	return a.(task.Task).Item.Content < b.(task.Task).Item.Content
}

func convertIn(tasks []task.Task) []fmt.Stringer {
	strs := make([]fmt.Stringer, len(tasks))
	for i, t := range tasks {
		strs[i] = t
	}
	return strs
}

func convertOut(strs []fmt.Stringer) []task.Task {
	tasks := make([]task.Task, len(strs))
	for i, t := range tasks {
		tasks[i] = t
	}
	return tasks
}

func (tl *TaskList) ResetItems(tasks []task.Task) {
	i, _ := tl.list.GetCursorIndex()
	tl.list.ResetItems(convertIn(tasks)...)
	tl.list.SetCursor(i)
}

func (tl *TaskList) Top() {
	tl.list.Top()
}

func updateTask(t task.Task) func(fmt.Stringer) (fmt.Stringer, error) {
	return func(fmt.Stringer) (fmt.Stringer, error) {
		return t, nil
	}
}

func (tl *TaskList) UpdateCurrentTask(t task.Task) {
	idx, _ := tl.list.GetCursorIndex()
	tl.list.UpdateItem(idx, updateTask(t))
}

func (tl *TaskList) Bottom() {
	tl.list.Bottom()
}

func (tl *TaskList) Sort(ts TaskSort) string {
	if ts == tl.sort {
		tl.sort = DefaultSort
	} else {
		tl.sort = ts
	}
	tl.list.LessFunc = sortLessFunc[tl.sort]
	tl.list.Sort()
	return sortDesc[tl.sort]
}

// todo needed?
func (tl *TaskList) GetAllItems() []task.Task {
	return convertOut(tl.list.GetAllItems())
}

func (tl *TaskList) AddItems(items ...fmt.Stringer) {
	tl.list.AddItems(items...)
	tl.Sort(tl.sort)
}

func (tl *TaskList) SetHeight(h int) {
	tl.list.Height = h
}

func (tl *TaskList) SetWidth(w int) {
	tl.list.Width = w
}

func (tl *TaskList) MoveCursor(i int) {
	tl.list.MoveCursor(i)
}

func (tl *TaskList) SetLessFunc(lf func(fmt.Stringer, fmt.Stringer) bool) {
	tl.list.LessFunc = lf
}

func (tl *TaskList) GetCursorItem() (task.Task, error) {
	str, err := tl.list.GetCursorItem()
	return str.(task.Task), err
}

func (tl *TaskList) RemoveCurrentItem() (task.Task, error) {
	idx, err := tl.list.GetCursorIndex()
	if err != nil {
		return task.Task{}, err
	}
	str, err := tl.list.RemoveIndex(idx)
	if err != nil {
		return task.Task{}, err
	}
	return str.(task.Task), nil
}

func equals(a fmt.Stringer, b fmt.Stringer) bool {
	// todo this might explode in a few ways
	// guess we'll find out
	return a.(task.Task).Item.ID == b.(task.Task).Item.ID
}

func New(logger func(...any)) TaskList {

	bl := bubblelister.NewModel()
	bl.LessFunc = childOrderLess
	bl.EqualsFunc = equals
	pfxr := bubblelister.NewPrefixer()
	pfxr.Number = false
	pfxr.NumberRelative = false
	bl.PrefixGen = pfxr

	p := termenv.ColorProfile()
	// todo maybe fork bubblelister to use lipgloss?
	// adaptive color would probably be the motivation
	bl.CurrentStyle = termenv.Style{}.Foreground(p.Color("#F793FF"))

	return TaskList{
		list:   bl,
		logger: logger,
		sort:   DefaultSort,
	}
}

func (tl *TaskList) View() string {
	return tl.list.View()
}
