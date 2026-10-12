package pageforms

import (
	"tartarus/views/pages"
	pageutils "tartarus/views/pages/utils"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

type MenuStg struct {
	form      *huh.Form
	selected  string
	shortcuts pages.Shortcuts
	width     int
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

	form := pageutils.NewForm(huh.NewGroup(selection)).WithKeyMap(keymap)

	return &MenuStg{
		form:      form,
		shortcuts: pages.Shortcuts{pages.MenuConfigsShortcut, pages.QuitShortcut},
	}
}

func (m *MenuStg) Init() tea.Cmd {
	return m.form.Init()
}

func (m *MenuStg) Update(msg tea.Msg) (pages.Page, tea.Cmd) {
	if size, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = size.Width
	}

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

func (m *MenuStg) View() string {
	content := m.form.View()

	if m.form.State == huh.StateCompleted {
		content = "Selected: " + m.selected
	}

	return content + "\n\n" + pageutils.FormHelp(m.form, m.shortcuts, m.width)
}
