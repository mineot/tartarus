package views

import (
	"tartarus/repositories"
	"tartarus/views/pages"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestNavigationSuppliesExistingDimensions(t *testing.T) {
	m := initialModel(&repositories.Repos{})
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
	m = updated.(model)
	updated, cmd := m.Update(pages.NavigateMsg{To: pages.PageMenuCmds})
	m = updated.(model)
	if _, ok := m.page.(*pages.MenuCmds); !ok {
		t.Fatalf("unexpected page type %T", m.page)
	}
	if cmd == nil {
		t.Fatal("expected initial query without another terminal resize")
	}
	_, cmd = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("expected navigation on escape")
	}
	if msg, ok := cmd().(pages.NavigateMsg); !ok || msg.To != pages.PageMenuConfig {
		t.Fatal("escape did not return to configurations")
	}
}
