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
	MenuMain PageID = iota
	MenuConfigurations
	CommandList
	CommandEditor
	ManualList
	StorageList
)

type NavigateMsg struct {
	To PageID
}
