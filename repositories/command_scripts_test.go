package repositories

import (
	"errors"
	"testing"

	"tartarus/store"
)

func TestGetCommandScriptsByName(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "deploy"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	for _, name := range []string{"buildimage", "pushimage", "restart"} {
		script := &CommandScript{Name: name, Script: "run " + name}
		if err := r.AppendCommandScript(c.ID, script); err != nil {
			t.Fatalf("AppendCommandScript: %v", err)
		}
	}

	got, err := r.GetCommandScriptsByName(c.ID, "image")
	if err != nil {
		t.Fatalf("GetCommandScriptsByName: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("want 2 scripts, got %d", len(got))
	}

	if got[0].Name != "buildimage" || got[1].Name != "pushimage" {
		t.Fatalf("unexpected scripts: %+v", got)
	}

	all, err := r.GetCommandScriptsByName(c.ID, "")
	if err != nil {
		t.Fatalf("GetCommandScriptsByName: %v", err)
	}

	if len(all) != 3 {
		t.Fatalf("want 3 scripts, got %d", len(all))
	}
}

func TestGetCommandScriptsByNameRequiresCommandID(t *testing.T) {
	r := New(newTestStore(t))

	_, err := r.GetCommandScriptsByName(0, "x")

	if !errors.Is(err, ErrCommandIDRequired) {
		t.Fatalf("want ErrCommandIDRequired, got %v", err)
	}
}

func TestGetCommandScriptsByNamePropagatesAClosedStore(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	_, err := r.GetCommandScriptsByName(1, "x")

	if !errors.Is(err, store.ErrClosed) {
		t.Fatalf("want store.ErrClosed, got %v", err)
	}
}

func TestGetCommandScriptsByNameIsRejectedInsideATransaction(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	err := s.WithTx(func(tx *store.Tx) error {
		_, err := r.GetCommandScriptsByName(1, "x")
		return err
	})

	if !errors.Is(err, store.ErrUseTx) {
		t.Fatalf("want store.ErrUseTx, got %v", err)
	}
}

func TestValidateCommandScriptName(t *testing.T) {
	for _, tc := range []struct {
		name    string
		wantErr error
	}{
		{"build", nil},
		{"build1", nil},
		{"a12345678901", nil},
		{"", ErrCommandScriptNameRequired},
		{"a123456789012", ErrCommandScriptNameTooLong},
		{"Build", ErrCommandScriptNameInvalid},
		{"build-1", ErrCommandScriptNameInvalid},
		{"build_1", ErrCommandScriptNameInvalid},
		{"build 1", ErrCommandScriptNameInvalid},
	} {
		if err := ValidateCommandScriptName(tc.name); !errors.Is(err, tc.wantErr) {
			t.Fatalf("ValidateCommandScriptName(%q): want %v, got %v", tc.name, tc.wantErr, err)
		}
	}
}

func TestAppendCommandScriptRequiresName(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "deploy"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	err := r.AppendCommandScript(c.ID, &CommandScript{Script: "ls"})

	if !errors.Is(err, ErrCommandScriptNameRequired) {
		t.Fatalf("want ErrCommandScriptNameRequired, got %v", err)
	}
}

func TestUpdateCommandScriptRequiresName(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "deploy"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	script := &CommandScript{Name: "lista", Script: "ls"}
	if err := r.AppendCommandScript(c.ID, script); err != nil {
		t.Fatalf("AppendCommandScript: %v", err)
	}

	script.Name = ""

	err := r.UpdateCommandScript(c.ID, script)

	if !errors.Is(err, ErrCommandScriptNameRequired) {
		t.Fatalf("want ErrCommandScriptNameRequired, got %v", err)
	}
}

func TestAppendCommandScriptRejectsInvalidName(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "deploy"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	err := r.AppendCommandScript(c.ID, &CommandScript{Name: "build-1", Script: "ls"})

	if !errors.Is(err, ErrCommandScriptNameInvalid) {
		t.Fatalf("want ErrCommandScriptNameInvalid, got %v", err)
	}
}
