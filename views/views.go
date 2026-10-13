package views

import (
	"strings"
	"tartarus/repositories"
	"tartarus/views/pages"
	pageforms "tartarus/views/pages/forms"
	pagemenus "tartarus/views/pages/menus"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type model struct {
	page          pages.Page
	repos         *repositories.Repos
	width, height int
	sized         bool
}

func initialModel(r *repositories.Repos) model {
	return model{
		page:  pagemenus.NewMenuMain(),
		repos: r,
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

	case tea.WindowSizeMsg:
		m.width, m.height, m.sized = msg.Width, msg.Height, true
		cmd := m.resizePage()
		return m, cmd

	case pages.NavigateMsg:
		switch msg.To {
		case pages.MenuMain:
			m.page = pagemenus.NewMenuMain()

		case pages.MenuConfigurations:
			m.page = pagemenus.NewMenuConfigurations()

		case pages.MenuCommands:
			m.page = pagemenus.NewMenuCommands()

		case pages.CommandList:
			m.page = pageforms.NewCommandList(m.repos)

		case pages.CommandEditor:
			m.page = pageforms.NewCommandEditor(m.repos, msg.RefID)

		case pages.ManualList:
			m.page = pageforms.NewMenuMans()

		case pages.StorageList:
			m.page = pageforms.NewMenuStg()

		default:
			return m, nil
		}

		resize := m.resizePage()
		return m, tea.Batch(resize, m.page.Init())
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

	if page, ok := m.page.(interface{ Cursor() *tea.Cursor }); ok {
		app.Cursor = page.Cursor()

		if app.Cursor != nil {
			app.Cursor.X += window.GetPaddingLeft()
			app.Cursor.Y += window.GetPaddingTop() + lipgloss.Height(renderLogo()) + 1
		}
	}

	return app
}

func (m *model) resizePage() tea.Cmd {
	if !m.sized {
		return nil
	}

	var cmd tea.Cmd

	m.page, cmd = m.page.Update(tea.WindowSizeMsg{
		Width:  max(1, m.width-2),
		Height: max(1, m.height-2-lipgloss.Height(renderLogo())-2),
	})

	return cmd
}

func RunView(r *repositories.Repos) error {
	p := tea.NewProgram(initialModel(r))
	_, err := p.Run()
	return err
}
