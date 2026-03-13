package tui

import "charm.land/lipgloss/v2"

// Color palette — port/network themed
var (
	colorBlue    = lipgloss.Color("#4A9EFF")
	colorCyan    = lipgloss.Color("#00D4AA")
	colorRed     = lipgloss.Color("#FF5555")
	colorGray    = lipgloss.Color("#6C7086")
	colorDark    = lipgloss.Color("#1E1E2E")
	colorGreen   = lipgloss.Color("#50FA7B")
	colorWhite   = lipgloss.Color("#CDD6F4")
	colorSurface = lipgloss.Color("#313244")
)

// AppTitle is the styled application name shown in the header.
var AppTitle = lipgloss.NewStyle().
	Foreground(colorBlue).
	Bold(true).
	Render("portui")

var (
	TableHeaderStyle = lipgloss.NewStyle().
				Foreground(colorBlue).
				Bold(true).
				Padding(0, 1)

	TableRowStyle = lipgloss.NewStyle().
			Padding(0, 1)

	SelectedRowStyle = lipgloss.NewStyle().
				Foreground(colorDark).
				Background(colorCyan).
				Bold(true)

	StatusBarStyle = lipgloss.NewStyle().
			Foreground(colorGray).
			Background(colorSurface).
			Padding(0, 1)

	DetailPanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorBlue).
				Padding(0, 1)

	SearchInputStyle = lipgloss.NewStyle().
				Foreground(colorCyan).
				Border(lipgloss.NormalBorder(), false, false, true, false).
				BorderForeground(colorCyan).
				Padding(0, 1)

	HelpStyle = lipgloss.NewStyle().
			Foreground(colorGray).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorGray).
			Padding(1, 2)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(colorRed).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(colorGreen).
			Bold(true)

	EmptyStateStyle = lipgloss.NewStyle().
			Foreground(colorGray).
			Italic(true).
			Padding(2, 4)

	ConfirmDialogStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorRed).
				Padding(1, 2)

	MutedStyle = lipgloss.NewStyle().
			Foreground(colorGray)
)
