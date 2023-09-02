package tasklist

import (
	"github.com/muesli/termenv"
	"github.com/treilik/bubblelister"
)

type TaskList struct {
	List   bubblelister.Model
	gMenu  bool
	logger func(...any)
}

func New(logger func(...any)) TaskList {

	bl := bubblelister.NewModel()
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
