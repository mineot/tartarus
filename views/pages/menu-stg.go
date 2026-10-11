package pages

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

type MenuStg struct {
	form      *huh.Form
	selected  string
	shortcuts Shortcuts
}

func NewMenuStg() *MenuStg {
	selection := huh.NewSelect[string]().
		Key("action").
		Title("💾 Storage").
		Options(
			huh.NewOption("storages actioins", "cmds"),
		)

	keymap := huh.NewDefaultKeyMap()
	keymap.Select.Filter.SetKeys()

	form := NewForm(huh.NewGroup(selection)).WithKeyMap(keymap)

	return &MenuStg{
		form:      form,
		shortcuts: Shortcuts{MenuConfigsShortcut, QuitShortcut},
	}
}

func (m *MenuStg) Init() tea.Cmd {
	return m.form.Init()
}

func (m *MenuStg) Update(msg tea.Msg) (Page, tea.Cmd) {
	if cmd, handled := m.shortcuts.Handle(msg, FormEditing(m.form)); handled {
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

func (m *MenuStg) View() string {
	content := m.form.View()

	if m.form.State == huh.StateCompleted {
		content = "Selected: " + m.selected
	}

	return content + "\n\n" + FormHelp(m.form, m.shortcuts)
}
