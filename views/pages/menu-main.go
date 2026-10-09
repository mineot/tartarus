package pages

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

type MenuMain struct {
	form      *huh.Form
	selected  string
	shortcuts Shortcuts
}

func NewMenuMain() *MenuMain {
	selection := huh.NewSelect[string]().
		Key("action").
		Title("Select an option").
		Options(
			huh.NewOption("🖥️ Run a Command", "run"),
			huh.NewOption("📜 Read Manual", "man"),
			huh.NewOption("⚙️ Configurations", "config"),
		)

	keymap := huh.NewDefaultKeyMap()
	keymap.Select.Filter.SetKeys()

	form := NewForm(huh.NewGroup(selection)).WithKeyMap(keymap)

	return &MenuMain{
		form:      form,
		shortcuts: Shortcuts{QuitShortcut},
	}
}

func (m *MenuMain) Init() tea.Cmd {
	return m.form.Init()
}

func (m *MenuMain) Update(msg tea.Msg) (Page, tea.Cmd) {
	if cmd, handled := m.shortcuts.Handle(msg); handled {
		return m, cmd
	}

	var cmd tea.Cmd
	var completed bool

	m.form, cmd, completed = UpdateForm(m.form, msg)

	if completed {
		m.selected = m.form.GetString("action")

		switch m.selected {
		case "config":
			navigate := func() tea.Msg {
				return NavigateMsg{To: PageMenuConfig}
			}

			return m, tea.Batch(cmd, navigate)
		}
	}

	return m, cmd
}

func (m *MenuMain) View() string {
	content := m.form.View()

	if m.form.State == huh.StateCompleted {
		content = "Selected: " + m.selected
	}

	return content + "\n\n" + FormHelp(m.form, m.shortcuts)
}
