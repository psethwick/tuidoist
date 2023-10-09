package input

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/psethwick/tuidoist/style"
)

type InputModel struct {
	Width    int
	show     func()
	hide     func()
	content  textinput.Model
	onAccept func(string) tea.Cmd
	onExit   func()
	repeat   bool
}

func New(show func(), hide func()) InputModel {
	ti := textinput.New()
	return InputModel{
		Width:    20, // todo??
		show:     show,
		hide:     hide,
		content:  ti,
		onAccept: func(c string) tea.Cmd { return nil },
	}
}

func (im *InputModel) Height() int {
	return 4 // height of textinput + dialog
}

func (im *InputModel) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			cmds = append(cmds, im.onAccept(im.content.Value()))
			im.content.SetValue("")
			if !im.repeat {
				im.content.Blur()
				im.hide()
			}
		case "esc":
			im.content.SetValue("")
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
	im.content.Prompt = fmt.Sprintf("%s > ", prompt)
	im.content.SetValue(initialValue)
	im.onAccept = onAccept
	im.repeat = false
	im.show()
	im.content.Focus()
}

func (im *InputModel) View() string {
	return lipgloss.Place(im.Width, 3,
		lipgloss.Left, lipgloss.Left,
		style.DialogBox.Render(im.content.View()),
	)
}
