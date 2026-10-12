package pageforms

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestCommandEditorRetainsIDOnResize(t *testing.T) {
	for _, id := range []uint64{0, 93} {
		m := NewCommandEditor(nil, id)
		updated, cmd := m.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
		if updated != m || cmd != nil || m.command.ID != id {
			t.Fatalf("editor did not retain ID %d", id)
		}
	}
}
