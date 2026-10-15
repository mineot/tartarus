package components

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

func newShortcutModel(t *testing.T, shortcuts []FormShortcut) formModel {
	t.Helper()

	m, err := newFormModel(FormOpts{
		Fields:    []huh.Field{huh.NewInput().Key("name")},
		Shortcuts: shortcuts,
	})

	if err != nil {
		t.Fatalf("newFormModel: %v", err)
	}

	return m
}

func TestNewFormRequiresFields(t *testing.T) {
	if _, err := newForm(FormOpts{}); err == nil {
		t.Fatal("want an error when no fields are provided")
	}
}

func TestNewFormRejectsNilField(t *testing.T) {
	if _, err := newForm(FormOpts{Fields: []huh.Field{nil}}); err == nil {
		t.Fatal("want an error for an untyped nil field")
	}
}

func TestNewFormRejectsTypedNilField(t *testing.T) {
	var field *huh.Input

	if _, err := newForm(FormOpts{Fields: []huh.Field{field}}); err == nil {
		t.Fatal("want an error for a typed nil field")
	}
}

func TestNewFormRejectsMissingKey(t *testing.T) {
	if _, err := newForm(FormOpts{Fields: []huh.Field{huh.NewInput()}}); err == nil {
		t.Fatal("want an error for a field without a key")
	}
}

func TestNewFormRejectsBlankKey(t *testing.T) {
	if _, err := newForm(FormOpts{Fields: []huh.Field{huh.NewInput().Key("   ")}}); err == nil {
		t.Fatal("want an error for a whitespace-only key")
	}
}

func TestNewFormRejectsDuplicateKey(t *testing.T) {
	fields := []huh.Field{
		huh.NewInput().Key("name"),
		huh.NewInput().Key("name"),
	}

	if _, err := newForm(FormOpts{Fields: fields}); err == nil {
		t.Fatal("want an error for a duplicate key")
	}
}

func TestNewFormAllowsNoteWithoutKey(t *testing.T) {
	fields := []huh.Field{
		huh.NewNote().Title("notice"),
		huh.NewInput().Key("name"),
	}

	if _, err := newForm(FormOpts{Fields: fields}); err != nil {
		t.Fatalf("a note without a key should be allowed: %v", err)
	}
}

func TestNewFormAcceptsValidFields(t *testing.T) {
	fields := []huh.Field{
		huh.NewInput().Key("name"),
		huh.NewConfirm().Key("confirm"),
	}

	form, err := newForm(FormOpts{
		Title:       "Title",
		Description: "Description",
		Fields:      fields,
	})

	if err != nil {
		t.Fatalf("newForm: %v", err)
	}

	if form == nil {
		t.Fatal("want a form")
	}
}

func TestNewShortcutsRejectsEmptyKey(t *testing.T) {
	if _, err := newShortcuts([]FormShortcut{{Label: "save"}}, nil); err == nil {
		t.Fatal("want an error for a shortcut without a key")
	}
}

func TestNewShortcutsRejectsEmptyLabel(t *testing.T) {
	if _, err := newShortcuts([]FormShortcut{{Key: "ctrl+s"}}, nil); err == nil {
		t.Fatal("want an error for a shortcut without a label")
	}
}

func TestNewShortcutsRejectsDuplicateKey(t *testing.T) {
	shortcuts := []FormShortcut{
		{Key: "ctrl+s", Label: "save"},
		{Key: "ctrl+s", Label: "save as"},
	}

	if _, err := newShortcuts(shortcuts, nil); err == nil {
		t.Fatal("want an error for a duplicate key")
	}
}

func TestNewShortcutsRejectsQuitKey(t *testing.T) {
	if _, err := newShortcuts([]FormShortcut{{Key: "ctrl+c", Label: "quit"}}, nil); err == nil {
		t.Fatal("want an error for the quit key")
	}
}

func TestNewShortcutsRejectsFieldKey(t *testing.T) {
	fields := []huh.Field{huh.NewInput().Key("name")}

	if _, err := newForm(FormOpts{Fields: fields}); err != nil {
		t.Fatalf("newForm: %v", err)
	}

	if _, err := newShortcuts([]FormShortcut{{Key: "enter", Label: "go"}}, fields); err == nil {
		t.Fatal("want an error for a key bound by a field")
	}
}

func TestNewShortcutsAcceptsValidKeys(t *testing.T) {
	fields := []huh.Field{huh.NewInput().Key("name")}

	if _, err := newForm(FormOpts{Fields: fields}); err != nil {
		t.Fatalf("newForm: %v", err)
	}

	shortcuts, err := newShortcuts([]FormShortcut{{Key: "ctrl+s", Label: "save"}}, fields)

	if err != nil {
		t.Fatalf("newShortcuts: %v", err)
	}

	if _, ok := shortcuts["ctrl+s"]; !ok {
		t.Fatalf("want the shortcut to be registered, got %v", shortcuts)
	}
}

func TestFormModelRunsShortcut(t *testing.T) {
	ran := false
	focus := ""

	m := newShortcutModel(t, []FormShortcut{{
		Key:   "ctrl+s",
		Label: "save",
		Run: func(got string) (bool, error) {
			ran = true
			focus = got
			return false, nil
		},
	}})

	updated, _ := m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	fm := updated.(formModel)

	if !ran {
		t.Fatal("want the shortcut to run")
	}

	if focus != "name" {
		t.Fatalf("want focus %q, got %q", "name", focus)
	}

	if fm.runErr != nil || fm.aborted {
		t.Fatalf("unexpected state: runErr=%v aborted=%v", fm.runErr, fm.aborted)
	}
}

func TestFormModelShortcutQuitAborts(t *testing.T) {
	m := newShortcutModel(t, []FormShortcut{{
		Key:   "ctrl+x",
		Label: "exit",
		Run:   func(string) (bool, error) { return true, nil },
	}})

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'x', Mod: tea.ModCtrl})
	fm := updated.(formModel)

	if !fm.aborted {
		t.Fatal("want the shortcut to abort the form")
	}
}

func TestFormModelShortcutErrorIsKept(t *testing.T) {
	want := errors.New("boom")

	m := newShortcutModel(t, []FormShortcut{{
		Key:   "ctrl+b",
		Label: "explode",
		Run:   func(string) (bool, error) { return false, want },
	}})

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'b', Mod: tea.ModCtrl})
	fm := updated.(formModel)

	if !errors.Is(fm.runErr, want) {
		t.Fatalf("want %v, got %v", want, fm.runErr)
	}
}

func TestFormModelNilRunIsDisplayOnly(t *testing.T) {
	m := newShortcutModel(t, []FormShortcut{{Key: "f1", Label: "help"}})

	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyF1})
	fm := updated.(formModel)

	if cmd != nil {
		t.Fatal("want no command for a display-only shortcut")
	}

	if fm.runErr != nil || fm.aborted {
		t.Fatalf("unexpected state: runErr=%v aborted=%v", fm.runErr, fm.aborted)
	}
}

func TestFormModelViewShowsShortcuts(t *testing.T) {
	m := newShortcutModel(t, []FormShortcut{{Key: "ctrl+s", Label: "save"}})

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	content := updated.(formModel).View().Content

	for _, want := range []string{"ctrl+s", "save", "ctrl+c", "quit"} {
		if !strings.Contains(content, want) {
			t.Fatalf("want %q in the footer, got %q", want, content)
		}
	}
}

func TestFormModelFooterOrder(t *testing.T) {
	m := newShortcutModel(t, []FormShortcut{{Key: "ctrl+s", Label: "save"}})

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	footer := updated.(formModel).footerView()

	field := strings.Index(footer, "submit")
	shortcut := strings.Index(footer, "save")
	quit := strings.Index(footer, "quit")

	if field < 0 || shortcut < 0 || quit < 0 {
		t.Fatalf("want field, shortcut and quit in the footer, got %q", footer)
	}

	if !(field < shortcut && shortcut < quit) {
		t.Fatalf("want field < shortcut < quit, got %q", footer)
	}
}

func TestFormThemeIsNotNil(t *testing.T) {
	for _, dark := range []bool{true, false} {
		if formTheme(dark) == nil {
			t.Fatalf("want a theme for dark=%v", dark)
		}
	}
}

func TestFormRejectsNilContext(t *testing.T) {
	var ctx context.Context

	_, err := Form(ctx, FormOpts{Fields: []huh.Field{huh.NewInput().Key("name")}})

	if err == nil {
		t.Fatal("want an error for a nil context")
	}

	if !strings.Contains(err.Error(), "context is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFormRejectsCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := Form(ctx, FormOpts{Fields: []huh.Field{huh.NewInput().Key("name")}})

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("want context.Canceled, got %v", err)
	}
}
