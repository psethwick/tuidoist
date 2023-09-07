package tasklist

import (
	"fmt"

	"github.com/muesli/termenv"
	"github.com/psethwick/bubblelister"
)

type TaskList struct {
	list   bubblelister.Model
	logger func(...any)
}

// TODO make the interface to this package _my_ types
func (tl *TaskList) ResetItems(s ...fmt.Stringer) {
	i, _ := tl.list.GetCursorIndex()
	tl.list.ResetItems(s...)
	tl.list.SetCursor(i)
}

func (tl *TaskList) Top() {
	tl.list.Top()
}

func (tl *TaskList) Bottom() {
	tl.list.Bottom()
}

// todo needed?
func (tl *TaskList) Sort() {
	tl.list.Sort()
}

// todo needed?
func (tl *TaskList) GetAllItems() []fmt.Stringer {
	return tl.list.GetAllItems()
}

// todo needed?
func (tl *TaskList) GetCursorIndex() (int, error) {
	return tl.list.GetCursorIndex()
}

// todo needed?
func (tl *TaskList) UpdateItem(idx int, updater func(fmt.Stringer) (fmt.Stringer, error)) {
	tl.list.UpdateItem(idx, updater)
}

// todo needed?
func (tl *TaskList) GetItem(i int) (fmt.Stringer, error) {
	return tl.list.GetItem(i)
}

func (tl *TaskList) AddItems(items ...fmt.Stringer) {
	tl.list.AddItems(items...)
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

func (tl *TaskList) GetCursorItem() (fmt.Stringer, error) {
	return tl.list.GetCursorItem()
}

func (tl *TaskList) RemoveCurrentItem() (fmt.Stringer, error) {
	idx, err := tl.list.GetCursorIndex()

	tl.logger("idx", idx)
	if err != nil {
		return nil, err
	}
	tl.logger("cound", len(tl.list.GetAllItems()))
	str, err := tl.list.RemoveIndex(idx)
	if err != nil {
		return nil, err
	}
	return str, nil
}

func New(lessFunc func(fmt.Stringer, fmt.Stringer) bool, logger func(...any)) TaskList {

	bl := bubblelister.NewModel()
	bl.LessFunc = lessFunc
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
	}
}

func (tl *TaskList) View() string {
	return tl.list.View()
}
