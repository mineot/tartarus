package pageforms

import (
	"fmt"
	"tartarus/repositories"
	"tartarus/views/pages"
	pagehelpers "tartarus/views/pages/helpers"
	pageutils "tartarus/views/pages/utils"
	"tartarus/views/themes"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
)

type commandListItem struct {
	command repositories.Command
}

func (i commandListItem) Title() string       { return i.command.Name }
func (i commandListItem) Description() string { return i.command.Description }
func (i commandListItem) FilterValue() string { return i.command.Name }

type commandsLoadedMsg struct {
	commands []repositories.Command
	owner    *FormCommandList
	listMsg  pagehelpers.ListLoadedMsg
}

type FormCommandList struct {
	loadCommands func(int, int, string) ([]repositories.Command, int, error)
	listModel    pagehelpers.ListModel
}

func NewFormCommandList(r *repositories.Repos) *FormCommandList {
	return &FormCommandList{
		loadCommands: r.GetCommands,
		listModel: pagehelpers.NewList(
			"Name or description",
			pages.Shortcuts{
				pages.MenuConfigsShortcut,
				pages.QuitShortcut,
			},
		),
	}
}

func (m *FormCommandList) Init() tea.Cmd {
	m.listModel.Started = true

	if !m.listModel.Sized {
		return nil
	}

	return m.fetch(0)
}

func (m *FormCommandList) fetch(page int) tea.Cmd {
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

func (m *FormCommandList) Update(msg tea.Msg) (pages.Page, tea.Cmd) {
	if model, cmd := m.listModel.ListSearchUpdate(msg, m, m.fetch); model != nil || cmd != nil {
		return model, cmd
	}

	if model, cmd := m.listModel.ListShortcutHandlerUpdate(msg, m); model != nil || cmd != nil {
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.listModel.ListWindowSizeUpdate(msg, m, m.fetch)

	case tea.KeyPressMsg:
		if model, cmd := m.listModel.ListKeyPressUpdate(msg, m, m.fetch); model != nil || cmd != nil {
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

		if model, cmd := m.listModel.ListVerifyPages(msg.listMsg, m, m.fetch); model != nil || cmd != nil {
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

		return m, cmd
	}

	if m.listModel.SearchEditing {
		var cmd tea.Cmd
		m.listModel.SearchInput, cmd = m.listModel.SearchInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *FormCommandList) View() string {
	status := "Loading commands..."

	if m.listModel.Sized && !m.listModel.Loading {
		if m.listModel.Total == 0 {
			status = "No commands found."
		} else {
			status = m.listModel.ListPrintFooterView("commands")
		}
	}

	if m.listModel.Err != nil {
		status = fmt.Sprintf("Could not load commands: %v", m.listModel.Err)
	}

	bindings := m.listModel.ListGetBindingsView()

	if m.listModel.Err != nil {
		bindings = append(bindings, m.listModel.Retry)
	}

	search := m.listModel.ListDisplayedSearchView()
	help := pageutils.RenderHelp(bindings, m.listModel.Shortcuts)
	help = m.listModel.ListSearchedShortcutsView(help, bindings)

	if searching := m.listModel.ListSearchingStatusView(); searching != "" {
		status = searching
	}

	search, help = m.listModel.ListSearchEditingView(search, help)
	title := themes.Form(true).Focused.Title.Render("🖥️ Commands")

	return title + "\n" + search + "\n\n" + m.listModel.List.View() + "\n" + status + "\n" + help
}
