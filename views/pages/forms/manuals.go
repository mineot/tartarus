package pageforms

import (
	"tartarus/views/pages"
	pageutils "tartarus/views/pages/utils"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

type MenuMans struct {
	form      *huh.Form
	selected  string
	shortcuts pages.Shortcuts
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

	form := pageutils.NewForm(huh.NewGroup(selection)).WithKeyMap(keymap)

	return &MenuMans{
		form:      form,
		shortcuts: pages.Shortcuts{pages.MenuConfigsShortcut, pages.QuitShortcut},
	}
}

func (m *MenuMans) Init() tea.Cmd {
	return m.form.Init()
}

func (m *MenuMans) Update(msg tea.Msg) (pages.Page, tea.Cmd) {
	if cmd, handled := m.shortcuts.Handle(msg, pageutils.FormEditing(m.form)); handled {
		return m, cmd
	}

	var cmd tea.Cmd
	var completed bool

	m.form, cmd, completed = pageutils.UpdateForm(m.form, msg)

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

	return content + "\n\n" + pageutils.FormHelp(m.form, m.shortcuts)
}
