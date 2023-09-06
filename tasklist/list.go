package tasklist

import (
	"fmt"

	"github.com/muesli/termenv"
	"github.com/psethwick/bubblelister"
)

type TaskList struct {
	List   bubblelister.Model
	logger func(...any)
}

func (tl *TaskList) ResetItems(s ...fmt.Stringer) {
	i, _ := tl.List.GetCursorIndex()
	tl.List.ResetItems(s...)
	tl.List.SetCursor(i)
}

func (tl *TaskList) RemoveCurrentItem() (fmt.Stringer, error) {
	idx, err := tl.List.GetCursorIndex()

	tl.logger("idx", idx)
	if err != nil {
		return nil, err
	}
	tl.logger("cound", len(tl.List.GetAllItems()))
	str, err := tl.List.RemoveIndex(idx)
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
		List:   bl,
		logger: logger,
	}
}

func (tl *TaskList) View() string {
	return tl.List.View()
}
