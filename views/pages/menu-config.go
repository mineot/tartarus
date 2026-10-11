package pages

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

type MenuConfig struct {
	form      *huh.Form
	selected  string
	shortcuts Shortcuts
}

func NewMenuConfig() *MenuConfig {
	selection := huh.NewSelect[string]().
		Key("action").
		Title("⚙️ Configurations").
		Options(
			huh.NewOption("🖥️ Commands", "cmds"),
			huh.NewOption("📑 Manuals", "mans"),
			huh.NewOption("💾 Storage", "stg"),
		)

	keymap := huh.NewDefaultKeyMap()
	keymap.Select.Filter.SetKeys()

	form := NewForm(huh.NewGroup(selection)).WithKeyMap(keymap)

	return &MenuConfig{
		form:      form,
		shortcuts: Shortcuts{MenuMainShortcut, QuitShortcut},
	}
}

func (m *MenuConfig) Init() tea.Cmd {
	return m.form.Init()
}

func (m *MenuConfig) Update(msg tea.Msg) (Page, tea.Cmd) {
	if cmd, handled := m.shortcuts.Handle(msg, FormEditing(m.form)); handled {
		return m, cmd
	}

	var cmd tea.Cmd
	var completed bool

	m.form, cmd, completed = UpdateForm(m.form, msg)

	if completed {
		m.selected = m.form.GetString("action")

		switch m.selected {
		case "cmds":
			navigate := func() tea.Msg {
				return NavigateMsg{To: PageMenuCmds}
			}

			return m, tea.Batch(cmd, navigate)

		case "mans":
			navigate := func() tea.Msg {
				return NavigateMsg{To: PageMenuMans}
			}

			return m, tea.Batch(cmd, navigate)

		case "stg":
			navigate := func() tea.Msg {
				return NavigateMsg{To: PageMenuStg}
			}

			return m, tea.Batch(cmd, navigate)
		}
	}

	return m, cmd
}

func (m *MenuConfig) View() string {
	content := m.form.View()

	if m.form.State == huh.StateCompleted {
		content = "Selected: " + m.selected
	}

	return content + "\n\n" + FormHelp(m.form, m.shortcuts)
}
