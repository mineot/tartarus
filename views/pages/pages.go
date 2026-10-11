package pages

import (
	tea "charm.land/bubbletea/v2"
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
	PageMenuCmds
	PageMenuMans
	PageMenuStg
)

type NavigateMsg struct {
	To PageID
}
