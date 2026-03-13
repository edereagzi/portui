package tui

import "charm.land/bubbles/v2/key"

// keyMap defines all keyboard shortcuts for portui.
type keyMap struct {
	Up      key.Binding
	Down    key.Binding
	Enter   key.Binding
	Kill    key.Binding
	Search  key.Binding
	Esc     key.Binding
	Quit    key.Binding
	Help    key.Binding
	Refresh key.Binding
}

// Keys is the application-wide keymap.
var Keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "toggle detail panel"),
	),
	Kill: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "kill process"),
	),
	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search/filter"),
	),
	Esc: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back/cancel"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
}

// ShortHelp returns the condensed help text.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Kill, k.Search, k.Help, k.Quit}
}

// FullHelp returns the complete help text.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter},
		{k.Kill, k.Search, k.Refresh},
		{k.Help, k.Esc, k.Quit},
	}
}
