package views

import (
	"fmt"
	"os"
	"strings"
	"tartarus/views/pages"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type model struct {
	page pages.Page
}

func initialModel() model {
	return model{
		page: pages.NewMenuMain(),
	}
}

func (m model) Init() tea.Cmd {
	return m.page.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case pages.NavigateMsg:
		switch msg.To {
		case pages.PageMenuMain:
			m.page = pages.NewMenuMain()

		case pages.PageMenuConfig:
			m.page = pages.NewMenuConfig()

		default:
			return m, nil
		}

		return m, m.page.Init()
	}

	var cmd tea.Cmd
	m.page, cmd = m.page.Update(msg)
	return m, cmd
}

func (m model) View() tea.View {
	var sb strings.Builder
	sb.WriteString(renderLogo())
	sb.WriteString("\n\n")
	sb.WriteString(m.page.View())

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
