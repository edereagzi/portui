package tui

import "charm.land/lipgloss/v2"

func renderHelpOverlay(width, height int) string {
	content := `Keybindings

Key          Action
──────────────────────────────
j / ↓        Move down
k / ↑        Move up
Enter        Toggle detail panel
x            Kill process
/            Search/filter
r            Refresh
?            Toggle this help
Esc          Back/cancel
q / Ctrl+C   Quit`

	rendered := HelpStyle.Render(content)
	if width > 0 && height > 0 {
		rendered = lipgloss.NewStyle().
			Width(width).
			Height(height).
			Align(lipgloss.Center, lipgloss.Center).
			Render(rendered)
	}
	return rendered
}
