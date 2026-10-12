package pageforms

import (
	"fmt"
	"tartarus/repositories"
	"tartarus/views/pages"
	pagehelpers "tartarus/views/pages/helpers"
	pageutils "tartarus/views/pages/utils"
	"tartarus/views/themes"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
)

var (
	NewCommand = pages.Shortcut{
		Binding: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "New Command"),
		),
		Command: func() tea.Msg {
			return pages.NavigateMsg{To: pages.CommandEditor}
		},
	}

	EditCommand = pages.Shortcut{
		Binding: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "Edit Command"),
		),
	}
)

type commandListItem struct {
	command repositories.Command
}

func (i commandListItem) Title() string       { return i.command.Name }
func (i commandListItem) Description() string { return i.command.Description }
func (i commandListItem) FilterValue() string { return i.command.Name }

type commandsLoadedMsg struct {
	commands []repositories.Command
	owner    *CommandList
	listMsg  pagehelpers.ListLoadedMsg
}

type CommandList struct {
	loadCommands func(int, int, string) ([]repositories.Command, int, error)
	listModel    pagehelpers.ListModel
}

func NewCommandList(r *repositories.Repos) *CommandList {
	return &CommandList{
		loadCommands: r.GetCommands,
		listModel: pagehelpers.NewList(
			"Name or description",
			pages.Shortcuts{
				NewCommand,
				EditCommand,
				pages.MenuConfigsShortcut,
				pages.QuitShortcut,
			},
		),
	}
}

func (m *CommandList) Cursor() *tea.Cursor {
	if !m.listModel.SearchEditing {
		return nil
	}

	cursor := m.listModel.SearchInput.Cursor()

	if cursor != nil {
		cursor.Y++
	}

	return cursor
}

func (m *CommandList) Init() tea.Cmd {
	m.listModel.Started = true

	if !m.listModel.Sized {
		return nil
	}

	return m.fetch(0)
}

func (m *CommandList) fetch(page int) tea.Cmd {
	m.listModel.Request++
	m.listModel.Loading, m.listModel.Err, m.listModel.PendingPage = true, nil, page
	request, size, search := m.listModel.Request, m.listModel.Size, m.listModel.Search

	return func() tea.Msg {
		commands, total, err := m.loadCommands(size, page*size, search)

		return commandsLoadedMsg{
			owner:    m,
			commands: commands,
			listMsg: pagehelpers.ListLoadedMsg{
				Err:     err,
				Page:    page,
				Request: request,
				Search:  search,
				Size:    size,
				Total:   total,
			},
		}
	}
}

func (m *CommandList) Update(msg tea.Msg) (pages.Page, tea.Cmd) {
	if model, cmd := m.listModel.SearchUpdate(msg, m, m.fetch); model != nil || cmd != nil {
		return model, cmd
	}

	if pressed, ok := msg.(tea.KeyPressMsg); ok && !m.listModel.SearchEditing && key.Matches(pressed, EditCommand.Binding) {
		if m.listModel.Loading || m.listModel.Err != nil {
			return m, nil
		}

		selected, ok := m.listModel.List.SelectedItem().(commandListItem)

		if !ok || selected.command.ID == 0 {
			return m, nil
		}

		id := selected.command.ID

		return m, func() tea.Msg {
			return pages.NavigateMsg{To: pages.CommandEditor, RefID: id}
		}
	}

	if model, cmd := m.listModel.ShortcutsUpdate(msg, m); model != nil || cmd != nil {
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.listModel.WindowSizeUpdate(msg, m, m.fetch)

	case tea.KeyPressMsg:
		if model, cmd := m.listModel.KeyPressUpdate(msg, m, m.fetch); model != nil || cmd != nil {
			return model, cmd
		}

	case commandsLoadedMsg:
		if msg.owner != m || msg.listMsg.Request != m.listModel.Request {
			return m, nil
		}

		m.listModel.Loading = false

		if msg.listMsg.Err != nil {
			m.listModel.Err = msg.listMsg.Err
			return m, nil
		}

		if model, cmd := m.listModel.PageBoundsUpdate(msg.listMsg, m, m.fetch); model != nil || cmd != nil {
			return model, cmd
		}

		m.listModel.Page, m.listModel.Total, m.listModel.Err = msg.listMsg.Page, msg.listMsg.Total, nil
		m.listModel.DisplayedSize = msg.listMsg.Size
		m.listModel.DisplayedSearch = msg.listMsg.Search
		items := make([]list.Item, len(msg.commands))

		for i, command := range msg.commands {
			items[i] = commandListItem{command: command}
		}

		cmd := m.listModel.List.SetItems(items)
		m.listModel.List.ResetSelected()

		return m, tea.Batch(cmd, tea.ClearScreen)
	}

	if m.listModel.SearchEditing {
		return m, m.listModel.SearchInputUpdate(msg)
	}

	return m, nil
}

func (m *CommandList) View() string {
	status := "Loading commands..."

	if m.listModel.Sized && !m.listModel.Loading {
		if m.listModel.Total == 0 {
			status = "No commands found."
		} else {
			status = m.listModel.PaginationView("commands")
		}
	}

	if m.listModel.Err != nil {
		status = fmt.Sprintf("Could not load commands: %v", m.listModel.Err)
	}

	bindings := m.listModel.HelpBindings()

	if m.listModel.Err != nil {
		bindings = append(bindings, m.listModel.Retry)
	}

	search := m.listModel.SearchView()
	help := pageutils.RenderHelp(bindings, m.listModel.Shortcuts, m.listModel.List.Width())
	help = m.listModel.SearchHelpView(help, bindings)

	if searching := m.listModel.SearchStatusView(); searching != "" {
		status = searching
	}

	search, help = m.listModel.SearchEditingView(search, help)
	title := themes.Form(true).Focused.Title.Render("🖥️ Commands")

	return title + "\n" + search + "\n\n" + m.listModel.List.View() + "\n" + status + "\n" + help
}
