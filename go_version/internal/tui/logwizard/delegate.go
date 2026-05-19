package logwizard

import (
	"fmt"
	"io"
	"strings"

	"daily-tasks/go_version/internal/tui/theme"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type checkboxDelegate struct {
	styles       list.DefaultItemStyles
	showDesc     bool
	descSelected bool // show description for all items when true; else only selected row
}

func newCheckboxDelegate(showDesc bool) checkboxDelegate {
	d := list.NewDefaultDelegate()
	return checkboxDelegate{styles: d.Styles, showDesc: showDesc, descSelected: true}
}

func (d checkboxDelegate) Height() int {
	if d.showDesc {
		return 2
	}
	return 1
}

func (d checkboxDelegate) Spacing() int { return 0 }

func (d checkboxDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d checkboxDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	ci, ok := item.(checkItem)
	if !ok {
		return
	}
	mark := theme.CheckboxOff.Render("[ ]")
	if ci.checked {
		mark = theme.CheckboxOn.Render("[x]")
	}
	title := fmt.Sprintf("%s #%d · %s", mark, ci.id, ci.name)
	selected := index == m.Index()
	style := d.styles.NormalTitle
	if selected {
		style = d.styles.SelectedTitle
	}
	fmt.Fprint(w, style.Render(title))
	desc := strings.TrimSpace(ci.desc)
	if d.showDesc && desc != "" && (selected || !d.descSelected) {
		fmt.Fprint(w, "\n")
		if selected {
			fmt.Fprint(w, "  "+theme.TextStyle.Render(desc))
		} else {
			fmt.Fprint(w, "  "+theme.Subtitle.Render(desc))
		}
	}
}
