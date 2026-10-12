package pageforms

import (
	"errors"
	"tartarus/repositories"
	"tartarus/views/pages"
	"testing"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
)

func TestNewCommandNavigatesWithoutID(t *testing.T) {
	m := NewCommandList(&repositories.Repos{})
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'n', Mod: tea.ModCtrl})
	if cmd == nil {
		t.Fatal("expected navigation command")
	}
	msg, ok := cmd().(pages.NavigateMsg)
	if !ok || msg.To != pages.CommandEditor || msg.RefID != 0 {
		t.Fatalf("unexpected navigation: %#v", msg)
	}
}

func TestEditCommandCapturesSelectedID(t *testing.T) {
	m := NewCommandList(&repositories.Repos{})
	m.listModel.List.SetSize(80, 12)
	m.listModel.List.SetItems([]list.Item{
		commandListItem{command: repositories.Command{ID: 41, Name: "first"}},
		commandListItem{command: repositories.Command{ID: 93, Name: "second"}},
	})
	m.listModel.List.Select(1)
	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected navigation command")
	}
	m.listModel.List.Select(0)
	msg, ok := cmd().(pages.NavigateMsg)
	if !ok || msg.To != pages.CommandEditor || msg.RefID != 93 {
		t.Fatalf("unexpected navigation: %#v", msg)
	}
}

func TestEditCommandRequiresAvailableSelection(t *testing.T) {
	for _, state := range []string{"empty", "loading", "error", "zero ID"} {
		t.Run(state, func(t *testing.T) {
			m := NewCommandList(&repositories.Repos{})
			if state != "empty" {
				id := uint64(41)
				if state == "zero ID" {
					id = 0
				}
				m.listModel.List.SetItems([]list.Item{commandListItem{command: repositories.Command{ID: id}}})
			}
			m.listModel.Loading = state == "loading"
			if state == "error" {
				m.listModel.Err = errors.New("query failed")
			}
			updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
			if updated != m || cmd != nil {
				t.Fatal("edit navigated without an available selection")
			}
		})
	}
}

func TestSearchEnterAppliesSearchInsteadOfEditingCommand(t *testing.T) {
	m := NewCommandList(&repositories.Repos{})
	m.listModel.Sized, m.listModel.Size = true, 5
	m.listModel.List.SetItems([]list.Item{commandListItem{command: repositories.Command{ID: 41}}})
	m.Update(tea.KeyPressMsg{Code: '/', Text: "/"})
	m.listModel.SearchInput.SetValue("build")
	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if updated != m || cmd == nil || m.listModel.Search != "build" || !m.listModel.Loading || m.listModel.SearchEditing {
		t.Fatal("enter did not apply search")
	}
}

func TestNewCommandDoesNotNavigateWhileSearching(t *testing.T) {
	m := NewCommandList(&repositories.Repos{})
	m.Update(tea.KeyPressMsg{Code: '/', Text: "/"})
	updated, cmd := m.Update(tea.KeyPressMsg{Code: 'n', Mod: tea.ModCtrl})
	if updated != m || !m.listModel.SearchEditing {
		t.Fatal("new command interrupted search editing")
	}
	if cmd != nil {
		if _, ok := cmd().(pages.NavigateMsg); ok {
			t.Fatal("new command navigated while searching")
		}
	}
}
