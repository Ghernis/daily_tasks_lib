package theme

import "github.com/charmbracelet/lipgloss"

// Tokyo Night Neon-inspired palette.
const (
	ColorBG      = "#1a1b26"
	ColorSurface = "#24283b"
	ColorFG      = "#c0caf5"
	ColorCyan    = "#7dcfff"
	ColorBlue    = "#7aa2f7"
	ColorMagenta = "#bb9af7"
	ColorPink    = "#f7768e"
	ColorGreen   = "#9ece6a"
	ColorYellow  = "#e0af68"
	ColorMuted   = "#565f89"
	ColorBorder  = "#3b4261"
)

var (
	AppBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColorMagenta)).
			Padding(0, 1)

	Header = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorCyan)).
		Background(lipgloss.Color(ColorSurface)).
		Padding(0, 1).
		MarginBottom(1)

	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorMagenta))

	Subtitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorBlue))

	Help = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorMuted)).
		Italic(true)

	Error = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorPink))

	Success = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorGreen))

	Hint = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorYellow))

	CheckboxOn = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorGreen)).
			Bold(true)

	CheckboxOff = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorMuted))

	ListTitleBar = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(ColorCyan)).
			Background(lipgloss.Color(ColorSurface)).
			Padding(0, 1)

	ListNormalTitle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorFG))

	ListSelectedTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(ColorFG)).
			Background(lipgloss.Color(ColorMagenta)).
			PaddingLeft(1)

	ListFilterMatch = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorYellow))

	PromptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorCyan))

	TextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorFG))

	PlaceholderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorMuted))

	CursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorPink))
)
