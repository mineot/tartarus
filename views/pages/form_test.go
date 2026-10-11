package pages

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

func TestFormEditingProtectsTextFields(t *testing.T) {
	for _, tc := range []struct {
		name  string
		field huh.Field
	}{
		{"input", huh.NewInput()},
		{"text", huh.NewText()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			form := NewForm(huh.NewGroup(tc.field))
			form.Init()
			if !FormEditing(form) {
				t.Fatal("text field was not recognized")
			}
			msg := tea.KeyPressMsg{Code: 'q', Text: "q"}
			if _, handled := (Shortcuts{QuitShortcut}).Handle(msg, FormEditing(form)); handled {
				t.Fatal("quit intercepted text entry")
			}
			form, _, _ = UpdateForm(form, msg)
			if got := form.GetFocusedField().GetValue(); got != "q" {
				t.Fatalf("want text q, got %v", got)
			}
			if strings.Contains(FormHelp(form, Shortcuts{QuitShortcut}), "quit") {
				t.Fatal("help advertised disabled quit shortcut")
			}
			form.State = huh.StateCompleted
			if FormEditing(form) {
				t.Fatal("completed form is still editing")
			}
		})
	}
}

func TestFormEditingKeepsMenuShortcuts(t *testing.T) {
	form := NewForm(huh.NewGroup(huh.NewSelect[string]().Options(huh.NewOption("Commands", "commands"))))
	form.Init()
	if FormEditing(form) {
		t.Fatal("menu selection treated as text editing")
	}
	if !strings.Contains(FormHelp(form, Shortcuts{QuitShortcut}), "quit") {
		t.Fatal("menu help lost quit shortcut")
	}
	if FormEditing(nil) {
		t.Fatal("nil form is editing")
	}
}
