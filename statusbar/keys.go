package statusbar

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Quit        key.Binding
	KillProcess key.Binding
	Move        key.Binding
	Filter      key.Binding
}

var Keys = KeyMap{
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	),
	KillProcess: key.NewBinding(
		key.WithKeys("ctrl+x"),
		key.WithHelp("ctrl+x", "kill process"),
	),
	Move: key.NewBinding(
		key.WithKeys("↑←↓→"),
		key.WithHelp("↑←↓→", "navigate"),
	),
	Filter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter"),
	),
}
