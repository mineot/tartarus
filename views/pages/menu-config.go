package pages

import (
	"tartarus/views/themes"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
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
		Title("Select an option").
		Options(
			huh.NewOption("Configurations", "configs"),
		)

	keymap := huh.NewDefaultKeyMap()
	keymap.Select.Filter.SetKeys()

	form := huh.
		NewForm(huh.NewGroup(selection)).
		WithTheme(huh.ThemeFunc(themes.Form)).
		WithKeyMap(keymap).
		WithShowHelp(false)

	return &MenuConfig{
		form:      form,
		shortcuts: Shortcuts{MainMenuShortcut, QuitShortcut},
	}
}

func (m *MenuConfig) Init() tea.Cmd {
	return m.form.Init()
}

func (m *MenuConfig) Update(msg tea.Msg) (Page, tea.Cmd) {
	if cmd, handled := m.shortcuts.Handle(msg); handled {
		return m, cmd
	}

	if m.form.State != huh.StateNormal {
		return m, nil
	}

	updated, cmd := m.form.Update(msg)

	if form, ok := updated.(*huh.Form); ok {
		m.form = form
	}

	if m.form.State == huh.StateCompleted {
		m.selected = m.form.GetString("action")
	}

	return m, cmd
}

func (m *MenuConfig) View() string {
	content := m.form.View()

	var bindings []key.Binding

	if m.form.State == huh.StateNormal {
		bindings = m.form.KeyBinds()
	}

	bindings = append(bindings, m.shortcuts.Bindings()...)

	if m.form.State == huh.StateCompleted {
		content = "Selected: " + m.selected
	}

	h := help.New()

	return content + "\n\n" + h.ShortHelpView(bindings)
}
