package pagemenus

import (
	"tartarus/views/pages"
	"tartarus/views/pages/pageutils"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

func NewMenuMain() *PageMenuMain {
	selection := huh.NewSelect[string]().
		Key("action").
		Title("Select an option").
		Options(
			huh.NewOption("🚀 Run Command", "run"),
			huh.NewOption("📜 Read Manual", "man"),
			huh.NewOption("⚙️ Configurations", "config"),
		)

	keymap := huh.NewDefaultKeyMap()
	keymap.Select.Filter.SetKeys()

	form := pageutils.NewForm(huh.NewGroup(selection)).WithKeyMap(keymap)

	return &PageMenuMain{
		form:      form,
		shortcuts: pages.Shortcuts{pages.QuitShortcut},
	}
}

func (m *PageMenuMain) Init() tea.Cmd {
	return m.form.Init()
}

func (m *PageMenuMain) Update(msg tea.Msg) (pages.Page, tea.Cmd) {
	if cmd, handled := m.shortcuts.Handle(msg, pageutils.FormEditing(m.form)); handled {
		return m, cmd
	}

	var cmd tea.Cmd
	var completed bool

	m.form, cmd, completed = pageutils.UpdateForm(m.form, msg)

	if completed {
		m.selected = m.form.GetString("action")

		switch m.selected {
		case "config":
			navigate := func() tea.Msg {
				return pages.NavigateMsg{To: pages.PageMenuConfig}
			}

			return m, tea.Batch(cmd, navigate)
		}
	}

	return m, cmd
}

func (m *PageMenuMain) View() string {
	content := m.form.View()

	if m.form.State == huh.StateCompleted {
		content = "Selected: " + m.selected
	}

	return content + "\n\n" + pageutils.FormHelp(m.form, m.shortcuts)
}
