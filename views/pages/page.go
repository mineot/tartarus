package pages

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

var QuitKey = key.NewBinding(
	key.WithKeys("q"),
	key.WithHelp("q", "quit"),
)

type Page interface {
	Init() tea.Cmd
	Update(tea.Msg) (Page, tea.Cmd)
	View() string
}

type PageID uint

const (
	PageMenuMain PageID = iota
	PageMenuConfig
)

type NavigateMsg struct {
	To PageID
}
