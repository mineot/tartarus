package pages

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

var (
	QuitShortcut = Shortcut{
		Binding: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "quit"),
		),
		Command: tea.Quit,
	}

	MenuMainShortcut = Shortcut{
		Binding: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "main menu"),
		),
		Command: func() tea.Msg {
			return NavigateMsg{To: PageMenuMain}
		},
	}

	MenuConfigsShortcut = Shortcut{
		Binding: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "configs menu"),
		),
		Command: func() tea.Msg {
			return NavigateMsg{To: PageMenuConfig}
		},
	}
)
