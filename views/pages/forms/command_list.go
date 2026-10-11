package pageforms

import (
	"fmt"
	"strings"
	"tartarus/repositories"
	"tartarus/views/pages"
	pagehelpers "tartarus/views/pages/helpers"
	pageutils "tartarus/views/pages/utils"
	"tartarus/views/themes"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
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
	msg      pagehelpers.ListLoadedMsg
}

type FormCommandList struct {
	loadCommands func(int, int, string) ([]repositories.Command, int, error)
	model        pagehelpers.ListModel
}

func NewFormCommandList(r *repositories.Repos) *FormCommandList {
	delegate := list.NewDefaultDelegate()
	delegate.Styles = themes.ListItems(true)
	delegate.SetSpacing(0)

	items := list.New(nil, delegate, 0, 0)
	items.SetShowTitle(false)
	items.SetShowStatusBar(false)
	items.SetShowPagination(false)
	items.SetShowHelp(false)
	items.SetFilteringEnabled(false)
	items.DisableQuitKeybindings()

	input := textinput.New()
	input.Prompt = "Search: "
	input.Placeholder = "Name or description"

	return &FormCommandList{
		loadCommands: r.GetCommands,
		model: pagehelpers.ListModel{
			SearchInput: input,
			SearchKey:   key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
			List:        items,
			Shortcuts:   pages.Shortcuts{pages.MenuConfigsShortcut, pages.QuitShortcut},
			Previous:    key.NewBinding(key.WithKeys("left", "pgup"), key.WithHelp("←/pgup", "previous page")),
			Next:        key.NewBinding(key.WithKeys("right", "pgdown"), key.WithHelp("→/pgdn", "next page")),
			Retry:       key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "retry")),
		},
	}
}

func (m *FormCommandList) Init() tea.Cmd {
	m.model.Started = true

	if !m.model.Sized {
		return nil
	}

	return m.fetch(0)
}

func (m *FormCommandList) fetch(page int) tea.Cmd {
	m.model.Request++
	request, size, search := m.model.Request, m.model.Size, m.model.Search
	load := m.loadCommands
	m.model.Loading, m.model.Err, m.model.PendingPage = true, nil, page

	return func() tea.Msg {
		commands, total, err := load(size, page*size, search)

		return commandsLoadedMsg{
			owner:    m,
			commands: commands,
			msg: pagehelpers.ListLoadedMsg{
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
	if pressed, ok := msg.(tea.KeyPressMsg); ok {
		if m.model.Editing {
			switch pressed.String() {
			case "enter":
				m.model.Search = strings.TrimSpace(m.model.SearchInput.Value())
				m.model.SearchInput.Blur()
				m.model.Editing = false
				if m.model.Sized {
					return m, m.fetch(0)
				}
				return m, nil

			case "esc":
				m.model.SearchInput.SetValue(m.model.Search)
				m.model.SearchInput.Blur()
				m.model.Editing = false
				return m, nil

			default:
				var cmd tea.Cmd
				m.model.SearchInput, cmd = m.model.SearchInput.Update(msg)
				return m, cmd
			}
		}

		if key.Matches(pressed, m.model.SearchKey) {
			m.model.SearchInput.SetValue(m.model.Search)
			m.model.Editing = true
			return m, m.model.SearchInput.Focus()
		}

		if pressed.String() == "esc" && (m.model.Search != "" || m.model.DisplayedSearch != "") {
			m.model.Search = ""
			m.model.SearchInput.SetValue("")

			if m.model.Sized {
				return m, m.fetch(0)
			}

			return m, nil
		}
	}

	if cmd, handled := m.model.Shortcuts.Handle(msg, m.model.Editing); handled {
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.model.SearchInput.SetWidth(max(1, msg.Width-8))
		height := max(2, msg.Height-8)
		size := min(5, height/2)
		oldSize := m.model.Size
		m.model.List.SetSize(max(1, msg.Width), height)
		m.model.Size, m.model.Sized = size, true

		if m.model.Started && oldSize != size {
			page := m.model.Page * m.model.DisplayedSize / size
			if m.model.Search != m.model.DisplayedSearch {
				page = 0
			}
			return m, m.fetch(page)
		}

		return m, nil

	case commandsLoadedMsg:
		if msg.owner != m || msg.msg.Request != m.model.Request {
			return m, nil
		}

		m.model.Loading = false

		if msg.msg.Err != nil {
			m.model.Err = msg.msg.Err
			return m, nil
		}

		if msg.msg.Page > 0 && msg.msg.Page*msg.msg.Size >= msg.msg.Total {
			return m, m.fetch(max(0, (msg.msg.Total-1)/msg.msg.Size))
		}

		m.model.Page, m.model.Total, m.model.Err = msg.msg.Page, msg.msg.Total, nil
		m.model.DisplayedSize = msg.msg.Size
		m.model.DisplayedSearch = msg.msg.Search
		items := make([]list.Item, len(msg.commands))

		for i, command := range msg.commands {
			items[i] = commandListItem{command: command}
		}

		cmd := m.model.List.SetItems(items)
		m.model.List.ResetSelected()

		return m, cmd

	case tea.KeyPressMsg:
		if m.model.Loading || !m.model.Sized {
			return m, nil
		}

		switch {
		case key.Matches(msg, m.model.Retry):
			if m.model.Err != nil {
				return m, m.fetch(m.model.PendingPage)
			}

		case key.Matches(msg, m.model.Previous):
			if m.model.Err == nil && m.model.Page > 0 {
				return m, m.fetch(m.model.Page - 1)
			}

		case key.Matches(msg, m.model.Next):
			if m.model.Err == nil && (m.model.Page+1)*m.model.Size < m.model.Total {
				return m, m.fetch(m.model.Page + 1)
			}

		case key.Matches(msg, m.model.List.KeyMap.CursorUp):
			m.model.List.CursorUp()

		case key.Matches(msg, m.model.List.KeyMap.CursorDown):
			m.model.List.CursorDown()
		}
	}

	if m.model.Editing {
		var cmd tea.Cmd
		m.model.SearchInput, cmd = m.model.SearchInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *FormCommandList) View() string {
	status := "Loading commands..."

	if m.model.Sized && !m.model.Loading {
		if m.model.Total == 0 {
			status = "No commands found."
		} else {
			status = fmt.Sprintf("Page %d of %d · %d commands", m.model.Page+1, 1+(m.model.Total-1)/max(1, m.model.DisplayedSize), m.model.Total)
		}
	}

	if m.model.Err != nil {
		status = fmt.Sprintf("Could not load commands: %v", m.model.Err)
	}

	bindings := []key.Binding{m.model.List.KeyMap.CursorUp, m.model.List.KeyMap.CursorDown, m.model.Previous, m.model.Next, m.model.SearchKey}

	if m.model.Err != nil {
		bindings = append(bindings, m.model.Retry)
	}

	search := "Search: " + m.model.DisplayedSearch
	help := pageutils.RenderHelp(bindings, m.model.Shortcuts)

	if m.model.Search != "" || m.model.DisplayedSearch != "" {
		shortcuts := append(pages.Shortcuts(nil), m.model.Shortcuts...)
		shortcuts[0].Binding = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "clear search"))
		help = pageutils.RenderHelp(bindings, shortcuts)
	}

	if m.model.Loading && m.model.Search != "" {
		status = fmt.Sprintf("Searching for %q...", m.model.Search)
	}

	if m.model.Editing {
		search = m.model.SearchInput.View()
		help = pageutils.RenderHelp([]key.Binding{
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "search")),
			key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
		}, nil)
	}

	title := themes.Form(true).Focused.Title.Render("🖥️ Commands")

	return title + "\n" + search + "\n\n" + m.model.List.View() + "\n" + status + "\n" + help
}
