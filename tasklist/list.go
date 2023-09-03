package tasklist

import (
	"fmt"

	"github.com/muesli/termenv"
	"github.com/treilik/bubblelister"
)

type TaskList struct {
	List   bubblelister.Model
	gMenu  bool
	logger func(...any)
}

func (tl *TaskList) RemoveCurrentItem() (fmt.Stringer, error) {
	idx, err := tl.List.GetCursorIndex()
	if err != nil {
		return nil, err
	}
	str, err := tl.List.RemoveIndex(idx)
	if err != nil {
		return nil, err
	}
	return str, nil
}


func New(lessFunc func(fmt.Stringer, fmt.Stringer) bool, logger func(...any)) TaskList {

	bl := bubblelister.NewModel()
	bl.LessFunc = lessFunc
	p := termenv.ColorProfile()
	// todo maybe fork bubblelister to use lipgloss?
	// adaptive color would probably be the motivation
	bl.CurrentStyle = termenv.Style{}.Foreground(p.Color("#F793FF"))

	return TaskList{
		List: bl,
	}
}

func (tl *TaskList) View() string {
	return tl.List.View()
}
