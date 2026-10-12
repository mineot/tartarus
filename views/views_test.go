package views

import (
	"tartarus/repositories"
	"tartarus/views/pages"
	pageforms "tartarus/views/pages/forms"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestNavigationSuppliesExistingDimensions(t *testing.T) {
	m := initialModel(&repositories.Repos{})
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
	m = updated.(model)
	updated, cmd := m.Update(pages.NavigateMsg{To: pages.CommandList})
	m = updated.(model)
	if _, ok := m.page.(*pageforms.CommandList); !ok {
		t.Fatalf("unexpected page type %T", m.page)
	}
	if cmd == nil {
		t.Fatal("expected initial query without another terminal resize")
	}
	_, cmd = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("expected navigation on escape")
	}
	if msg, ok := cmd().(pages.NavigateMsg); !ok || msg.To != pages.MenuConfigurations {
		t.Fatal("escape did not return to configurations")
	}
}

func TestCommandEditorNavigationKeepsPageAfterResize(t *testing.T) {
	for _, id := range []uint64{0, 93} {
		m := initialModel(&repositories.Repos{})
		updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
		m = updated.(model)
		updated, _ = m.Update(pages.NavigateMsg{To: pages.CommandEditor, RefID: id})
		m = updated.(model)
		if _, ok := m.page.(*pageforms.CommandEditor); !ok {
			t.Fatalf("unexpected editor page: %T", m.page)
		}
		m.View()
	}
}

func TestControlCQuitsWhileEditing(t *testing.T) {
	m := initialModel(&repositories.Repos{})
	m.page = pageforms.NewCommandList(&repositories.Repos{})
	m.page, _ = m.page.Update(tea.KeyPressMsg{Code: '/', Text: "/"})
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	if cmd == nil {
		t.Fatal("expected global quit while editing")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatal("expected quit message")
	}
}

func TestSearchCursorTracksEditing(t *testing.T) {
	m := initialModel(&repositories.Repos{})
	m.page = pageforms.NewCommandList(&repositories.Repos{})
	m.page, _ = m.page.Update(tea.WindowSizeMsg{Width: 78, Height: 25})
	if m.View().Cursor != nil {
		t.Fatal("expected no cursor outside search editing")
	}

	m.page, _ = m.page.Update(tea.KeyPressMsg{Code: '/', Text: "/"})
	assertCursor := func(x, y int) {
		t.Helper()
		cursor := m.View().Cursor
		if cursor == nil {
			t.Fatal("expected a terminal cursor while editing")
		}
		if cursor.X != x || cursor.Y != y {
			t.Fatalf("cursor = (%d, %d), want (%d, %d)", cursor.X, cursor.Y, x, y)
		}
	}
	assertCursor(9, 4)
	assertCursor(9, 4)
	m.page, _ = m.page.Update(tea.KeyPressMsg{Code: 'a', Text: "a"})
	assertCursor(10, 4)
	m.page, _ = m.page.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
	assertCursor(9, 4)
	m.page, _ = m.page.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if m.View().Cursor != nil {
		t.Fatal("expected no cursor after canceling search")
	}
}
