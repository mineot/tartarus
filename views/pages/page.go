package pages

import tea "charm.land/bubbletea/v2"

// Page owns the state, updates, and content of a screen.
type Page interface {
	Init() tea.Cmd
	Update(tea.Msg) (Page, tea.Cmd)
	View() string
}

// PageID identifies a navigation destination.
type PageID uint

const (
	// PageMenuMain identifies the main menu.
	PageMenuMain PageID = iota
)

// NavigateMsg asks the application to open a page.
type NavigateMsg struct {
	To PageID
}
