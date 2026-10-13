package pagemenus

import (
	"tartarus/views/pages"
	pageutils "tartarus/views/pages/utils"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

func NewMenuCommands() *PageMenuCommands {
	selection := huh.NewSelect[string]().
		Key("action").
		Title("🖥️ Commands").
		Options(
			huh.NewOption("List All", "ls-cmds"),
			huh.NewOption("New Command", "n-cmd"),
			huh.NewOption("Edit Command", "e-cmd"),
			huh.NewOption("Remove Command", "r-cmd"),
		)

	selection2 := huh.NewSelect[string]().
		Key("action").
		Title("📋 Scripts").
		Options(
			huh.NewOption("List All Scripts", "ls-scps"),
			huh.NewOption("Append Script", "a-scp"),
			huh.NewOption("Edit Script", "e-scp"),
			huh.NewOption("Remove Script", "r-scp"),
		)

	keymap := huh.NewDefaultKeyMap()
	keymap.Select.Filter.SetKeys()

	form := pageutils.NewForm(huh.NewGroup(selection, selection2)).WithKeyMap(keymap)

	return &PageMenuCommands{
		form: form,
		shortcuts: pages.Shortcuts{
			pages.MenuConfigsShortcut,
			pages.QuitShortcut,
		},
	}
}

func (m *PageMenuCommands) Init() tea.Cmd {
	return m.form.Init()
}

func (m *PageMenuCommands) Update(msg tea.Msg) (pages.Page, tea.Cmd) {
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

		// switch m.selected {
		// case "cmds":
		// 	navigate := func() tea.Msg {
		// 		return pages.NavigateMsg{To: pages.CommandList}
		// 	}

		// 	return m, tea.Batch(cmd, navigate)

		// case "mans":
		// 	navigate := func() tea.Msg {
		// 		return pages.NavigateMsg{To: pages.ManualList}
		// 	}

		// 	return m, tea.Batch(cmd, navigate)

		// case "stg":
		// 	navigate := func() tea.Msg {
		// 		return pages.NavigateMsg{To: pages.StorageList}
		// 	}

		// 	return m, tea.Batch(cmd, navigate)
		// }
	}

	return m, cmd
}

func (m *PageMenuCommands) View() string {
	content := m.form.View()

	if m.form.State == huh.StateCompleted {
		content = "Selected: " + m.selected
	}

	return content + "\n\n" + pageutils.FormHelp(m.form, m.shortcuts, m.width)
}
