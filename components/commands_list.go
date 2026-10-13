package components

import (
	"fmt"
	"tartarus/repositories"
)

func CommandsList(r *repositories.Repos) {
	// if _, err := tea.NewProgram(newCommandsList(r.GetCommands)).Run(); err != nil {
	// 	fmt.Fprintf(os.Stderr, "Unable to display commands: %v\n", err)
	// }
	fmt.Println("list")
}

// type commandListItem struct {
// 	command repositories.Command
// }

// func (i commandListItem) Title() string       { return i.command.Name }
// func (i commandListItem) Description() string { return i.command.Description }
// func (i commandListItem) FilterValue() string { return i.command.Name }

// type commandsListModel struct {
// 	query       func(int, int, string) ([]repositories.Command, int, error)
// 	list        list.Model
// 	searchInput textinput.Model
// 	search      string
// 	editing     bool
// 	page        int
// 	pageSize    int
// 	total       int
// 	width       int
// 	height      int
// 	loading     bool
// 	err         error
// 	requestID   uint64
// }

// type commandsLoadedMsg struct {
// 	owner     *commandsListModel
// 	requestID uint64
// 	commands  []repositories.Command
// 	total     int
// 	err       error
// }

// func newCommandsList(query func(int, int, string) ([]repositories.Command, int, error)) *commandsListModel {
// 	delegate := list.NewDefaultDelegate()
// 	delegate.SetHeight(2)
// 	delegate.SetSpacing(1)

// 	delegate.Styles.NormalTitle = lipgloss.NewStyle().
// 		Foreground(lipgloss.Color(themes.PrimaryColor)).
// 		PaddingLeft(2)

// 	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
// 		Foreground(lipgloss.Color(themes.AssetColor)).
// 		Bold(true).
// 		Border(lipgloss.NormalBorder(), false, false, false, true).
// 		BorderForeground(lipgloss.Color(themes.AssetColor)).
// 		PaddingLeft(1)

// 	delegate.Styles.SelectedDesc = delegate.Styles.SelectedTitle.
// 		Bold(false).
// 		Foreground(lipgloss.Color(themes.AssetColor))

// 	l := list.New(nil, delegate, 0, 0)
// 	l.SetShowTitle(false)
// 	l.SetShowStatusBar(false)
// 	l.SetShowPagination(false)
// 	l.SetShowHelp(false)
// 	l.SetFilteringEnabled(false)
// 	l.DisableQuitKeybindings()

// 	input := textinput.New()
// 	input.Prompt = "Search: "
// 	input.Placeholder = "/ to search name or description"

// 	return &commandsListModel{
// 		query:       query,
// 		list:        l,
// 		searchInput: input,
// 	}
// }

// func (m *commandsListModel) Init() tea.Cmd {
// 	return nil
// }

// func (m *commandsListModel) loadPage() tea.Cmd {
// 	m.loading = true
// 	m.err = nil
// 	m.requestID++
// 	id, size, offset := m.requestID, m.pageSize, m.page*m.pageSize
// 	query := m.query
// 	search := m.search

// 	return func() tea.Msg {
// 		commands, total, err := query(size, offset, search)
// 		return commandsLoadedMsg{
// 			owner:     m,
// 			requestID: id,
// 			commands:  commands,
// 			total:     total,
// 			err:       err,
// 		}
// 	}
// }

// func (m *commandsListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
// 	switch msg := msg.(type) {
// 	case tea.WindowSizeMsg:
// 		m.width, m.height = msg.Width, msg.Height
// 		m.searchInput.SetWidth(max(1, msg.Width-8))
// 		bodyHeight := max(0, msg.Height-9)
// 		m.list.SetSize(max(0, msg.Width), bodyHeight)
// 		size := min(5, bodyHeight/3)

// 		if msg.Width < 20 {
// 			size = 0
// 		}

// 		if size != m.pageSize {
// 			m.pageSize = size
// 			m.page = 0

// 			if size == 0 {
// 				m.requestID++
// 				m.loading = false
// 				return m, nil
// 			}

// 			return m, m.loadPage()
// 		}

// 		return m, nil

// 	case commandsLoadedMsg:
// 		if msg.owner != m || msg.requestID != m.requestID {
// 			return m, nil
// 		}

// 		m.loading = false
// 		m.err = msg.err

// 		if msg.err != nil {
// 			return m, nil
// 		}

// 		m.total = msg.total
// 		lastPage := max(0, (m.total-1)/m.pageSize)

// 		if m.page > lastPage {
// 			m.page = lastPage
// 			return m, m.loadPage()
// 		}

// 		items := make([]list.Item, len(msg.commands))

// 		for i, command := range msg.commands {
// 			items[i] = commandListItem{command: command}
// 		}

// 		cmd := m.list.SetItems(items)
// 		m.list.ResetSelected()

// 		return m, cmd

// 	case tea.KeyPressMsg:
// 		if msg.String() == "ctrl+c" {
// 			return m, tea.Quit
// 		}

// 		if m.editing {
// 			switch msg.String() {
// 			case "esc":
// 				m.editing = false
// 				m.searchInput.Blur()
// 				m.searchInput.SetValue(m.search)

// 				return m, nil

// 			case "enter":
// 				m.editing = false
// 				m.searchInput.Blur()
// 				m.search = m.searchInput.Value()
// 				m.page = 0

// 				if m.pageSize > 0 {
// 					return m, m.loadPage()
// 				}

// 				return m, nil
// 			}

// 			var cmd tea.Cmd
// 			m.searchInput, cmd = m.searchInput.Update(msg)
// 			return m, cmd
// 		}

// 		switch msg.String() {
// 		case "q":
// 			return m, tea.Quit

// 		case "esc":
// 			if m.search == "" {
// 				return m, tea.Quit
// 			}

// 			m.search = ""
// 			m.searchInput.SetValue("")
// 			m.page = 0

// 			if m.pageSize > 0 {
// 				return m, m.loadPage()
// 			}

// 			return m, nil
// 		case "/":
// 			if m.pageSize == 0 {
// 				return m, nil
// 			}

// 			m.editing = true
// 			return m, m.searchInput.Focus()
// 		}

// 		if m.loading || m.pageSize == 0 {
// 			return m, nil
// 		}

// 		if m.err != nil {
// 			if msg.String() == "r" {
// 				return m, m.loadPage()
// 			}

// 			return m, nil
// 		}

// 		switch msg.String() {
// 		case "left", "h", "pgup":
// 			if m.page > 0 {
// 				m.page--
// 				return m, m.loadPage()
// 			}

// 			return m, nil

// 		case "right", "l", "pgdown":
// 			if (m.page+1)*m.pageSize < m.total {
// 				m.page++
// 				return m, m.loadPage()
// 			}

// 			return m, nil

// 		}
// 	}

// 	var cmd tea.Cmd

// 	if m.editing {
// 		m.searchInput, cmd = m.searchInput.Update(msg)
// 		return m, cmd
// 	}

// 	m.list, cmd = m.list.Update(msg)
// 	return m, cmd
// }

// func (m *commandsListModel) View() tea.View {
// 	var content string

// 	help := "↑/↓ select · ←/→ page · / search · q quit"

// 	if m.editing {
// 		help = "enter search · esc cancel"
// 	}

// 	switch {
// 	case m.width == 0 || m.height == 0:
// 		content = "Preparing command list..."

// 	case m.pageSize == 0:
// 		content = "Enlarge the terminal."

// 	case m.loading:
// 		content = "Loading commands...\n\n" + help

// 	case m.err != nil:
// 		content = fmt.Sprintf("Unable to load commands: %v\n\n%s", m.err, help)
// 		if !m.editing {
// 			content += " · r retry"
// 		}

// 	case m.total == 0:
// 		content = "No commands found.\n\n" + help

// 	default:
// 		pages := (m.total + m.pageSize - 1) / m.pageSize

// 		s := strings.Builder{}
// 		s.WriteString("%s\n")
// 		s.WriteString("Page %d of %d")
// 		s.WriteString(" · %d commands\n")
// 		s.WriteString(help)

// 		content = fmt.Sprintf(
// 			s.String(),
// 			m.list.View(),
// 			m.page+1,
// 			pages,
// 			m.total,
// 		)
// 	}

// 	if m.pageSize > 0 {
// 		s := strings.Builder{}
// 		s.WriteString(themes.RenderLogo("List with all commands"))
// 		s.WriteString("\n\n")
// 		s.WriteString(m.searchInput.View())
// 		s.WriteString("\n\n")
// 		s.WriteString(content)
// 		content = s.String()
// 	}

// 	v := tea.NewView(content)
// 	v.AltScreen = true
// 	return v
// }
