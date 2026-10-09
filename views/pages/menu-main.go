package pages

import (
	"tartarus/views/themes"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

// MenuMain owns the main menu form and its selected action.
type MenuMain struct {
	form     *huh.Form
	selected string
}

// NewMenuMain creates a main menu with a fresh form.
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

	form := huh.
		NewForm(huh.NewGroup(selection)).
		WithTheme(huh.ThemeFunc(themes.Form)).
		WithKeyMap(keymap)

	return &MenuMain{form: form}
}

// Init initializes the menu form.
func (m *MenuMain) Init() tea.Cmd {
	return m.form.Init()
}

// Update handles menu input and the confirmed selection.
func (m *MenuMain) Update(msg tea.Msg) (Page, tea.Cmd) {
	if key, ok := msg.(tea.KeyPressMsg); ok && key.String() == "q" {
		return m, tea.Quit
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

// View renders the menu content and its local shortcut.
func (m *MenuMain) View() string {
	content := m.form.View()

	if m.form.State == huh.StateCompleted {
		content = "Selected: " + m.selected
	}

	return content + "\n\nPress q to quit.\n"
}
