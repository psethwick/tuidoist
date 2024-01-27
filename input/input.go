package input

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/psethwick/tuidoist/keys"
)

type InputModel struct {
	Width    int
	show     func()
	hide     func()
	content  textinput.Model
	onAccept func(string) tea.Cmd
	repeat   bool
	keyMap   keys.InputKeyMap
}

func New(show func(), hide func(), km keys.InputKeyMap) InputModel {
	ti := textinput.New()
	ti.Prompt = ""
	return InputModel{
		Width:    20, // TODO
		show:     show,
		hide:     hide,
		content:  ti,
		onAccept: func(c string) tea.Cmd { return nil },
		keyMap:   km,
	}
}

func (im *InputModel) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, im.keyMap.Accept):
			cmds = append(cmds, im.onAccept(im.content.Value()))
			im.content.SetValue("")
			if !im.repeat {
				im.content.Prompt = ""
				im.content.Blur()
				im.hide()
			}
		case key.Matches(msg, im.keyMap.Cancel):
			im.content.SetValue("")
			im.content.Prompt = ""
			im.content.Blur()
			im.hide()
		}
	}
	input, cmd := im.content.Update(msg)
	im.content = input
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

func (im *InputModel) GetRepeat(prompt string, initialValue string, onAccept func(string) tea.Cmd) {
	im.GetOnce(prompt, initialValue, onAccept)
	im.repeat = true
}

func (im *InputModel) GetOnce(prompt string, initialValue string, onAccept func(string) tea.Cmd) {
	im.content.Prompt = fmt.Sprintf("%s ", prompt)
	im.content.SetValue(initialValue)
	im.onAccept = onAccept
	im.repeat = false
	im.show()
	im.content.Focus()
}

func (im *InputModel) View() string {
	return im.content.View()
}
