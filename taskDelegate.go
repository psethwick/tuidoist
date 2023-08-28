package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/truncate"
)

type taskDelegate struct{}

func (td taskDelegate) Height() int {
	return 2
}

func (td taskDelegate) Spacing() int {
	return 1
}

func (td taskDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

// the only thing really different from DefaultDelegate is I turned off the filter rune highlight
func (td taskDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	var (
		title, desc string
		s           = list.NewDefaultItemStyles()
		ellipsis    = "…"
	)

	if i, ok := item.(task); ok {
		title = i.Title()
		desc = i.Description()
	} else {
		return
	}

	if m.Width() <= 0 {
		// short-circuit
		return
	}

	// Prevent text from exceeding list width
	textwidth := uint(m.Width() - s.NormalTitle.GetPaddingLeft() - s.NormalTitle.GetPaddingRight())
	title = truncate.StringWithTail(title, textwidth, ellipsis)
	var lines []string
	for i, line := range strings.Split(desc, "\n") {
		if i >= td.Height()-1 {
			break
		}
		lines = append(lines, truncate.StringWithTail(line, textwidth, ellipsis))
	}
	desc = strings.Join(lines, "\n")

	var (
		isSelected  = index == m.Index()
		emptyFilter = m.FilterState() == list.Filtering && m.FilterValue() == ""
	)

	if emptyFilter {
		title = s.DimmedTitle.Render(title)
		desc = s.DimmedDesc.Render(desc)
	} else if isSelected && m.FilterState() != list.Filtering {
		title = s.SelectedTitle.Render(title)
		desc = s.SelectedDesc.Render(desc)
	} else {
		title = s.NormalTitle.Render(title)
		desc = s.NormalDesc.Render(desc)
	}

	fmt.Fprintf(w, "%s\n%s", title, desc)
}
