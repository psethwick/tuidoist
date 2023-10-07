package tasklist

import (
	"fmt"

	"github.com/muesli/termenv"
	"github.com/psethwick/bubblelister"
	"github.com/psethwick/tuidoist/style"
	"github.com/psethwick/tuidoist/task"
)

type listModel struct {
	bubblelister.Model
	title string
}

type TaskList struct {
	OnTitleChange func(string)
	list          []listModel
	logger        func(...any)
	sort          TaskSort
	idx           int
	height        int
	width         int
}

type TaskSort uint

const (
	DefaultSort TaskSort = iota
	NameSort
	PrioritySort
	DateSort
	AssigneeSort
)

type lessFunc = func(fmt.Stringer, fmt.Stringer) bool

var sortLessFunc = map[TaskSort]lessFunc{
	DefaultSort: func(a fmt.Stringer, b fmt.Stringer) bool {
		return a.(task.Task).Item.ChildOrder < b.(task.Task).Item.ChildOrder
	},
	NameSort: func(a fmt.Stringer, b fmt.Stringer) bool {
		return a.(task.Task).Item.Content < b.(task.Task).Item.Content
	},
	PrioritySort: func(a fmt.Stringer, b fmt.Stringer) bool {
		return a.(task.Task).Item.Priority > b.(task.Task).Item.Priority
	},
	DateSort: func(a fmt.Stringer, b fmt.Stringer) bool {
		adue := a.(task.Task).Item.Due
		bdue := b.(task.Task).Item.Due
		if adue != nil && bdue != nil {
			return adue.Date < bdue.Date
		} else if adue != nil {
			return true
		}
		return false
	},
	AssigneeSort: func(a fmt.Stringer, b fmt.Stringer) bool {
		return true // lol
	},
}

var sortDesc = map[TaskSort]string{
	DefaultSort:  "",
	NameSort:     "name",
	PrioritySort: "priority",
	DateSort:     "date",
	AssigneeSort: "assignee",
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

type List struct {
	Title string
	Tasks []task.Task
}

func (tl *TaskList) ResetItems(lists []List, newIdx int) {
	ci := 0
	if len(tl.list) > 0 {
		ci, _ = tl.list[tl.idx].GetCursorIndex()
	}
	tl.list = make([]listModel, len(lists))
	for i, l := range lists {
		tl.list[i] = listModel{tl.newList(), l.Title}
		tl.list[i].ResetItems(convertIn(l.Tasks)...)
		tl.list[i].SetCursor(ci)
	}
	tl.idx = newIdx
	tl.OnTitleChange(tl.Title())
	for i, l := range tl.list {
		tl.logger(i, l.Len())
	}
}

func (tl *TaskList) Title() string {
	t := ""
	if tl.idx != 0 {
		t += "<-"
	}
	t += tl.list[tl.idx].title
	if tl.idx != len(tl.list)-1 {
		t += "->"
	}
	return t
}

func (tl *TaskList) Len() int {
	return tl.list[tl.idx].Len()
}

func (tl *TaskList) Top() {
	tl.list[tl.idx].Top()
}

func (tl *TaskList) HalfPageUp() {
	tl.list[tl.idx].MoveCursor(-5)
}
func (tl *TaskList) HalfPageDown() {
	tl.list[tl.idx].MoveCursor(5)
}

func (tl *TaskList) WholePageUp() {
	tl.list[tl.idx].MoveCursor(-10)
}
func (tl *TaskList) WholePageDown() {
	tl.list[tl.idx].MoveCursor(10)
}

func updateTask(t task.Task) func(fmt.Stringer) (fmt.Stringer, error) {
	return func(fmt.Stringer) (fmt.Stringer, error) {
		return t, nil
	}
}

func (tl *TaskList) UpdateCurrentTask(t task.Task) {
	idx, _ := tl.list[tl.idx].GetCursorIndex()
	tl.list[tl.idx].UpdateItem(idx, updateTask(t))
}

func (tl *TaskList) Bottom() {
	tl.list[tl.idx].Bottom()
}

func (tl *TaskList) Sort(ts TaskSort) string {
	if ts == tl.sort {
		tl.sort = DefaultSort
	} else {
		tl.sort = ts
	}

	for i, _ := range tl.list {
		tl.list[i].LessFunc = sortLessFunc[tl.sort]
		tl.list[i].Sort()
	}
	return sortDesc[tl.sort]
}

func (tl *TaskList) getAllItems() []task.Task {
	return convertOut(tl.list[tl.idx].GetAllItems())
}

func (tl *TaskList) AddItemTop(t task.Task) task.Task {
	minOrder := 0
	for _, t := range tl.getAllItems() {
		minOrder = min(minOrder, t.Item.ChildOrder)
	}
	t.Item.ChildOrder = minOrder - 1
	tl.list[tl.idx].AddItems(t)
	tl.list[tl.idx].Sort()
	tl.list[tl.idx].Top()
	return t
}

func (tl *TaskList) AddItem(t task.Task) {
	tl.list[tl.idx].AddItems(t)
	tl.Sort(tl.sort)
}

func (tl *TaskList) AddItemBottom(t task.Task) task.Task {
	maxOrder := 0
	for _, lt := range tl.list[tl.idx].GetAllItems() {
		maxOrder = max(maxOrder, lt.(task.Task).Item.ChildOrder)
	}
	t.Item.ChildOrder = maxOrder + 1
	tl.list[tl.idx].AddItems(t)
	tl.list[tl.idx].Sort()
	tl.list[tl.idx].Bottom()
	return t
}

func (tl *TaskList) SetHeight(h int) {
	tl.height = h
	for i, _ := range tl.list {
		tl.list[i].Height = h
	}
}

func (tl *TaskList) SetWidth(w int) {
	tl.width = w
	for i, _ := range tl.list {
		tl.list[i].Width = w
	}
}

func (tl *TaskList) MoveCursor(i int) {
	tl.list[tl.idx].MoveCursor(i)
}

func (tl *TaskList) GetCursorItem() (task.Task, error) {
	str, err := tl.list[tl.idx].GetCursorItem()
	if err != nil {
		return task.Task{}, err
	}
	return str.(task.Task), err

}

func (tl *TaskList) NextList() {
	tl.idx = min(tl.idx+1, len(tl.list)-1)
	tl.OnTitleChange(tl.Title())
}
func (tl *TaskList) PrevList() {
	tl.idx = max(0, tl.idx-1)
	tl.OnTitleChange(tl.Title())
}

func (tl *TaskList) RemoveCurrentItem() (task.Task, error) {
	idx, err := tl.list[tl.idx].GetCursorIndex()
	if err != nil {
		return task.Task{}, err
	}
	str, err := tl.list[tl.idx].RemoveIndex(idx)
	if err != nil {
		return task.Task{}, err
	}
	return str.(task.Task), nil
}

func equals(a fmt.Stringer, b fmt.Stringer) bool {
	return a.(task.Task).Item.ID == b.(task.Task).Item.ID
}

func (tl *TaskList) newList() bubblelister.Model {
	bl := bubblelister.NewModel()
	bl.LessFunc = sortLessFunc[tl.sort]
	bl.EqualsFunc = equals
	pfxr := bubblelister.NewPrefixer()
	pfxr.Number = false
	pfxr.NumberRelative = false
	bl.PrefixGen = pfxr
	bl.Width = tl.width
	bl.Height = tl.height

	p := termenv.ColorProfile()
	// todo maybe fork bubblelister to use lipgloss?
	// adaptive color would probably be the motivation
	bl.CurrentStyle = termenv.Style{}.Foreground(p.Color(style.Pink.Light))
	return bl
}

func New(onTitleChange func(string), logger func(...any)) TaskList {
	return TaskList{
		OnTitleChange: onTitleChange,
		list:          []listModel{},
		logger:        logger,
		sort:          DefaultSort,
	}
}

func (tl *TaskList) View() string {
	return tl.list[tl.idx].View()
}
