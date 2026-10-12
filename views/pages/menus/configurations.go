package pagemenus

import (
	"tartarus/views/pages"
	pageutils "tartarus/views/pages/utils"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

func NewMenuConfigurations() *PageMenuConfigurations {
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

	form := pageutils.NewForm(huh.NewGroup(selection)).WithKeyMap(keymap)

	return &PageMenuConfigurations{
		form:      form,
		shortcuts: pages.Shortcuts{pages.MenuMainShortcut, pages.QuitShortcut},
	}
}

func (m *PageMenuConfigurations) Init() tea.Cmd {
	return m.form.Init()
}

func (m *PageMenuConfigurations) Update(msg tea.Msg) (pages.Page, tea.Cmd) {
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

		switch m.selected {
		case "cmds":
			navigate := func() tea.Msg {
				return pages.NavigateMsg{To: pages.PageMenuCmds}
			}

			return m, tea.Batch(cmd, navigate)

		case "mans":
			navigate := func() tea.Msg {
				return pages.NavigateMsg{To: pages.PageMenuMans}
			}

			return m, tea.Batch(cmd, navigate)

		case "stg":
			navigate := func() tea.Msg {
				return pages.NavigateMsg{To: pages.PageMenuStg}
			}

			return m, tea.Batch(cmd, navigate)
		}
	}

	return m, cmd
}

func (m *PageMenuConfigurations) View() string {
	content := m.form.View()

	if m.form.State == huh.StateCompleted {
		content = "Selected: " + m.selected
	}

	return content + "\n\n" + pageutils.FormHelp(m.form, m.shortcuts, m.width)
}
