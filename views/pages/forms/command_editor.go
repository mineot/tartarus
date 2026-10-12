package pageforms

import (
	"tartarus/repositories"
	"tartarus/views/pages"

	tea "charm.land/bubbletea/v2"
)

type CommandEditor struct{}

func NewCommandEditor(r *repositories.Repos) *CommandEditor {
	return &CommandEditor{}
}

func (m *CommandEditor) Cursor() *tea.Cursor {
	return nil
}

func (m *CommandEditor) Init() tea.Cmd {
	return nil
}

func (m *CommandEditor) Update(msg tea.Msg) (pages.Page, tea.Cmd) {
	return nil, nil
}

func (m *CommandEditor) View() string {
	return ""
}
