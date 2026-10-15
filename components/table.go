package components

import (
	"fmt"
	"slices"
	"tartarus/themes"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TableOpts configures a standalone terminal table.
type TableOpts struct {
	Title    string
	Headers  []string
	Rows     [][]string
	EmptyMsg string
	// Widths specifies positive content widths, one per header. When omitted,
	// columns fit their headers and data. Explicit widths cap the content width.
	Widths []int
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

	if _, err := tea.NewProgram(m, tea.WithMouseCellMotion()).Run(); err != nil {
		return fmt.Errorf("components: running table: %w", err)
	}

	return nil
}

type tableModel struct {
	table         table.Model
	title         string
	emptyMsg      string
	widths        []int
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
		columnPadding: max(s.Header.GetHorizontalFrameSize(), s.Cell.GetHorizontalFrameSize()),
		headerHeight:  1 + s.Header.GetVerticalFrameSize(),
	}, nil
}

func (m *tableModel) Init() tea.Cmd {
	return nil
}

func (m *tableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.resize(msg.Width, msg.Height)
		return m, nil
	}

	if !m.ready || m.tooSmall || len(m.table.Rows()) == 0 {
		return m, nil
	}

	if mouse, ok := msg.(tea.MouseMsg); ok {
		if mouse.Action == tea.MouseActionPress {
			switch mouse.Button {
			case tea.MouseButtonWheelUp:
				m.table.MoveUp(1)
			case tea.MouseButtonWheelDown:
				m.table.MoveDown(1)
			}
		}

		return m, nil
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m *tableModel) resize(width, height int) {
	m.width, m.height, m.ready = width, height, true
	availableHeight := height - lipgloss.Height(m.footerView()) - 1

	if title := m.titleView(); title != "" {
		availableHeight -= lipgloss.Height(title) - 1
	}

	m.tooSmall = width < 1 || availableHeight < 1

	if len(m.table.Rows()) == 0 {
		return
	}

	cols := slices.Clone(m.table.Columns())
	availableWidth := width - len(cols)*m.columnPadding
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

	const help = "↑/k/PgUp up • ↓/j/PgDn down • q quit"
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(themes.TertiaryColor)).
		Width(min(m.width, lipgloss.Width(help))).
		Render(help)
}

func (m *tableModel) View() string {
	if !m.ready || m.width < 1 || m.height < 1 {
		return ""
	}

	frame := lipgloss.NewStyle().MaxWidth(m.width).MaxHeight(m.height)

	if m.tooSmall {
		return frame.Render("Terminal too small. Resize or press q to exit.")
	}

	if len(m.table.Rows()) == 0 {
		return frame.Render(m.titleView() + themes.GetTertiaryText(m.emptyMsg, true) + "\n\n" + m.footerView())
	}

	return frame.Render(m.titleView() + m.table.View() + "\n\n" + m.footerView())
}
