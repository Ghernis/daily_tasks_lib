package logwizard

import (
	"fmt"
	"io"

	"daily-tasks/go_version/internal/tui/theme"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type checkboxDelegate struct {
	styles list.DefaultItemStyles
}

func newCheckboxDelegate() checkboxDelegate {
	d := list.NewDefaultDelegate()
	return checkboxDelegate{styles: d.Styles}
}

func (d checkboxDelegate) Height() int                             { return 1 }
func (d checkboxDelegate) Spacing() int                            { return 0 }
func (d checkboxDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd    { return nil }

func (d checkboxDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	ci, ok := item.(checkItem)
	if !ok {
		return
	}
	mark := theme.CheckboxOff.Render("[ ]")
	if ci.checked {
		mark = theme.CheckboxOn.Render("[x]")
	}
	line := mark + " " + ci.name
	if index == m.Index() {
		fmt.Fprint(w, d.styles.SelectedTitle.Render(line))
		return
	}
	fmt.Fprint(w, d.styles.NormalTitle.Render(line))
}
