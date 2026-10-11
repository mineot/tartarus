package pages

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

type MenuMans struct {
	form      *huh.Form
	selected  string
	shortcuts Shortcuts
}

func NewMenuMans() *MenuMans {
	selection := huh.NewSelect[string]().
		Key("action").
		Title("📑 Manuals").
		Options(
			huh.NewOption("list os manuals", "abc"),
		)

	keymap := huh.NewDefaultKeyMap()
	keymap.Select.Filter.SetKeys()

	form := NewForm(huh.NewGroup(selection)).WithKeyMap(keymap)

	return &MenuMans{
		form:      form,
		shortcuts: Shortcuts{MenuConfigsShortcut, QuitShortcut},
	}
}

func (m *MenuMans) Init() tea.Cmd {
	return m.form.Init()
}

func (m *MenuMans) Update(msg tea.Msg) (Page, tea.Cmd) {
	if cmd, handled := m.shortcuts.Handle(msg); handled {
		return m, cmd
	}

	var cmd tea.Cmd
	var completed bool

	m.form, cmd, completed = UpdateForm(m.form, msg)

	if completed {
		m.selected = m.form.GetString("action")
	}

	return m, cmd
}

func (m *MenuMans) View() string {
	content := m.form.View()

	if m.form.State == huh.StateCompleted {
		content = "Selected: " + m.selected
	}

	return content + "\n\n" + FormHelp(m.form, m.shortcuts)
}
