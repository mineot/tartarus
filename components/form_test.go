package components

import (
	"context"
	"errors"
	"strings"
	"testing"

	"charm.land/huh/v2"
)

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
