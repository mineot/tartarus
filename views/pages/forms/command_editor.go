package pageforms

import (
	"fmt"
	"strings"
	"tartarus/repositories"
	"tartarus/views/pages"
	pageutils "tartarus/views/pages/utils"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

var (
	escShortcut = pages.Shortcut{
		Binding: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		Command: func() tea.Msg {
			return pages.NavigateMsg{To: pages.CommandList}
		},
	}

	saveShortcut = pages.Shortcut{
		Binding: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "save"),
		),
	}
)

type CommandEditor struct {
	command   repositories.Command
	err       error
	form      *huh.Form
	shortcuts pages.Shortcuts
	status    string
	width     int
}

func NewCommandEditor(r *repositories.Repos, commandID uint64) *CommandEditor {
	var editor CommandEditor = CommandEditor{}

	if commandID != 0 {
		command, err := r.GetCommand(commandID)

		if err != nil {
			editor.status = fmt.Sprintf(
				"Fail to get command by id %d: %v",
				commandID,
				err.Error(),
			)
		}

		editor = CommandEditor{command: command}
	}

	keymap := huh.NewDefaultKeyMap()

	nameInput := huh.NewInput().
		Key("name").
		Value(&editor.command.Name).
		Placeholder("Command Name").
		Validate(func(value string) error {
			if strings.TrimSpace(value) == "" {
				return fmt.Errorf("name is required")
			}
			return nil
		})

	descriptionInput := huh.NewInput().
		Key("description").
		Value(&editor.command.Description).
		Placeholder("Command Description").
		Validate(func(value string) error {
			if strings.TrimSpace(value) == "" {
				return fmt.Errorf("description is required")
			}
			return nil
		})

	editor.form = pageutils.
		NewForm(huh.NewGroup(nameInput, descriptionInput)).
		WithKeyMap(keymap)

	return &editor
}

func (m *CommandEditor) Cursor() *tea.Cursor {
	return nil
}

func (m *CommandEditor) Init() tea.Cmd {
	return m.form.Init()
}

func (m *CommandEditor) Update(msg tea.Msg) (pages.Page, tea.Cmd) {
	if size, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = size.Width
	}

	if pressed, ok := msg.(tea.KeyPressMsg); ok {
		switch pressed.String() {
		case "esc":
			return m, escShortcut.Command
		}
	}

	if cmd, handled := m.shortcuts.Handle(
		msg,
		pageutils.FormEditing(m.form),
	); handled {
		return m, cmd
	}

	var cmd tea.Cmd
	var completed bool

	m.form, cmd, completed = pageutils.UpdateForm(m.form, msg)

	if completed {
		m.command.Name = m.form.GetString("name")
		m.command.Description = m.form.GetString("description")
	}

	return m, cmd
}

func (m *CommandEditor) View() string {
	var bindings []key.Binding
	shortcuts := m.shortcuts

	if m.form.State == huh.StateNormal {
		bindings = append(bindings, m.form.KeyBinds()...)
	}

	if pageutils.FormEditing(m.form) {
		shortcuts = nil
	}

	bindings = append(bindings, escShortcut.Binding)

	content := m.form.View()

	if m.form.State == huh.StateCompleted {
		content = "Name: " + m.command.Name
		content += "\nDescription: " + m.command.Description
		bindings = append(bindings, saveShortcut.Binding)
	}

	return content + "\n\n" + pageutils.RenderHelp(bindings, shortcuts, m.width)
}
