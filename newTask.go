package main

import (
	"github.com/charmbracelet/bubbles/textinput"
)

type newTaskModel struct {
    input textinput.Model
}

func newNewTaskModel() newTaskModel {
    return newTaskModel{
        textinput.New(),
    }
}
