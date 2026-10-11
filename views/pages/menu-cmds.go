package pages

import (
	"fmt"
	"strings"
	"tartarus/repositories"

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
	search            string
	owner             *MenuCmds
	request           uint64
	page, size, total int
	commands          []repositories.Command
	err               error
}

type MenuCmds struct {
	input                   textinput.Model
	search, displayedSearch string
	editing                 bool
	searchKey               key.Binding
	displayedSize           int
	list                    list.Model
	loadCommands            func(int, int, string) ([]repositories.Command, int, error)
	shortcuts               Shortcuts
	previous, next, retry   key.Binding
	page, size, total       int
	pendingPage             int
	request                 uint64
	started, sized, loading bool
	err                     error
}

func NewMenuCmds(r *repositories.Repos) *MenuCmds {
	delegate := list.NewDefaultDelegate()
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

	return &MenuCmds{
		input:        input,
		searchKey:    key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
		list:         items,
		loadCommands: r.GetCommands,
		shortcuts:    Shortcuts{MenuConfigsShortcut, QuitShortcut},
		previous:     key.NewBinding(key.WithKeys("left", "pgup"), key.WithHelp("←/pgup", "previous page")),
		next:         key.NewBinding(key.WithKeys("right", "pgdown"), key.WithHelp("→/pgdn", "next page")),
		retry:        key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "retry")),
	}
}

func (m *MenuCmds) Init() tea.Cmd {
	m.started = true

	if !m.sized {
		return nil
	}

	return m.fetch(0)
}

func (m *MenuCmds) fetch(page int) tea.Cmd {
	m.request++
	request, size, search := m.request, m.size, m.search
	load := m.loadCommands
	m.loading, m.err, m.pendingPage = true, nil, page

	return func() tea.Msg {
		commands, total, err := load(size, page*size, search)

		return commandsLoadedMsg{
			owner:    m,
			search:   search,
			request:  request,
			page:     page,
			size:     size,
			total:    total,
			commands: commands,
			err:      err,
		}
	}
}

func (m *MenuCmds) Update(msg tea.Msg) (Page, tea.Cmd) {
	if pressed, ok := msg.(tea.KeyPressMsg); ok {
		if m.editing {
			switch pressed.String() {
			case "enter":
				m.search = strings.TrimSpace(m.input.Value())
				m.input.Blur()
				m.editing = false
				if m.sized {
					return m, m.fetch(0)
				}
				return m, nil

			case "esc":
				m.input.SetValue(m.search)
				m.input.Blur()
				m.editing = false
				return m, nil

			default:
				var cmd tea.Cmd
				m.input, cmd = m.input.Update(msg)
				return m, cmd
			}
		}

		if key.Matches(pressed, m.searchKey) {
			m.input.SetValue(m.search)
			m.editing = true
			return m, m.input.Focus()
		}

		if pressed.String() == "esc" && (m.search != "" || m.displayedSearch != "") {
			m.search = ""
			m.input.SetValue("")

			if m.sized {
				return m, m.fetch(0)
			}

			return m, nil
		}
	}

	if cmd, handled := m.shortcuts.Handle(msg); handled {
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.input.SetWidth(max(1, msg.Width-8))
		height := max(2, msg.Height-8)
		size := min(5, height/2)
		oldSize := m.size
		m.list.SetSize(max(1, msg.Width), height)
		m.size, m.sized = size, true

		if m.started && oldSize != size {
			page := m.page * m.displayedSize / size
			if m.search != m.displayedSearch {
				page = 0
			}
			return m, m.fetch(page)
		}

		return m, nil

	case commandsLoadedMsg:
		if msg.owner != m || msg.request != m.request {
			return m, nil
		}

		m.loading = false

		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}

		if msg.page > 0 && msg.page*msg.size >= msg.total {
			return m, m.fetch(max(0, (msg.total-1)/msg.size))
		}

		m.page, m.total, m.err = msg.page, msg.total, nil
		m.displayedSize = msg.size
		m.displayedSearch = msg.search
		items := make([]list.Item, len(msg.commands))

		for i, command := range msg.commands {
			items[i] = commandListItem{command: command}
		}

		cmd := m.list.SetItems(items)
		m.list.ResetSelected()

		return m, cmd

	case tea.KeyPressMsg:
		if m.loading || !m.sized {
			return m, nil
		}

		switch {
		case key.Matches(msg, m.retry):
			if m.err != nil {
				return m, m.fetch(m.pendingPage)
			}

		case key.Matches(msg, m.previous):
			if m.err == nil && m.page > 0 {
				return m, m.fetch(m.page - 1)
			}

		case key.Matches(msg, m.next):
			if m.err == nil && (m.page+1)*m.size < m.total {
				return m, m.fetch(m.page + 1)
			}

		case key.Matches(msg, m.list.KeyMap.CursorUp):
			m.list.CursorUp()

		case key.Matches(msg, m.list.KeyMap.CursorDown):
			m.list.CursorDown()
		}
	}

	if m.editing {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *MenuCmds) View() string {
	status := "Loading commands..."

	if m.sized && !m.loading {
		if m.total == 0 {
			status = "No commands found."
		} else {
			status = fmt.Sprintf("Page %d of %d · %d commands", m.page+1, 1+(m.total-1)/max(1, m.displayedSize), m.total)
		}
	}

	if m.err != nil {
		status = fmt.Sprintf("Could not load commands: %v", m.err)
	}

	bindings := []key.Binding{m.list.KeyMap.CursorUp, m.list.KeyMap.CursorDown, m.previous, m.next, m.searchKey}

	if m.err != nil {
		bindings = append(bindings, m.retry)
	}

	search := "Search: " + m.displayedSearch
	help := RenderHelp(bindings, m.shortcuts)

	if m.search != "" || m.displayedSearch != "" {
		shortcuts := append(Shortcuts(nil), m.shortcuts...)
		shortcuts[0].Binding = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "clear search"))
		help = RenderHelp(bindings, shortcuts)
	}

	if m.loading && m.search != "" {
		status = fmt.Sprintf("Searching for %q...", m.search)
	}

	if m.editing {
		search = m.input.View()
		help = RenderHelp([]key.Binding{
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "search")),
			key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
		}, nil)
	}

	return "🖥️ Commands\n" + search + "\n\n" + m.list.View() + "\n" + status + "\n" + help
}
