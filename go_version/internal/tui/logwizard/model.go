package logwizard

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"daily-tasks/go_version/internal/store"
	"daily-tasks/go_version/internal/tui/theme"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type step int

const (
	stepProjects step = iota
	stepTitle
	stepDescription
	stepHours
	stepSubtasks
	stepExtras
	stepDone
)

type checkItem struct {
	id      int64
	name    string
	desc    string
	value   string
	checked bool
}

func (i checkItem) Title() string       { return fmt.Sprintf("#%d · %s", i.id, i.name) }
func (i checkItem) Description() string { return strings.TrimSpace(i.desc) }
func (i checkItem) FilterValue() string { return i.value }

type projectItem struct {
	id   int64
	name string
	desc string
}

type model struct {
	db      *sql.DB
	logDate string

	step step

	projects     []projectItem
	projectList  list.Model
	selectedProj []projectItem
	projIndex    int

	titleInput textinput.Model
	descInput  textinput.Model
	hoursInput textinput.Model
	extrasInput textinput.Model

	subtaskItems []checkItem
	subtaskList  list.Model

	pendingTasks []store.TaskInput
	errMsg       string
	quitting     bool

	termWidth  int
	termHeight int
}

const defaultTermWidth = 80

func newStyledList(items []list.Item, title string, width, height int, withDescriptions bool) list.Model {
	d := newCheckboxDelegate(withDescriptions)
	d.styles.NormalTitle = theme.ListNormalTitle
	d.styles.SelectedTitle = theme.ListSelectedTitle
	d.styles.FilterMatch = theme.ListFilterMatch
	if width < 20 {
		width = defaultTermWidth
	}
	if height < 6 {
		height = 12
	}
	l := list.New(items, d, width, height)
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.TitleBar = theme.ListTitleBar
	l.Styles.PaginationStyle = theme.Help
	l.Styles.HelpStyle = theme.Help
	return l
}

func (m *model) listDimensions() (width, height int) {
	termW, termH := m.termWidth, m.termHeight
	if termW <= 0 {
		termW = defaultTermWidth
	}
	if termH <= 0 {
		termH = 24
	}
	width = termW - 6
	if width < 20 {
		width = 20
	}
	height = termH - 14
	if height < 8 {
		height = 8
	}
	return width, height
}

func (m *model) resizeLists() {
	w, h := m.listDimensions()
	m.projectList.SetSize(w, h)
	m.subtaskList.SetSize(w, h)
}

func styleTextInput(ti *textinput.Model) {
	ti.PromptStyle = theme.PromptStyle
	ti.TextStyle = theme.TextStyle
	ti.PlaceholderStyle = theme.PlaceholderStyle
	ti.Cursor.Style = theme.CursorStyle
}

func Run(db *sql.DB, logDate string) error {
	m, err := newModel(db, logDate)
	if err != nil {
		return err
	}
	p := tea.NewProgram(m, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return err
	}
	if fm, ok := final.(*model); ok && fm.errMsg != "" {
		return fmt.Errorf("%s", fm.errMsg)
	}
	return nil
}

func newModel(db *sql.DB, logDate string) (*model, error) {
	ctx := context.Background()
	rows, err := store.ListProjects(ctx, db)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("no projects yet; use: daily-tasks project add ...")
	}
	var items []projectItem
	var listItems []list.Item
	for _, p := range rows {
		pi := projectItem{id: p.ID, name: p.Name, desc: p.Description}
		items = append(items, pi)
		listItems = append(listItems, checkItem{
			id:      p.ID,
			name:    p.Name,
			desc:    p.Description,
			value:   fmt.Sprintf("%d", p.ID),
			checked: false,
		})
	}
	l := newStyledList(listItems, "◈ Select projects", defaultTermWidth, 12, true)

	ti := textinput.New()
	ti.Placeholder = "Ticket title"
	ti.Focus()
	ti.CharLimit = 500
	styleTextInput(&ti)

	di := textinput.New()
	di.Placeholder = "Ticket description"
	di.CharLimit = 4000
	styleTextInput(&di)

	hi := textinput.New()
	hi.Placeholder = "Hours (e.g. 0.5, 3)"
	hi.CharLimit = 16
	styleTextInput(&hi)

	ei := textinput.New()
	ei.Placeholder = "Extra subtasks (comma or newline)"
	ei.CharLimit = 4000
	styleTextInput(&ei)

	defaultTitle := fmt.Sprintf("Tareas diarias %s", logDate)
	ti.SetValue(defaultTitle)

	descDefault, err := store.DefaultTicketDescription(ctx, db, logDate)
	if err != nil {
		return nil, err
	}
	di.SetValue(descDefault)

	emptySubtasks := newStyledList([]list.Item{}, "◈ Subtasks", defaultTermWidth, 12, false)

	m := &model{
		db:           db,
		logDate:      logDate,
		step:         stepProjects,
		projects:     items,
		projectList:  l,
		subtaskList:  emptySubtasks,
		titleInput:   ti,
		descInput:    di,
		hoursInput:   hi,
		extrasInput:  ei,
		termWidth:    defaultTermWidth,
		termHeight:   24,
	}
	m.titleInput.Width = defaultTermWidth - 8
	m.descInput.Width = defaultTermWidth - 8
	m.hoursInput.Width = defaultTermWidth - 8
	m.extrasInput.Width = defaultTermWidth - 8
	return m, nil
}

func (m *model) Init() tea.Cmd {
	return textinput.Blink
}

func (m *model) activeInput() *textinput.Model {
	switch m.step {
	case stepTitle:
		return &m.titleInput
	case stepDescription:
		return &m.descInput
	case stepHours:
		return &m.hoursInput
	case stepExtras:
		return &m.extrasInput
	default:
		return nil
	}
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.termHeight = msg.Height
		m.resizeLists()
		m.titleInput.Width = msg.Width - 8
		m.descInput.Width = msg.Width - 8
		m.hoursInput.Width = msg.Width - 8
		m.extrasInput.Width = msg.Width - 8
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		}
	}

	switch m.step {
	case stepProjects:
		return m.updateProjects(msg)
	case stepTitle:
		return m.updateText(&m.titleInput, msg, stepDescription)
	case stepDescription:
		return m.updateText(&m.descInput, msg, stepHours)
	case stepHours:
		return m.updateHours(msg)
	case stepSubtasks:
		return m.updateSubtasks(msg)
	case stepExtras:
		return m.updateExtras(msg)
	case stepDone:
		if key, ok := msg.(tea.KeyMsg); ok {
			if key.String() == "q" || key.String() == "enter" {
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m *model) updateProjects(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case " ":
			idx := m.projectList.Index()
			if idx >= 0 && idx < len(m.projectList.Items()) {
				if it, ok := m.projectList.Items()[idx].(checkItem); ok {
					it.checked = !it.checked
					items := m.projectList.Items()
					items[idx] = it
					m.projectList.SetItems(items)
				}
			}
			return m, nil
		case "enter":
			var selected []projectItem
			for _, item := range m.projectList.Items() {
				ci, ok := item.(checkItem)
				if !ok || !ci.checked {
					continue
				}
				id := parseID(ci.value)
				for _, p := range m.projects {
					if p.id == id {
						selected = append(selected, p)
						break
					}
				}
			}
			if len(selected) == 0 {
				m.errMsg = "select at least one project"
				return m, nil
			}
			m.selectedProj = selected
			m.projIndex = 0
			m.step = stepTitle
			m.titleInput.Focus()
			return m, textinput.Blink
		}
	}
	var cmd tea.Cmd
	m.projectList, cmd = m.projectList.Update(msg)
	return m, cmd
}

func (m *model) updateText(inp *textinput.Model, msg tea.Msg, next step) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok && key.String() == "enter" {
		if m.step == stepDescription {
			m.projIndex = 0
			m.step = stepHours
			m.prepareHoursStep()
			m.hoursInput.Focus()
			return m, textinput.Blink
		}
		m.step = next
		if next == stepDescription {
			m.descInput.Focus()
		}
		return m, textinput.Blink
	}
	var cmd tea.Cmd
	*inp, cmd = inp.Update(msg)
	return m, cmd
}

func (m *model) prepareHoursStep() {
	if m.projIndex < len(m.selectedProj) {
		p := m.selectedProj[m.projIndex]
		m.hoursInput.SetValue("")
		m.hoursInput.Placeholder = fmt.Sprintf("Hours on %s", p.name)
	}
}

func (m *model) updateHours(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok && key.String() == "enter" {
		hours, err := store.ParseHours(m.hoursInput.Value())
		if err != nil || hours < 0 {
			m.errMsg = "enter a valid non-negative number of hours"
			return m, nil
		}
		m.errMsg = ""
		_ = hours // stored in advanceSubtasksStep
		if err := m.loadSubtaskStep(); err != nil {
			m.errMsg = err.Error()
			return m, nil
		}
		m.step = stepSubtasks
		return m, nil
	}
	var cmd tea.Cmd
	m.hoursInput, cmd = m.hoursInput.Update(msg)
	return m, cmd
}

func (m *model) loadSubtaskStep() error {
	ctx := context.Background()
	p := m.selectedProj[m.projIndex]
	defaults, err := store.ListDefaultSubtaskTitles(ctx, m.db, p.id)
	if err != nil {
		return err
	}
	past, err := store.ListPastSubtaskNames(ctx, m.db, p.id, m.logDate)
	if err != nil {
		return err
	}
	ordered, defaultSet := store.MergeSubtaskRows(defaults, past)
	var items []list.Item
	for _, label := range ordered {
		items = append(items, checkItem{
			name:    label,
			value:   label,
			checked: defaultSet[label],
		}) // id 0 for subtasks
	}
	w, h := m.listDimensions()
	m.subtaskList = newStyledList(items, fmt.Sprintf("◈ Subtasks · #%d %s", p.id, p.name), w, h, false)
	m.subtaskItems = nil
	for _, it := range items {
		if ci, ok := it.(checkItem); ok {
			m.subtaskItems = append(m.subtaskItems, ci)
		}
	}
	return nil
}

func (m *model) updateSubtasks(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case " ":
			idx := m.subtaskList.Index()
			items := m.subtaskList.Items()
			if idx >= 0 && idx < len(items) {
				if it, ok := items[idx].(checkItem); ok {
					it.checked = !it.checked
					items[idx] = it
					m.subtaskList.SetItems(items)
				}
			}
			return m, nil
		case "enter":
			m.step = stepExtras
			m.extrasInput.SetValue("")
			m.extrasInput.Focus()
			return m, textinput.Blink
		}
	}
	var cmd tea.Cmd
	m.subtaskList, cmd = m.subtaskList.Update(msg)
	return m, cmd
}

func (m *model) updateExtras(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok && key.String() == "enter" {
		hours, err := store.ParseHours(m.hoursInput.Value())
		if err != nil {
			m.errMsg = err.Error()
			return m, nil
		}
		p := m.selectedProj[m.projIndex]
		var names []string
		seen := make(map[string]bool)
		for _, it := range m.subtaskList.Items() {
			if ci, ok := it.(checkItem); ok && ci.checked {
				if !seen[ci.value] {
					seen[ci.value] = true
					names = append(names, ci.value)
				}
			}
		}
		for _, extra := range store.ParseSubtaskLines(m.extrasInput.Value()) {
			if !seen[extra] {
				seen[extra] = true
				names = append(names, extra)
			}
		}
		m.pendingTasks = append(m.pendingTasks, store.TaskInput{
			ProjectID:       p.id,
			Hours:           hours,
			TaskDescription: strings.TrimSpace(p.desc),
			SubtaskNames:    names,
		})
		m.projIndex++
		if m.projIndex < len(m.selectedProj) {
			m.step = stepHours
			m.prepareHoursStep()
			m.hoursInput.Focus()
			return m, textinput.Blink
		}
		// save
		ctx := context.Background()
		if err := store.SaveDailyTicket(ctx, m.db, m.logDate,
			m.titleInput.Value(), m.descInput.Value(), m.pendingTasks); err != nil {
			m.errMsg = err.Error()
			return m, nil
		}
		m.step = stepDone
		return m, nil
	}
	var cmd tea.Cmd
	m.extrasInput, cmd = m.extrasInput.Update(msg)
	return m, cmd
}

func (m *model) View() string {
	if m.quitting {
		return theme.Help.Render("Cancelled.") + "\n"
	}
	var body strings.Builder
	body.WriteString(theme.Header.Render(" daily-tasks "))
	body.WriteString(" ")
	body.WriteString(theme.Subtitle.Render(fmt.Sprintf("log · %s", m.logDate)))
	body.WriteString("\n\n")

	if m.errMsg != "" {
		body.WriteString(theme.Error.Render("✗ " + m.errMsg))
		body.WriteString("\n\n")
	}

	switch m.step {
	case stepProjects:
		body.WriteString(m.projectList.View())
		body.WriteString("\n")
		body.WriteString(theme.Help.Render("space toggle · enter continue · esc quit"))
	case stepTitle:
		body.WriteString(theme.Title.Render("Ticket title"))
		body.WriteString("\n")
		body.WriteString(m.titleInput.View())
		body.WriteString("\n")
		body.WriteString(theme.Help.Render("enter continue · esc quit"))
	case stepDescription:
		body.WriteString(theme.Title.Render("Ticket description"))
		body.WriteString("\n")
		body.WriteString(m.descInput.View())
		body.WriteString("\n")
		body.WriteString(theme.Help.Render("enter continue · esc quit"))
	case stepHours:
		p := m.selectedProj[m.projIndex]
		body.WriteString(theme.Title.Render(fmt.Sprintf("Hours · #%d %s", p.id, p.name)))
		body.WriteString(" ")
		body.WriteString(theme.Hint.Render(fmt.Sprintf("(%d/%d)", m.projIndex+1, len(m.selectedProj))))
		body.WriteString("\n")
		body.WriteString(m.hoursInput.View())
		body.WriteString("\n")
		if strings.TrimSpace(p.desc) != "" {
			body.WriteString(theme.Subtitle.Render("↳ " + strings.TrimSpace(p.desc)))
			body.WriteString("\n")
		}
		body.WriteString(theme.Help.Render("enter continue · esc quit"))
	case stepSubtasks:
		body.WriteString(m.subtaskList.View())
		body.WriteString("\n")
		body.WriteString(theme.Help.Render("space toggle · enter continue · esc quit"))
	case stepExtras:
		body.WriteString(theme.Title.Render("Extra subtasks (today only)"))
		body.WriteString("\n")
		body.WriteString(m.extrasInput.View())
		body.WriteString("\n")
		body.WriteString(theme.Help.Render("enter continue (empty ok) · esc quit"))
	case stepDone:
		body.WriteString(theme.Success.Render("✓ Saved"))
		body.WriteString("\n")
		body.WriteString(theme.TextStyle.Render(fmt.Sprintf("Daily ticket for %s · %d task(s).", m.logDate, len(m.pendingTasks))))
		body.WriteString("\n")
		body.WriteString(theme.Help.Render("enter or q to exit"))
	}

	return theme.AppBorder.Render(body.String()) + "\n"
}

func parseID(s string) int64 {
	var id int64
	fmt.Sscanf(s, "%d", &id)
	return id
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
