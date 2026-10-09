package views

import (
	"fmt"
	"os"
	"strings"
	views "tartarus/views/pages"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
)

const (
	menuMain uint = iota
)

type model struct {
	currentPage uint
	menu        *huh.Form
	selected    string
}

func initialModel() model {
	return model{
		currentPage: menuMain,
		menu:        views.NewMenuMain(),
	}
}

func (m model) Init() tea.Cmd {
	return m.menu.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	if m.currentPage != menuMain || m.menu.State != huh.StateNormal {
		return m, nil
	}

	updated, cmd := m.menu.Update(msg)
	if form, ok := updated.(*huh.Form); ok {
		m.menu = form
	}

	if m.menu.State == huh.StateCompleted {
		m.selected = m.menu.GetString("action")
	}

	return m, cmd
}

func (m model) View() tea.View {
	var sb strings.Builder
	sb.WriteString(renderLogo())
	sb.WriteString("\n\n")

	switch m.currentPage {
	case menuMain:
		if m.menu.State == huh.StateCompleted {
			sb.WriteString("Selected: " + m.selected)
		} else {
			sb.WriteString(m.menu.View())
		}
	}

	sb.WriteString("\n\nPress q to quit.\n")

	window := lipgloss.NewStyle().Padding(1)
	app := tea.NewView(window.Render(sb.String()))
	app.AltScreen = true
	return app
}

func RunView() {
	p := tea.NewProgram(initialModel())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Init Tartarus Failed: %v", err)
		os.Exit(1)
	}
}
