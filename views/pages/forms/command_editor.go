package pageforms

import (
	"tartarus/repositories"
	"tartarus/views/pages"

	tea "charm.land/bubbletea/v2"
)

type CommandEditor struct {
	commandID uint64
}

func NewCommandEditor(r *repositories.Repos, commandID uint64) *CommandEditor {
	return &CommandEditor{commandID: commandID}
}

func (m *CommandEditor) Cursor() *tea.Cursor {
	return nil
}

func (m *CommandEditor) Init() tea.Cmd {
	return nil
}

func (m *CommandEditor) Update(msg tea.Msg) (pages.Page, tea.Cmd) {
	return m, nil
}

func (m *CommandEditor) View() string {
	return ""
}
