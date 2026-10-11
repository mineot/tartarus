package pages

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
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
