package pages

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestShortcutsLeaveKeysToEditor(t *testing.T) {
	shortcuts := Shortcuts{QuitShortcut, MenuConfigsShortcut}
	for _, msg := range []tea.KeyPressMsg{
		{Code: 'q', Text: "q"},
		{Code: tea.KeyEscape},
	} {
		cmd, handled := shortcuts.Handle(msg, true)
		if handled || cmd != nil {
			t.Fatalf("page shortcut intercepted %q while editing", msg.String())
		}
	}
}

func TestShortcutsWorkOutsideEditor(t *testing.T) {
	shortcuts := Shortcuts{QuitShortcut, MenuConfigsShortcut}
	cmd, handled := shortcuts.Handle(tea.KeyPressMsg{Code: 'q', Text: "q"}, false)
	if !handled || cmd == nil {
		t.Fatal("quit was not handled")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatal("expected quit message")
	}
	cmd, handled = shortcuts.Handle(tea.KeyPressMsg{Code: tea.KeyEscape}, false)
	if !handled || cmd == nil {
		t.Fatal("escape was not handled")
	}
	if msg, ok := cmd().(NavigateMsg); !ok || msg.To != MenuConfigurations {
		t.Fatal("expected navigation message")
	}
	if cmd, handled := shortcuts.Handle(tea.WindowSizeMsg{}, false); handled || cmd != nil {
		t.Fatal("handled a non-key message")
	}
}
