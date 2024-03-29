package tasklist

import (
	"fmt"
	"math"

	"github.com/psethwick/tuidoist/bubblelister"
	"github.com/psethwick/tuidoist/style"
	"github.com/psethwick/tuidoist/task"
)

type listModel struct {
	bubblelister.Model
	title  string
	listId interface{}
}

type TaskList struct {
	OnTitleChange func(string)
	lists         []*listModel
	logger        func(...any)
	sort          TaskSort
	idx           int
	height        int
	width         int
}

type TaskSort uint

// map[item.ID]*task.Task
var selected map[string]*task.Task = map[string]*task.Task{}

var selectedParentLevel *string = nil

func compareParentId(p1, p2 *string) bool {
	return (p1 == nil && p2 == nil) || (p1 != nil && p2 != nil && *p1 == *p2)
}

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
		at := a.(task.Task)
		bt := b.(task.Task)
		return at.SortKey < bt.SortKey
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
		if _, ok := selected[t.Item.ID]; ok {
			t.Selected = true
		}
		strs[i] = t
	}
	return strs
}

func (tl *TaskList) Select() {
	ci, err := tl.lists[tl.idx].GetCursorIndex()
	if err != nil {
		return
	}
	str, err := tl.lists[tl.idx].GetItem(ci)
	if err != nil {
		return
	}
	task := str.(task.Task)

	tparent := task.Item.ParentID

	if _, ok := selected[task.Item.ID]; ok {
		task.Selected = false
		delete(selected, task.Item.ID)
		if len(selected) == 0 {
			selectedParentLevel = nil
		}
	} else {
		if len(selected) > 0 && !compareParentId(tparent, selectedParentLevel) {
			return
		}
		selectedParentLevel = tparent
		task.Selected = true
		selected[task.Item.ID] = &task
	}
	_ = tl.lists[tl.idx].UpdateItem(ci, updateTask(task))
}

func (tl *TaskList) selectItems(clear bool) []task.Task {
	var tasks []task.Task
	for _, t := range selected {
		tasks = append(tasks, *t)
	}
	if clear {
		tl.Unselect()
	}
	if len(tasks) == 0 {
		itm, err := tl.lists[tl.idx].GetCursorItem()
		if err != nil {
			return nil
		}
		tasks = append(tasks, itm.(task.Task))
	}
	return tasks
}

func (tl *TaskList) SelectedItems() []task.Task {
	return tl.selectItems(false)
}

func (tl *TaskList) SelectedItemsClear() []task.Task {
	return tl.selectItems(true)
}

func (tl *TaskList) Unselect() {
	selected = map[string]*task.Task{}
	selectedParentLevel = nil
	for idx, l := range tl.lists {
		for i, t := range l.GetAllItems() {
			task := t.(task.Task)
			if task.Selected {
				task.Selected = false
				_ = tl.lists[idx].UpdateItem(i, updateTask(task))
			}
		}
	}
}

type List struct {
	Title  string
	Tasks  []task.Task
	ListId interface{}
}

func (tl *TaskList) ResetItems(lists []List, newIdx int) {
	newLists := make([]*listModel, len(lists))
	for i, l := range lists {
		for _, ol := range tl.lists { // linear search on the order of 1
			if ol.listId == l.ListId {
				// retain state of list model if we already have it
				// we don't want to move cursor or offset if we don't have to
				newLists[i] = ol
			}
		}
		if newLists[i] == nil { // didn't find list or we're on a different set of lists
			newLists[i] = &listModel{tl.newList(), l.Title, l.ListId}
		}
		if err := newLists[i].ResetItems(convertIn(l.Tasks)...); err != nil {
			tl.logger(err)
		}
	}
	tl.lists = newLists
	tl.idx = newIdx
	tl.OnTitleChange(tl.Title())
}

func (tl *TaskList) Title() string {
	t := ""
	if tl.idx != 0 {
		t += "<-"
	}
	t += tl.lists[tl.idx].title
	if tl.idx != len(tl.lists)-1 {
		t += "->"
	}
	return t
}

func (tl *TaskList) Len() int {
	return tl.lists[tl.idx].Len()
}

func (tl *TaskList) Top() {
	_ = tl.lists[tl.idx].Top()
}

func (tl *TaskList) HalfPageUp() {
	_, _ = tl.lists[tl.idx].MoveCursor(-5)
}

func (tl *TaskList) SendToTop() []map[string]interface{} {
	lessThan := math.MaxInt
	for _, strangers := range tl.lists[tl.idx].GetAllItems() {
		item := strangers.(task.Task).Item
		if compareParentId(item.ParentID, selectedParentLevel) {
			lessThan = min(item.ChildOrder, lessThan)
		}
	}
	changes := *new([]map[string]interface{})
	for i, t := range tl.SelectedItems() {
		changes = append(changes, map[string]interface{}{"id": t.Item.ID, "child_order": lessThan - 1 - i})
	}
	tl.Unselect()
	return changes
}

func (tl *TaskList) SendToBottom() []map[string]interface{} {
	moreThan := math.MinInt
	for _, strangers := range tl.lists[tl.idx].GetAllItems() {
		item := strangers.(task.Task).Item
		if compareParentId(item.ParentID, selectedParentLevel) {
			moreThan = max(item.ChildOrder, moreThan)
		}
	}
	changes := *new([]map[string]interface{})
	for i, t := range tl.SelectedItems() {
		changes = append(changes, map[string]interface{}{"id": t.Item.ID, "child_order": moreThan + 1 + i})
	}
	tl.Unselect()
	return changes
}

func (tl *TaskList) HalfPageDown() {
	_, _ = tl.lists[tl.idx].MoveCursor(5)
}

func (tl *TaskList) WholePageUp() {
	_, _ = tl.lists[tl.idx].MoveCursor(-10)
}
func (tl *TaskList) WholePageDown() {
	_, _ = tl.lists[tl.idx].MoveCursor(10)
}

func updateTask(t task.Task) func(fmt.Stringer) (fmt.Stringer, error) {
	return func(fmt.Stringer) (fmt.Stringer, error) {
		return t, nil
	}
}

func (tl *TaskList) Bottom() {
	_ = tl.lists[tl.idx].Bottom()
}

func (tl *TaskList) Sort(ts TaskSort) string {
	if ts == tl.sort {
		tl.sort = DefaultSort
	} else {
		tl.sort = ts
	}

	for i := range tl.lists {
		tl.lists[i].LessFunc = sortLessFunc[tl.sort]
		tl.lists[i].Sort()
	}
	return sortDesc[tl.sort]
}

func (tl *TaskList) SetHeight(h int) {
	tl.height = h
	for i := range tl.lists {
		tl.lists[i].Height = h
	}
}

func (tl *TaskList) SetWidth(w int) {
	tl.width = w
	for i := range tl.lists {
		tl.lists[i].Width = w
	}
}

func (tl *TaskList) MoveCursor(i int) {
	_, err := tl.lists[tl.idx].MoveCursor(i)
	if err != nil {
		tl.logger("")
	}
}

func (tl *TaskList) GetAboveItem() (task.Task, error) {
	ci, err := tl.lists[tl.idx].GetCursorIndex()
	var t task.Task
	if err != nil {
		return t, err
	}
	item, err := tl.lists[tl.idx].GetItem(ci - 1)
	if err != nil {
		return t, err
	}
	t = item.(task.Task)
	return t, nil
}

func (tl *TaskList) GetCursorItem() (task.Task, error) {
	str, err := tl.lists[tl.idx].GetCursorItem()
	if err != nil {
		return task.Task{}, err
	}
	return str.(task.Task), err

}

func (tl *TaskList) NextList() interface{} {
	tl.idx = min(tl.idx+1, len(tl.lists)-1)
	tl.OnTitleChange(tl.Title())
	tl.Unselect()
	return tl.lists[tl.idx].listId
}

func (tl *TaskList) PrevList() interface{} {
	tl.idx = max(0, tl.idx-1)
	tl.OnTitleChange(tl.Title())
	tl.Unselect()
	return tl.lists[tl.idx].listId
}

func equals(a fmt.Stringer, b fmt.Stringer) bool {
	return a.(task.Task).Item.ID == b.(task.Task).Item.ID
}

func (tl *TaskList) newList() bubblelister.Model {
	bl := bubblelister.NewModel()
	bl.LessFunc = sortLessFunc[tl.sort]
	bl.EqualsFunc = equals
	bl.PrefixGen = bubblelister.NewPrefixer()
	bl.Width = tl.width
	bl.Height = tl.height
	bl.Logger = tl.logger

	bl.CurrentStyle = style.SelectedTitle
	return bl
}

func New(onTitleChange func(string), logger func(...any)) TaskList {
	return TaskList{
		OnTitleChange: onTitleChange,
		lists:         []*listModel{},
		logger:        logger,
		sort:          DefaultSort,
	}
}

func (tl *TaskList) View() string {
	if len(tl.lists) > 0 {
		return tl.lists[tl.idx].View()
	}
	return ""
}
