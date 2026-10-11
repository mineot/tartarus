package pages

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

type MenuCmds struct {
	form      *huh.Form
	selected  string
	shortcuts Shortcuts
}

func NewMenuCmds() *MenuCmds {
	selection := huh.NewSelect[string]().
		Key("action").
		Title("🖥️ Commands").
		Options(
			huh.NewOption("list of commands", "abc"),
		)

	keymap := huh.NewDefaultKeyMap()
	keymap.Select.Filter.SetKeys()
	keymap.Select.Submit.SetKeys()

	form := NewForm(huh.NewGroup(selection)).WithKeyMap(keymap)

	return &MenuCmds{
		form:      form,
		shortcuts: Shortcuts{MenuConfigsShortcut, QuitShortcut},
	}
}

func (m *MenuCmds) Init() tea.Cmd {
	return m.form.Init()
}

func (m *MenuCmds) Update(msg tea.Msg) (Page, tea.Cmd) {
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

func (m *MenuCmds) View() string {
	content := m.form.View()

	if m.form.State == huh.StateCompleted {
		content = "Selected: " + m.selected
	}

	return content + "\n\n" + FormHelp(m.form, m.shortcuts)
}
