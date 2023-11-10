package task

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/psethwick/tuidoist/style"

	todoist "github.com/sachaos/todoist/lib"
)

type Task struct {
	Item      todoist.Item
	Title     string
	Summary   string
	Completed bool
	Url       string
	Selected  bool
}

func (t Task) String() string {
	s := fmt.Sprintf("%s\n%s", t.Title, t.Summary)
	if t.Selected {
		return style.Selected.Render(s)
	}
	return s
}

func reformatDate(d string, from string, to string) string {
	// slicing d because _sometimes_ there's timezone info on the date
	// ain't nobody got time for that
	t, err := time.Parse(from, d[:len(from)])
	if err != nil {
		return err.Error()
	}
	return t.Format(to)
}

var mdUrlRegex = regexp.MustCompile(`\[([^\]]+)\]\((https?:\/\/[^\)]+)\)`)

// heres where reminders is
// todoist.Store.Reminders
// ⏰ ??
// todo overdue should be red somewhere
// today, maybe also highlighted?
func New(store *todoist.Store, item todoist.Item) Task {
	indent := strings.Repeat(" ", len(todoist.SearchItemParents(store, &item)))
	var checkbox string
	switch item.Priority {
	case 1:
		checkbox = " ⚪ "
	case 2:
		checkbox = " 🔵 "
	case 3:
		checkbox = " 🟡 "
	case 4:
		checkbox = " 🔴 "
	}

	if indent != "" {
		checkbox = fmt.Sprint("╰", checkbox)
	}
	labels := ""
	for _, l := range item.LabelNames {
		labels += fmt.Sprint(" 🏷️ ", l)
	}
	summary := ""
	if item.Due != nil {
		summary += " 🗓️ "
		var fd string
		if strings.Contains(item.Due.Date, "T") {
			fd = reformatDate(item.Due.Date, "2006-01-02T15:04:05", "02 Jan 06 15:04")
		} else {
			fd = reformatDate(item.Due.Date, "2006-01-02", "02 Jan 06")
		}
		if item.Due.IsRecurring {
			fd += " 🔁"
		}
		summary += fd
	}

	content := item.Content
	// todo this only handles one url
	// also doesn't handle bare urls
	urlMatch := mdUrlRegex.FindStringSubmatch(item.Content)
	url := ""
	if len(urlMatch) > 0 {
		content = style.Underline.Render(strings.Replace(content, urlMatch[0], urlMatch[1], 1))
		content += " 🔗"
		url = urlMatch[2]
	}
	title := fmt.Sprint(indent, checkbox, content, labels)
	summary = fmt.Sprintf("%s\n%s", fmt.Sprint(indent, "   ", summary), fmt.Sprint(indent, item.Description))

	return Task{
		Item:    item,
		Title:   title,
		Summary: summary,
		Url:     url,
	}
}
