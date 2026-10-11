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

type Shortcut struct {
	Binding key.Binding
	Command tea.Cmd
}

type Shortcuts []Shortcut

// Handle leaves page shortcuts to the focused field while editing.
func (s Shortcuts) Handle(msg tea.Msg, editing bool) (tea.Cmd, bool) {
	if editing {
		return nil, false
	}

	pressed, ok := msg.(tea.KeyPressMsg)

	if !ok {
		return nil, false
	}

	for _, shortcut := range s {
		if key.Matches(pressed, shortcut.Binding) {
			return shortcut.Command, true
		}
	}

	return nil, false
}

func (s Shortcuts) Bindings() []key.Binding {
	bindings := make([]key.Binding, 0, len(s))

	for _, shortcut := range s {
		bindings = append(bindings, shortcut.Binding)
	}

	return bindings
}

type Page interface {
	Init() tea.Cmd
	Update(tea.Msg) (Page, tea.Cmd)
	View() string
}

type PageID uint

const (
	PageMenuMain PageID = iota
	PageMenuConfig
	PageMenuCmds
	PageMenuMans
	PageMenuStg
)

type NavigateMsg struct {
	To PageID
}
