package components

import (
	"fmt"
	"slices"
	"strings"
	"tartarus/themes"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const noMatchMsg = "No matching rows."

// Shortcut is a custom key binding rendered in the table footer.
type Shortcut struct {
	// Key is matched against the pressed key, for example "e" or "ctrl+d".
	Key string
	// Label is shown next to the key in the footer, for example "edit".
	Label string
	// Run is called when Key is pressed. selected is the current row index, or
	// -1 when the table has no rows. Returning quit stops the table; returning
	// a non-nil error stops the table and is returned by Table. A nil Run
	// displays the shortcut without acting on it.
	Run func(selected int) (quit bool, err error)
}

// TableOpts configures a standalone terminal table.
type TableOpts struct {
	Title    string
	Headers  []string
	Rows     [][]string
	EmptyMsg string
	// Widths specifies positive content widths, one per header. When omitted,
	// columns fit their headers and data. Explicit widths cap the content width.
	Widths []int
	// Shortcuts adds custom key bindings to the footer. Keys the table already
	// uses for navigation or quitting are rejected.
	Shortcuts []Shortcut
	// Deprecated: selection is always highlighted and navigation is always enabled.
	// Retained for caller compatibility. Enter does not perform an action.
	Selectable bool
}

// Table validates opts and runs the table until q, Escape, or Ctrl+C is pressed.
// Errors are returned to the caller without being printed.
func Table(opts TableOpts) error {
	m, err := newTable(opts)

	if err != nil {
		return err
	}

	final, err := tea.NewProgram(m).Run()

	if err != nil {
		return fmt.Errorf("components: running table: %w", err)
	}

	if fm, ok := final.(*tableModel); ok && fm.runErr != nil {
		return fm.runErr
	}

	return nil
}

type tableModel struct {
	table         table.Model
	title         string
	emptyMsg      string
	widths        []int
	allRows       []table.Row
	shortcuts     map[string]Shortcut
	helpText      string
	runErr        error
	search        textinput.Model
	searchFocused bool
	columnPadding int
	headerHeight  int
	width         int
	height        int
	ready         bool
	tooSmall      bool
}

func newTable(opts TableOpts) (*tableModel, error) {
	if len(opts.Headers) == 0 {
		return nil, fmt.Errorf("components: table: at least one header is required")
	}

	if len(opts.Widths) != 0 && len(opts.Widths) != len(opts.Headers) {
		return nil, fmt.Errorf("components: table: expected %d column widths, got %d", len(opts.Headers), len(opts.Widths))
	}

	for i, width := range opts.Widths {
		if width <= 0 {
			return nil, fmt.Errorf("components: table: column %d width must be positive", i+1)
		}
	}

	reserved := reservedKeys()
	shortcuts := make(map[string]Shortcut, len(opts.Shortcuts))

	for i, shortcut := range opts.Shortcuts {
		if shortcut.Key == "" {
			return nil, fmt.Errorf("components: table: shortcut %d: key is required", i+1)
		}

		if shortcut.Label == "" {
			return nil, fmt.Errorf("components: table: shortcut %d: label is required", i+1)
		}

		if _, taken := reserved[shortcut.Key]; taken {
			return nil, fmt.Errorf("components: table: shortcut %d: key %q is reserved", i+1, shortcut.Key)
		}

		if _, duplicate := shortcuts[shortcut.Key]; duplicate {
			return nil, fmt.Errorf("components: table: shortcut %d: duplicate key %q", i+1, shortcut.Key)
		}

		shortcuts[shortcut.Key] = shortcut
	}

	help := strings.Builder{}
	help.WriteString("↑/k/PgUp up • ↓/j/PgDn down • / search")

	for _, shortcut := range opts.Shortcuts {
		help.WriteString(" • ")
		help.WriteString(shortcut.Key)
		help.WriteString(" ")
		help.WriteString(shortcut.Label)
	}

	help.WriteString(" • q quit")

	rows := make([]table.Row, len(opts.Rows))

	for i, row := range opts.Rows {
		if len(row) != len(opts.Headers) {
			return nil, fmt.Errorf("components: table: row %d has %d cells; expected %d", i+1, len(row), len(opts.Headers))
		}

		rows[i] = table.Row(slices.Clone(row))
	}

	cols := make([]table.Column, len(opts.Headers))
	widths := make([]int, len(opts.Headers))

	for i, header := range opts.Headers {
		widths[i] = max(1, lipgloss.Width(header))

		for _, row := range rows {
			widths[i] = max(widths[i], lipgloss.Width(row[i]))
		}

		if len(opts.Widths) != 0 {
			widths[i] = min(widths[i], opts.Widths[i])
		}

		cols[i] = table.Column{Title: header, Width: 1}
	}

	s := table.DefaultStyles()

	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(themes.PrimaryColor)).
		BorderBottom(true).
		Bold(true)

	s.Selected = lipgloss.NewStyle().
		Foreground(lipgloss.Color(themes.AssetColor)).
		Reverse(true).
		Bold(true)

	emptyMsg := opts.EmptyMsg

	if emptyMsg == "" {
		emptyMsg = "No rows found."
	}

	search := textinput.New()
	search.Prompt = "/ "
	search.Placeholder = "type to filter"
	search.SetStyles(searchStyles())

	return &tableModel{
		table: table.New(
			table.WithColumns(cols),
			table.WithRows(rows),
			table.WithFocused(true),
			table.WithStyles(s),
		),
		title:         opts.Title,
		emptyMsg:      emptyMsg,
		widths:        widths,
		allRows:       rows,
		shortcuts:     shortcuts,
		helpText:      help.String(),
		search:        search,
		columnPadding: max(s.Header.GetHorizontalFrameSize(), s.Cell.GetHorizontalFrameSize()),
		headerHeight:  1 + s.Header.GetVerticalFrameSize(),
	}, nil
}

func searchStyles() textinput.Styles {
	styles := textinput.DefaultStyles(true)

	styles.Focused.Prompt = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(themes.AssetColor))
	styles.Focused.Text = lipgloss.NewStyle().Foreground(lipgloss.Color(themes.PrimaryColor))
	styles.Focused.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color(themes.TertiaryColor))
	styles.Blurred.Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color(themes.TertiaryColor))
	styles.Blurred.Text = lipgloss.NewStyle().Foreground(lipgloss.Color(themes.PrimaryColor))
	styles.Blurred.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color(themes.TertiaryColor))
	styles.Cursor.Color = lipgloss.Color(themes.AssetColor)

	return styles
}

// reservedKeys returns the keys the table handles itself: quitting plus every
// key in the default navigation keymap.
func reservedKeys() map[string]struct{} {
	reserved := map[string]struct{}{
		"q":      {},
		"esc":    {},
		"ctrl+c": {},
		"/":      {},
	}

	km := table.DefaultKeyMap()

	for _, binding := range []key.Binding{
		km.LineUp,
		km.LineDown,
		km.PageUp,
		km.PageDown,
		km.HalfPageUp,
		km.HalfPageDown,
		km.GotoTop,
		km.GotoBottom,
	} {
		for _, k := range binding.Keys() {
			reserved[k] = struct{}{}
		}
	}

	return reserved
}

func (m *tableModel) Init() tea.Cmd {
	return nil
}

func (m *tableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		key := msg.String()

		if key == "ctrl+c" {
			return m, tea.Quit
		}

		if m.searchFocused {
			switch key {
			case "esc":
				m.search.SetValue("")
				m.search.Blur()
				m.searchFocused = false
				m.applyFilter()
				return m, nil
			case "enter":
				m.search.Blur()
				m.searchFocused = false
				return m, nil
			}

			var cmd tea.Cmd
			m.search, cmd = m.search.Update(msg)
			m.applyFilter()
			return m, cmd
		}

		switch key {
		case "q", "esc":
			return m, tea.Quit
		case "/":
			m.searchFocused = true
			m.search.CursorEnd()
			return m, m.search.Focus()
		}

		if shortcut, ok := m.shortcuts[key]; ok {
			if shortcut.Run == nil {
				return m, nil
			}

			selected := -1

			if len(m.table.Rows()) > 0 {
				selected = m.table.Cursor()
			}

			quit, err := shortcut.Run(selected)

			if err != nil {
				m.runErr = err
				return m, tea.Quit
			}

			if quit {
				return m, tea.Quit
			}

			return m, nil
		}

	case tea.WindowSizeMsg:
		m.resize(msg.Width, msg.Height)
		return m, nil
	}

	if !m.ready || m.tooSmall || len(m.table.Rows()) == 0 {
		return m, nil
	}

	if mouse, ok := msg.(tea.MouseWheelMsg); ok {
		switch mouse.Button {
		case tea.MouseWheelUp:
			m.table.MoveUp(1)
		case tea.MouseWheelDown:
			m.table.MoveDown(1)
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m *tableModel) resize(width, height int) {
	m.width, m.height, m.ready = width, height, true
	m.layout()
}

func (m *tableModel) layout() {
	m.search.SetWidth(max(1, m.width-lipgloss.Width(m.search.Prompt)))

	availableHeight := m.height - lipgloss.Height(m.footerView()) - lipgloss.Height(m.searchView()) - 2

	if title := m.titleView(); title != "" {
		availableHeight -= lipgloss.Height(title) - 1
	}

	m.tooSmall = m.width < 1 || availableHeight < 1

	if len(m.table.Rows()) == 0 {
		return
	}

	cols := slices.Clone(m.table.Columns())
	availableWidth := m.width - len(cols)*m.columnPadding
	m.tooSmall = m.tooSmall || availableWidth < len(cols) || availableHeight < m.headerHeight+1

	if m.tooSmall {
		return
	}

	// Give each column one cell, then share the remaining space, respecting
	// explicit widths as upper bounds and redistributing unused space.
	for i := range cols {
		cols[i].Width = 1
	}

	remaining := availableWidth - len(cols)

	for remaining > 0 {
		active := 0

		for i, col := range cols {
			if col.Width < m.widths[i] {
				active++
			}
		}

		if active == 0 {
			break
		}

		share := max(1, remaining/active)

		for i := range cols {
			growth := min(share, remaining, m.widths[i]-cols[i].Width)
			cols[i].Width += growth
			remaining -= growth
		}
	}

	tableWidth := len(cols) * m.columnPadding

	for _, col := range cols {
		tableWidth += col.Width
	}

	m.table.SetColumns(cols)
	m.table.SetWidth(tableWidth)
	m.table.SetHeight(min(availableHeight, m.headerHeight+min(10, len(m.table.Rows()))))

	// Reset the scroll offset before restoring the selection at the new height.
	cursor := m.table.Cursor()
	m.table.SetCursor(0)
	m.table.MoveUp(0)
	m.table.MoveDown(cursor)
}

func (m *tableModel) applyFilter() {
	query := strings.ToLower(m.search.Value())

	filtered := m.allRows

	if query != "" {
		filtered = make([]table.Row, 0, len(m.allRows))

		for _, row := range m.allRows {
			for _, cell := range row {
				if strings.Contains(strings.ToLower(cell), query) {
					filtered = append(filtered, row)
					break
				}
			}
		}
	}

	m.table.SetRows(filtered)
	m.table.SetCursor(0)
	m.table.MoveUp(0)

	if m.ready {
		m.layout()
	}
}

func (m *tableModel) searchView() string {
	if m.width < 1 {
		return ""
	}

	return m.search.View()
}

func (m *tableModel) titleView() string {
	if m.title == "" || m.width < 1 {
		return ""
	}

	return lipgloss.NewStyle().MaxWidth(m.width).
		Render(themes.GetPrimaryText(m.title, true)) + "\n\n"
}

func (m *tableModel) footerView() string {
	if m.width < 1 {
		return ""
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(themes.TertiaryColor)).
		Width(min(m.width, lipgloss.Width(m.helpText))).
		Render(m.helpText)
}

func (m *tableModel) View() tea.View {
	view := tea.NewView(m.content())
	view.MouseMode = tea.MouseModeCellMotion
	return view
}

func (m *tableModel) content() string {
	if !m.ready || m.width < 1 || m.height < 1 {
		return ""
	}

	frame := lipgloss.NewStyle().MaxWidth(m.width).MaxHeight(m.height)

	if m.tooSmall {
		return frame.Render("Terminal too small. Resize or press q to exit.")
	}

	var body string

	switch {
	case len(m.allRows) == 0:
		body = themes.GetTertiaryText(m.emptyMsg, true)
	case len(m.table.Rows()) == 0:
		body = themes.GetTertiaryText(noMatchMsg, true)
	default:
		body = m.table.View()
	}

	return frame.Render(m.titleView() + body + "\n\n" + m.searchView() + "\n\n" + m.footerView())
}
