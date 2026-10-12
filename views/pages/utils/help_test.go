package pageutils

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestRenderHelpWrapsBetweenBindings(t *testing.T) {
	disabled := key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "hidden"))
	disabled.SetEnabled(false)
	bindings := []key.Binding{
		key.NewBinding(key.WithKeys("up"), key.WithHelp("↑", "up")),
		disabled,
		key.NewBinding(key.WithKeys("down"), key.WithHelp("↓", "down")),
		key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
	}
	for _, tc := range []struct {
		width int
		want  string
	}{
		{80, "↑ up • ↓ down • esc back"},
		{13, "↑ up • ↓ down\nesc back"},
		{8, "↑ up\n↓ down\nesc back"},
		{0, "↑ up • ↓ down • esc back"},
	} {
		got := ansi.Strip(RenderHelp(bindings, nil, tc.width))
		if got != tc.want {
			t.Errorf("width %d: got %q, want %q", tc.width, got, tc.want)
		}
	}
}

func TestRenderHelpWrapsOversizedBinding(t *testing.T) {
	bindings := []key.Binding{
		key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "clear search")),
		key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
	}
	got := RenderHelp(bindings, nil, 8)
	for _, line := range strings.Split(got, "\n") {
		if lipgloss.Width(line) > 8 {
			t.Fatalf("help line exceeds available width: %q", line)
		}
	}
	plain := strings.Join(strings.Fields(ansi.Strip(got)), " ")
	if plain != "esc clear search q quit" {
		t.Fatalf("unexpected help content: %q", plain)
	}
}
