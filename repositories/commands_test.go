package repositories

import (
	"errors"
	"testing"
	"time"

	"tartarus/store"
)

func insertTestCommand(t *testing.T, s *store.Store, name string, description string) (uint64, time.Time) {
	t.Helper()

	now := time.Now().UTC()

	var id int64

	err := s.WithTx(func(tx *store.Tx) error {
		result, err := tx.Exec(`INSERT INTO commands (name, description, created_at, updated_at) VALUES (?, ?, ?, ?)`, name, nullable(description), now, now)

		if err != nil {
			return err
		}

		id, err = result.LastInsertId()

		return err
	})

	if err != nil {
		t.Fatalf("inserting command %q: %v", name, err)
	}

	return uint64(id), now
}

func insertTestCommandScript(t *testing.T, s *store.Store, commandID uint64, name string, script string, description string) (uint64, time.Time) {
	t.Helper()

	now := time.Now().UTC()

	var id int64

	err := s.WithTx(func(tx *store.Tx) error {
		result, err := tx.Exec(`INSERT INTO command_scripts (command_id, name, script, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`, commandID, name, script, nullable(description), now, now)

		if err != nil {
			return err
		}

		id, err = result.LastInsertId()

		return err
	})

	if err != nil {
		t.Fatalf("inserting command script %q: %v", script, err)
	}

	return uint64(id), now
}

func nullable(s string) any {
	if s == "" {
		return nil
	}

	return s
}

func TestGetCommandsOnEmptyDatabase(t *testing.T) {
	s := newTestStore(t)

	r := New(s)
	commands, err := r.GetCommands("")

	if err != nil {
		t.Fatalf("GetCommands: %v", err)
	}

	if len(commands) != 0 {
		t.Fatalf("want zero commands, got %d", len(commands))
	}
}

func TestGetCommandsReturnsEveryCommandInCreationOrder(t *testing.T) {
	s := newTestStore(t)

	type inserted struct {
		id          uint64
		name        string
		description string
		at          time.Time
	}

	want := make([]inserted, 0, 3)

	for _, m := range []struct {
		name        string
		description string
	}{
		{"zebra", "ultimo alfabeticamente"},
		{"alpha", "primeiro alfabeticamente"},
		{"mango", "no meio"},
	} {
		id, at := insertTestCommand(t, s, m.name, m.description)
		want = append(want, inserted{id: id, name: m.name, description: m.description, at: at})
	}

	r := New(s)
	commands, err := r.GetCommands("")

	if err != nil {
		t.Fatalf("GetCommands: %v", err)
	}

	if len(commands) != len(want) {
		t.Fatalf("want %d commands, got %d", len(want), len(commands))
	}

	for i, w := range want {
		got := commands[i]

		if got.ID != w.id {
			t.Fatalf("position %d: want id %d, got %d", i, w.id, got.ID)
		}
		if got.Name != w.name {
			t.Fatalf("position %d: want name %q, got %q", i, w.name, got.Name)
		}
		if got.Description != w.description {
			t.Fatalf("position %d: want description %q, got %q", i, w.description, got.Description)
		}
		if got.CreatedAt.Unix() != w.at.Unix() {
			t.Fatalf("position %d: want created_at %d, got %d", i, w.at.Unix(), got.CreatedAt.Unix())
		}
		if got.UpdatedAt.Unix() != w.at.Unix() {
			t.Fatalf("position %d: want updated_at %d, got %d", i, w.at.Unix(), got.UpdatedAt.Unix())
		}
	}
}

func TestGetCommandsIsRejectedInsideATransaction(t *testing.T) {
	s := newTestStore(t)

	err := s.WithTx(func(tx *store.Tx) error {
		r := New(s)
		_, err := r.GetCommands("")
		return err
	})

	if !errors.Is(err, store.ErrUseTx) {
		t.Fatalf("want store.ErrUseTx, got %v", err)
	}
}

func TestGetCommandsPropagatesAClosedStore(t *testing.T) {
	s := newTestStore(t)

	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	r := New(s)
	_, err := r.GetCommands("")

	if !errors.Is(err, store.ErrClosed) {
		t.Fatalf("want store.ErrClosed, got %v", err)
	}
}

func TestGetCommandFound(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	id, at := insertTestCommand(t, s, "build", "compila o projeto")

	got, err := r.GetCommand("build")
	if err != nil {
		t.Fatalf("GetCommand: %v", err)
	}

	if got.ID != id {
		t.Fatalf("want ID %d, got %d", id, got.ID)
	}
	if got.Name != "build" {
		t.Fatalf("want name %q, got %q", "build", got.Name)
	}
	if got.Description != "compila o projeto" {
		t.Fatalf("want description %q, got %q", "compila o projeto", got.Description)
	}
	if got.CreatedAt.Unix() != at.Unix() {
		t.Fatalf("want created_at %d, got %d", at.Unix(), got.CreatedAt.Unix())
	}
	if got.UpdatedAt.Unix() != at.Unix() {
		t.Fatalf("want updated_at %d, got %d", at.Unix(), got.UpdatedAt.Unix())
	}
}

func TestGetCommandReadsANullDescriptionAsEmpty(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	insertTestCommand(t, s, "legacy", "")

	got, err := r.GetCommand("legacy")
	if err != nil {
		t.Fatalf("GetCommand: %v", err)
	}

	if got.Description != "" {
		t.Fatalf("want an empty description for NULL, got %q", got.Description)
	}
}

func TestGetCommandNotFound(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	got, err := r.GetCommand("missing")
	if err != nil {
		t.Fatalf("GetCommand: %v", err)
	}
	if got.ID != 0 || got.Name != "" || got.Description != "" {
		t.Fatalf("want an empty command, got %+v", got)
	}
}

func TestGetCommandPropagatesAClosedStore(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	_, err := r.GetCommand("x")

	if !errors.Is(err, store.ErrClosed) {
		t.Fatalf("want store.ErrClosed, got %v", err)
	}
}

func TestGetCommandIsRejectedInsideATransaction(t *testing.T) {
	s := newTestStore(t)

	err := s.WithTx(func(tx *store.Tx) error {
		r := New(s)
		_, err := r.GetCommand("x")
		return err
	})

	if !errors.Is(err, store.ErrUseTx) {
		t.Fatalf("want store.ErrUseTx, got %v", err)
	}
}

func TestInsertCommandSuccess(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "run", Description: "executa o projeto"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	if c.ID == 0 {
		t.Fatal("want ID populated")
	}
	if c.Name != "run" {
		t.Fatalf("want name %q, got %q", "run", c.Name)
	}
	if c.Description != "executa o projeto" {
		t.Fatalf("want description %q, got %q", "executa o projeto", c.Description)
	}
	if c.CreatedAt.IsZero() {
		t.Fatal("want CreatedAt populated")
	}
	if c.UpdatedAt.IsZero() {
		t.Fatal("want UpdatedAt populated")
	}
	if !c.CreatedAt.Equal(c.UpdatedAt) {
		t.Fatalf("want CreatedAt == UpdatedAt on insert, got %v vs %v", c.CreatedAt, c.UpdatedAt)
	}

	got, err := r.GetCommand("run")
	if err != nil {
		t.Fatalf("GetCommand: %v", err)
	}
	if got.ID != c.ID {
		t.Fatalf("want ID %d, got %d", c.ID, got.ID)
	}
	if got.Name != "run" {
		t.Fatalf("want name %q, got %q", "run", got.Name)
	}
	if got.Description != "executa o projeto" {
		t.Fatalf("want description %q, got %q", "executa o projeto", got.Description)
	}
}

func TestInsertCommandWithoutDescription(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "run"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	if c.Description != "" {
		t.Fatalf("want an empty description, got %q", c.Description)
	}

	got, err := r.GetCommand("run")
	if err != nil {
		t.Fatalf("GetCommand: %v", err)
	}
	if got.Description != "" {
		t.Fatalf("want an empty description persisted, got %q", got.Description)
	}
}

func TestInsertCommandMissingName(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: ""}
	err := r.InsertCommand(c)
	if !errors.Is(err, ErrCommandNameRequired) {
		t.Fatalf("want ErrCommandNameRequired, got %v", err)
	}

	commands, err := r.GetCommands("")
	if err != nil {
		t.Fatalf("GetCommands: %v", err)
	}
	if len(commands) != 0 {
		t.Fatalf("want the database unchanged, got %d commands", len(commands))
	}
}

func TestInsertCommandPropagatesAClosedStore(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	err := r.InsertCommand(&Command{Name: "x"})
	if !errors.Is(err, store.ErrClosed) {
		t.Fatalf("want store.ErrClosed, got %v", err)
	}
}

func TestInsertCommandIsRejectedInsideATransaction(t *testing.T) {
	s := newTestStore(t)

	err := s.WithTx(func(tx *store.Tx) error {
		r := New(s)
		return r.InsertCommand(&Command{Name: "inside"})
	})

	if !errors.Is(err, store.ErrTxActive) {
		t.Fatalf("want store.ErrTxActive, got %v", err)
	}
}

func TestInsertCommandNameUniqueness(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	if err := r.InsertCommand(&Command{Name: "dup"}); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	err := r.InsertCommand(&Command{Name: "dup"})
	if err == nil {
		t.Fatal("want a UNIQUE error")
	}
	if errors.Is(err, ErrCommandNameRequired) {
		t.Fatalf("want a UNIQUE violation, got %v", err)
	}
}

func TestUpdateCommandSuccess(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "old", Description: "descricao antiga"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}
	id := c.ID
	created := c.CreatedAt
	updatedBefore := c.UpdatedAt

	time.Sleep(time.Microsecond)

	c.Name = "new"
	c.Description = "descricao nova"
	if err := r.UpdateCommand(c); err != nil {
		t.Fatalf("UpdateCommand: %v", err)
	}

	if c.ID != id {
		t.Fatalf("want ID preserved as %d, got %d", id, c.ID)
	}
	if c.Name != "new" {
		t.Fatalf("want name %q, got %q", "new", c.Name)
	}
	if c.Description != "descricao nova" {
		t.Fatalf("want description %q, got %q", "descricao nova", c.Description)
	}
	if !c.CreatedAt.Equal(created) {
		t.Fatalf("want CreatedAt preserved as %v, got %v", created, c.CreatedAt)
	}
	if c.UpdatedAt.Unix() < updatedBefore.Unix() {
		t.Fatalf("want UpdatedAt updated")
	}

	got, err := r.GetCommand("new")
	if err != nil {
		t.Fatalf("GetCommand: %v", err)
	}
	if got.Name != "new" {
		t.Fatalf("want name persisted as %q, got %q", "new", got.Name)
	}
	if got.Description != "descricao nova" {
		t.Fatalf("want description persisted as %q, got %q", "descricao nova", got.Description)
	}
}

func TestUpdateCommandClearsTheDescription(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "run", Description: "descricao que sera removida"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	c.Description = ""
	if err := r.UpdateCommand(c); err != nil {
		t.Fatalf("UpdateCommand: %v", err)
	}
	if c.Description != "" {
		t.Fatalf("want an empty description on the struct, got %q", c.Description)
	}

	got, err := r.GetCommand("run")
	if err != nil {
		t.Fatalf("GetCommand: %v", err)
	}
	if got.Description != "" {
		t.Fatalf("want an empty description persisted, got %q", got.Description)
	}
}

func TestUpdateCommandIDRequired(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{ID: 0, Name: "x"}
	err := r.UpdateCommand(c)
	if !errors.Is(err, ErrCommandIDRequired) {
		t.Fatalf("want ErrCommandIDRequired, got %v", err)
	}
}

func TestUpdateCommandMissingName(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "valid"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	c.Name = ""
	err := r.UpdateCommand(c)
	if !errors.Is(err, ErrCommandNameRequired) {
		t.Fatalf("want ErrCommandNameRequired, got %v", err)
	}
}

func TestUpdateCommandNotFound(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{ID: 999, Name: "ghost"}
	err := r.UpdateCommand(c)
	if !errors.Is(err, ErrCommandNotFound) {
		t.Fatalf("want ErrCommandNotFound, got %v", err)
	}
}

func TestUpdateCommandPropagatesAClosedStore(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "temp"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	c.Name = "changed"
	err := r.UpdateCommand(c)
	if !errors.Is(err, store.ErrClosed) {
		t.Fatalf("want store.ErrClosed, got %v", err)
	}
}

func TestUpdateCommandIsRejectedInsideATransaction(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "temp"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	err := s.WithTx(func(tx *store.Tx) error {
		c.Name = "inside"
		return r.UpdateCommand(c)
	})

	if !errors.Is(err, store.ErrTxActive) {
		t.Fatalf("want store.ErrTxActive, got %v", err)
	}
}

func TestUpdateCommandNameUniqueness(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	if err := r.InsertCommand(&Command{Name: "a"}); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}
	c := &Command{Name: "b"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	c.Name = "a"
	err := r.UpdateCommand(c)
	if err == nil {
		t.Fatal("want a UNIQUE error")
	}
}

func TestDeleteCommandSuccess(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "target"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	insertTestCommandScript(t, s, c.ID, "echo1", "echo 1", "primeiro item")
	insertTestCommandScript(t, s, c.ID, "echo2", "echo 2", "segundo item")

	if err := r.DeleteCommand(c.ID); err != nil {
		t.Fatalf("DeleteCommand: %v", err)
	}

	got, err := r.GetCommand("target")
	if err != nil {
		t.Fatalf("GetCommand: %v", err)
	}
	if got.ID != 0 || got.Name != "" {
		t.Fatalf("want the command removed, got %+v", got)
	}

	items, err := r.GetCommandScripts(c.ID)
	if err != nil {
		t.Fatalf("GetCommandScripts: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("want the items removed by CASCADE, got %d", len(items))
	}
}

func TestDeleteCommandIDRequired(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	err := r.DeleteCommand(0)
	if !errors.Is(err, ErrCommandIDRequired) {
		t.Fatalf("want ErrCommandIDRequired, got %v", err)
	}
}

func TestDeleteCommandNotFoundIsIdempotent(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	err := r.DeleteCommand(999)
	if err != nil {
		t.Fatalf("want nil (idempotent), got %v", err)
	}
}

func TestDeleteCommandPropagatesAClosedStore(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "temp"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	err := r.DeleteCommand(c.ID)
	if !errors.Is(err, store.ErrClosed) {
		t.Fatalf("want store.ErrClosed, got %v", err)
	}
}

func TestDeleteCommandIsRejectedInsideATransaction(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "temp"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	err := s.WithTx(func(tx *store.Tx) error {
		return r.DeleteCommand(c.ID)
	})

	if !errors.Is(err, store.ErrTxActive) {
		t.Fatalf("want store.ErrTxActive, got %v", err)
	}
}

func TestGetCommandScriptsEmpty(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "cmd"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	items, err := r.GetCommandScripts(c.ID)
	if err != nil {
		t.Fatalf("GetCommandScripts: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("want zero items, got %d", len(items))
	}
}

func TestGetCommandScriptsReturnsAllInOrder(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "cmd"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	type inserted struct {
		id          uint64
		commandID   uint64
		script      string
		description string
		at          time.Time
	}

	want := make([]inserted, 0, 3)

	for _, m := range []struct{ script, description string }{
		{"z", "descricao de z"},
		{"a", ""},
		{"m", "descricao de m"},
	} {
		id, at := insertTestCommandScript(t, s, c.ID, m.script, m.script, m.description)
		want = append(want, inserted{id: id, commandID: c.ID, script: m.script, description: m.description, at: at})
	}

	items, err := r.GetCommandScripts(c.ID)
	if err != nil {
		t.Fatalf("GetCommandScripts: %v", err)
	}

	if len(items) != len(want) {
		t.Fatalf("want %d items, got %d", len(want), len(items))
	}

	for i, w := range want {
		got := items[i]
		if got.ID != w.id {
			t.Fatalf("position %d: want id %d, got %d", i, w.id, got.ID)
		}
		if got.CommandID != w.commandID {
			t.Fatalf("position %d: want command_id %d, got %d", i, w.commandID, got.CommandID)
		}
		if got.Script != w.script {
			t.Fatalf("position %d: want script %q, got %q", i, w.script, got.Script)
		}
		if got.Description != w.description {
			t.Fatalf("position %d: want description %q, got %q", i, w.description, got.Description)
		}
		if got.CreatedAt.Unix() != w.at.Unix() {
			t.Fatalf("position %d: want created_at %d, got %d", i, w.at.Unix(), got.CreatedAt.Unix())
		}
		if got.UpdatedAt.Unix() != w.at.Unix() {
			t.Fatalf("position %d: want updated_at %d, got %d", i, w.at.Unix(), got.UpdatedAt.Unix())
		}
	}
}

func TestGetCommandScriptsForNonExistentCommand(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	items, err := r.GetCommandScripts(999)
	if err != nil {
		t.Fatalf("GetCommandScripts: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("want zero items, got %d", len(items))
	}
}

func TestGetCommandScriptsIsRejectedInsideATransaction(t *testing.T) {
	s := newTestStore(t)

	err := s.WithTx(func(tx *store.Tx) error {
		r := New(s)
		_, err := r.GetCommandScripts(1)
		return err
	})

	if !errors.Is(err, store.ErrUseTx) {
		t.Fatalf("want store.ErrUseTx, got %v", err)
	}
}

func TestGetCommandScriptsPropagatesAClosedStore(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	_, err := r.GetCommandScripts(1)
	if !errors.Is(err, store.ErrClosed) {
		t.Fatalf("want store.ErrClosed, got %v", err)
	}
}

func TestGetCommandScriptFound(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "cmd"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	id, at := insertTestCommandScript(t, s, c.ID, "echotest", "echo test", "imprime test")

	got, err := r.GetCommandScript(c.ID, "echotest")
	if err != nil {
		t.Fatalf("GetCommandScript: %v", err)
	}

	if got.ID != id {
		t.Fatalf("want ID %d, got %d", id, got.ID)
	}
	if got.CommandID != c.ID {
		t.Fatalf("want CommandID %d, got %d", c.ID, got.CommandID)
	}
	if got.Script != "echo test" {
		t.Fatalf("want script %q, got %q", "echo test", got.Script)
	}
	if got.Description != "imprime test" {
		t.Fatalf("want description %q, got %q", "imprime test", got.Description)
	}
	if got.CreatedAt.Unix() != at.Unix() {
		t.Fatalf("want created_at %d, got %d", at.Unix(), got.CreatedAt.Unix())
	}
	if got.UpdatedAt.Unix() != at.Unix() {
		t.Fatalf("want updated_at %d, got %d", at.Unix(), got.UpdatedAt.Unix())
	}
}

func TestGetCommandScriptReadsANullDescriptionAsEmpty(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "cmd"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	insertTestCommandScript(t, s, c.ID, "legacy", "legacy", "")

	got, err := r.GetCommandScript(c.ID, "legacy")
	if err != nil {
		t.Fatalf("GetCommandScript: %v", err)
	}

	if got.Description != "" {
		t.Fatalf("want an empty description for NULL, got %q", got.Description)
	}
}

func TestGetCommandScriptNotFoundWrongItem(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "cmd"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	got, err := r.GetCommandScript(c.ID, "missing")
	if err != nil {
		t.Fatalf("GetCommandScript: %v", err)
	}
	if got.ID != 0 || got.CommandID != 0 || got.Script != "" {
		t.Fatalf("want empty, got %+v", got)
	}
}

func TestGetCommandScriptNotFoundWrongCommand(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "cmd"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	insertTestCommandScript(t, s, c.ID, "script", "script", "descricao do script")

	got, err := r.GetCommandScript(999, "script")
	if err != nil {
		t.Fatalf("GetCommandScript: %v", err)
	}
	if got.ID != 0 || got.CommandID != 0 || got.Script != "" {
		t.Fatalf("want empty, got %+v", got)
	}
}

func TestGetCommandScriptPropagatesAClosedStore(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "cmd"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	insertTestCommandScript(t, s, c.ID, "s", "s", "")

	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	_, err := r.GetCommandScript(c.ID, "s")
	if !errors.Is(err, store.ErrClosed) {
		t.Fatalf("want store.ErrClosed, got %v", err)
	}
}

func TestGetCommandScriptIsRejectedInsideATransaction(t *testing.T) {
	s := newTestStore(t)

	err := s.WithTx(func(tx *store.Tx) error {
		r := New(s)
		_, err := r.GetCommandScript(1, "s")
		return err
	})

	if !errors.Is(err, store.ErrUseTx) {
		t.Fatalf("want store.ErrUseTx, got %v", err)
	}
}

func TestAppendCommandScriptSuccess(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "cmd"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	item := &CommandScript{Name: "npmtest", Script: "npm test", Description: "roda a suite de testes"}
	item.CommandID = c.ID
	if err := r.AppendCommandScript(item); err != nil {
		t.Fatalf("AppendCommandScript: %v", err)
	}

	if item.ID == 0 {
		t.Fatal("want ID populated")
	}
	if item.CommandID != c.ID {
		t.Fatalf("want CommandID %d, got %d", c.ID, item.CommandID)
	}
	if item.Script != "npm test" {
		t.Fatalf("want script %q, got %q", "npm test", item.Script)
	}
	if item.Description != "roda a suite de testes" {
		t.Fatalf("want description %q, got %q", "roda a suite de testes", item.Description)
	}
	if item.CreatedAt.IsZero() || item.UpdatedAt.IsZero() {
		t.Fatal("want timestamps populated")
	}
	if !item.CreatedAt.Equal(item.UpdatedAt) {
		t.Fatalf("want CreatedAt == UpdatedAt on append")
	}

	items, err := r.GetCommandScripts(c.ID)
	if err != nil {
		t.Fatalf("GetCommandScripts: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("want 1 item, got %d", len(items))
	}
	if items[0].Description != "roda a suite de testes" {
		t.Fatalf("want description persisted as %q, got %q", "roda a suite de testes", items[0].Description)
	}
}

func TestAppendCommandScriptWithoutDescription(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "cmd"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	item := &CommandScript{Name: "list", Script: "ls"}
	item.CommandID = c.ID
	if err := r.AppendCommandScript(item); err != nil {
		t.Fatalf("AppendCommandScript: %v", err)
	}

	if item.Description != "" {
		t.Fatalf("want an empty description, got %q", item.Description)
	}

	got, err := r.GetCommandScript(c.ID, item.Name)
	if err != nil {
		t.Fatalf("GetCommandScript: %v", err)
	}
	if got.Description != "" {
		t.Fatalf("want an empty description persisted, got %q", got.Description)
	}
}

func TestAppendCommandScriptCommandIDRequired(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	item := &CommandScript{Name: "x", Script: "x"}
	err := r.AppendCommandScript(item)
	if !errors.Is(err, ErrCommandIDRequired) {
		t.Fatalf("want ErrCommandIDRequired, got %v", err)
	}
}

func TestAppendCommandScriptScriptRequired(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "cmd"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	item := &CommandScript{Name: "x", Script: ""}
	item.CommandID = c.ID
	err := r.AppendCommandScript(item)
	if !errors.Is(err, ErrCommandScriptScriptRequired) {
		t.Fatalf("want ErrCommandScriptScriptRequired, got %v", err)
	}
}

func TestAppendCommandScriptToNonExistentCommand(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	item := &CommandScript{Name: "x", Script: "x"}
	item.CommandID = 999
	err := r.AppendCommandScript(item)
	if err == nil {
		t.Fatal("want an error appending to a nonexistent command")
	}
}

func TestAppendCommandScriptScriptUniqueness(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "cmd"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	if err := r.AppendCommandScript(&CommandScript{CommandID: c.ID, Name: "dup1", Script: "dup"}); err != nil {
		t.Fatalf("AppendCommandScript: %v", err)
	}

	err := r.AppendCommandScript(&CommandScript{CommandID: c.ID, Name: "dup2", Script: "dup"})
	if err == nil {
		t.Fatal("want a UNIQUE error")
	}
}

func TestAppendCommandScriptPropagatesAClosedStore(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "cmd"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	err := r.AppendCommandScript(&CommandScript{CommandID: c.ID, Name: "x", Script: "x"})
	if !errors.Is(err, store.ErrClosed) {
		t.Fatalf("want store.ErrClosed, got %v", err)
	}
}

func TestAppendCommandScriptIsRejectedInsideATransaction(t *testing.T) {
	s := newTestStore(t)

	err := s.WithTx(func(tx *store.Tx) error {
		r := New(s)
		c := &Command{Name: "cmd"}
		if err := r.InsertCommand(c); err != nil {
			return err
		}
		return r.AppendCommandScript(&CommandScript{CommandID: c.ID, Name: "x", Script: "x"})
	})

	if !errors.Is(err, store.ErrTxActive) {
		t.Fatalf("want store.ErrTxActive, got %v", err)
	}
}

func TestUpdateCommandScriptSuccess(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "cmd"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	item := &CommandScript{Name: "old", Script: "old", Description: "descricao antiga"}
	item.CommandID = c.ID
	if err := r.AppendCommandScript(item); err != nil {
		t.Fatalf("AppendCommandScript: %v", err)
	}
	id := item.ID
	cmdID := item.CommandID
	created := item.CreatedAt
	updatedBefore := item.UpdatedAt

	time.Sleep(time.Microsecond)

	item.Name = "renamed"
	item.Script = "new"
	item.Description = "descricao nova"
	if err := r.UpdateCommandScript(item); err != nil {
		t.Fatalf("UpdateCommandScript: %v", err)
	}

	if item.ID != id {
		t.Fatalf("want ID preserved as %d, got %d", id, item.ID)
	}
	if item.CommandID != cmdID {
		t.Fatalf("want CommandID preserved as %d, got %d", cmdID, item.CommandID)
	}
	if item.Script != "new" {
		t.Fatalf("want script %q, got %q", "new", item.Script)
	}
	if item.Description != "descricao nova" {
		t.Fatalf("want description %q, got %q", "descricao nova", item.Description)
	}
	if !item.CreatedAt.Equal(created) {
		t.Fatalf("want CreatedAt preserved")
	}
	if item.UpdatedAt.Unix() < updatedBefore.Unix() {
		t.Fatalf("want UpdatedAt updated")
	}

	got, err := r.GetCommandScript(cmdID, item.Name)
	if err != nil {
		t.Fatalf("GetCommandScript: %v", err)
	}
	if got.Script != "new" {
		t.Fatalf("want script persisted as %q, got %q", "new", got.Script)
	}
	if got.Description != "descricao nova" {
		t.Fatalf("want description persisted as %q, got %q", "descricao nova", got.Description)
	}
}

func TestUpdateCommandScriptClearsTheDescription(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "cmd"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	item := &CommandScript{Name: "s", Script: "s", Description: "descricao que sera removida"}
	item.CommandID = c.ID
	if err := r.AppendCommandScript(item); err != nil {
		t.Fatalf("AppendCommandScript: %v", err)
	}

	item.Description = ""
	if err := r.UpdateCommandScript(item); err != nil {
		t.Fatalf("UpdateCommandScript: %v", err)
	}
	if item.Description != "" {
		t.Fatalf("want an empty description on the struct, got %q", item.Description)
	}

	got, err := r.GetCommandScript(c.ID, item.Name)
	if err != nil {
		t.Fatalf("GetCommandScript: %v", err)
	}
	if got.Description != "" {
		t.Fatalf("want an empty description persisted, got %q", got.Description)
	}
}

func TestUpdateCommandScriptCommandIDRequired(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	item := &CommandScript{ID: 1, Name: "x", Script: "x"}
	err := r.UpdateCommandScript(item)
	if !errors.Is(err, ErrCommandIDRequired) {
		t.Fatalf("want ErrCommandIDRequired, got %v", err)
	}
}

func TestUpdateCommandScriptIDRequired(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	item := &CommandScript{CommandID: 1, ID: 0, Name: "x", Script: "x"}
	err := r.UpdateCommandScript(item)
	if !errors.Is(err, ErrCommandScriptIDRequired) {
		t.Fatalf("want ErrCommandScriptIDRequired, got %v", err)
	}
}

func TestUpdateCommandScriptScriptRequired(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "cmd"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	item := &CommandScript{Name: "old", Script: "old"}
	item.CommandID = c.ID
	if err := r.AppendCommandScript(item); err != nil {
		t.Fatalf("AppendCommandScript: %v", err)
	}

	item.Script = ""
	err := r.UpdateCommandScript(item)
	if !errors.Is(err, ErrCommandScriptScriptRequired) {
		t.Fatalf("want ErrCommandScriptScriptRequired, got %v", err)
	}
}

func TestUpdateCommandScriptNotFoundWrongID(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "cmd"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	item := &CommandScript{CommandID: c.ID, ID: 999, Name: "x", Script: "x"}
	err := r.UpdateCommandScript(item)
	if !errors.Is(err, ErrCommandScriptNotFound) {
		t.Fatalf("want ErrCommandScriptNotFound, got %v", err)
	}
}

func TestUpdateCommandScriptNotFoundWrongCommandID(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "cmd"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	item := &CommandScript{Name: "s", Script: "s"}
	item.CommandID = c.ID
	if err := r.AppendCommandScript(item); err != nil {
		t.Fatalf("AppendCommandScript: %v", err)
	}

	item.Script = "new"
	item.CommandID = 999
	err := r.UpdateCommandScript(item)
	if !errors.Is(err, ErrCommandScriptNotFound) {
		t.Fatalf("want ErrCommandScriptNotFound, got %v", err)
	}
}

func TestUpdateCommandScriptScriptUniqueness(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "cmd"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	if err := r.AppendCommandScript(&CommandScript{CommandID: c.ID, Name: "a", Script: "a"}); err != nil {
		t.Fatalf("AppendCommandScript: %v", err)
	}
	item := &CommandScript{Name: "b", Script: "b"}
	item.CommandID = c.ID
	if err := r.AppendCommandScript(item); err != nil {
		t.Fatalf("AppendCommandScript: %v", err)
	}

	item.Script = "a"
	err := r.UpdateCommandScript(item)
	if err == nil {
		t.Fatal("want a UNIQUE error")
	}
}

func TestUpdateCommandScriptPropagatesAClosedStore(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "cmd"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	item := &CommandScript{Name: "s", Script: "s"}
	item.CommandID = c.ID
	if err := r.AppendCommandScript(item); err != nil {
		t.Fatalf("AppendCommandScript: %v", err)
	}

	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	item.Script = "x"
	err := r.UpdateCommandScript(item)
	if !errors.Is(err, store.ErrClosed) {
		t.Fatalf("want store.ErrClosed, got %v", err)
	}
}

func TestUpdateCommandScriptIsRejectedInsideATransaction(t *testing.T) {
	s := newTestStore(t)

	err := s.WithTx(func(tx *store.Tx) error {
		r := New(s)
		c := &Command{Name: "cmd"}
		if err := r.InsertCommand(c); err != nil {
			return err
		}
		item := &CommandScript{Name: "s", Script: "s"}
		item.CommandID = c.ID
		if err := r.AppendCommandScript(item); err != nil {
			return err
		}
		item.Script = "x"
		return r.UpdateCommandScript(item)
	})

	if !errors.Is(err, store.ErrTxActive) {
		t.Fatalf("want store.ErrTxActive, got %v", err)
	}
}

func TestRemoveCommandScriptSuccess(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "cmd"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	item := &CommandScript{Name: "rm", Script: "rm"}
	item.CommandID = c.ID
	if err := r.AppendCommandScript(item); err != nil {
		t.Fatalf("AppendCommandScript: %v", err)
	}

	if err := r.RemoveCommandScript(c.ID, item.ID); err != nil {
		t.Fatalf("RemoveCommandScript: %v", err)
	}

	got, err := r.GetCommandScript(c.ID, item.Name)
	if err != nil {
		t.Fatalf("GetCommandScript: %v", err)
	}
	if got.ID != 0 || got.CommandID != 0 || got.Script != "" {
		t.Fatalf("want the item removed, got %+v", got)
	}

	items, err := r.GetCommandScripts(c.ID)
	if err != nil {
		t.Fatalf("GetCommandScripts: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("want an empty list, got %d", len(items))
	}
}

func TestRemoveCommandScriptCommandIDRequired(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	err := r.RemoveCommandScript(0, 1)
	if !errors.Is(err, ErrCommandIDRequired) {
		t.Fatalf("want ErrCommandIDRequired, got %v", err)
	}
}

func TestRemoveCommandScriptIDRequired(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	err := r.RemoveCommandScript(1, 0)
	if !errors.Is(err, ErrCommandScriptIDRequired) {
		t.Fatalf("want ErrCommandScriptIDRequired, got %v", err)
	}
}

func TestRemoveCommandScriptNotFoundIsIdempotent(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	err := r.RemoveCommandScript(1, 999)
	if err != nil {
		t.Fatalf("want nil (idempotent), got %v", err)
	}
}

func TestRemoveCommandScriptWrongCommandID(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "cmd"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	item := &CommandScript{Name: "s", Script: "s"}
	item.CommandID = c.ID
	if err := r.AppendCommandScript(item); err != nil {
		t.Fatalf("AppendCommandScript: %v", err)
	}

	err := r.RemoveCommandScript(999, item.ID)
	if err != nil {
		t.Fatalf("want nil (idempotent), got %v", err)
	}

	got, err := r.GetCommandScript(c.ID, item.Name)
	if err != nil {
		t.Fatalf("GetCommandScript: %v", err)
	}
	if got.ID != item.ID {
		t.Fatalf("want the item preserved")
	}
}

func TestRemoveCommandScriptPropagatesAClosedStore(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "cmd"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	item := &CommandScript{Name: "s", Script: "s"}
	item.CommandID = c.ID
	if err := r.AppendCommandScript(item); err != nil {
		t.Fatalf("AppendCommandScript: %v", err)
	}

	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	err := r.RemoveCommandScript(c.ID, item.ID)
	if !errors.Is(err, store.ErrClosed) {
		t.Fatalf("want store.ErrClosed, got %v", err)
	}
}

func TestRemoveCommandScriptIsRejectedInsideATransaction(t *testing.T) {
	s := newTestStore(t)

	err := s.WithTx(func(tx *store.Tx) error {
		r := New(s)
		c := &Command{Name: "cmd"}
		if err := r.InsertCommand(c); err != nil {
			return err
		}
		item := &CommandScript{Name: "s", Script: "s"}
		item.CommandID = c.ID
		if err := r.AppendCommandScript(item); err != nil {
			return err
		}
		return r.RemoveCommandScript(c.ID, item.ID)
	})

	if !errors.Is(err, store.ErrTxActive) {
		t.Fatalf("want store.ErrTxActive, got %v", err)
	}
}

func TestGetCommandsSearch(t *testing.T) {
	r := New(newTestStore(t))
	for _, value := range []struct{ name, description string }{
		{"build", "Compile the project"},
		{"deploy", "Build and publish"},
		{"build-release", ""},
		{"100% done", "literal percent"},
		{"under_score", "literal underscore"},
		{"bang!", "literal escape"},
		{"quote's", "literal quote"},
	} {
		insertTestCommand(t, r.Store, value.name, value.description)
	}
	for _, tc := range []struct {
		search string
		names  []string
	}{
		{"BUILD", []string{"build", "deploy", "build-release"}},
		{"compile", []string{"build"}},
		{"missing", nil},
		{"%", []string{"build", "deploy", "build-release", "100% done", "under_score", "bang!", "quote's"}},
		{"_", []string{"build", "deploy", "build-release", "100% done", "under_score", "bang!", "quote's"}},
		{"!", []string{"bang!"}},
		{"quote's", []string{"quote's"}},
		{"' OR 1=1 --", nil},
		{"", []string{"build", "deploy", "build-release", "100% done", "under_score", "bang!", "quote's"}},
	} {
		t.Run(tc.search, func(t *testing.T) {
			commands, err := r.GetCommands(tc.search)
			if err != nil {
				t.Fatal(err)
			}
			if len(commands) != len(tc.names) {
				t.Fatalf("want %d rows, got %d", len(tc.names), len(commands))
			}
			for i, command := range commands {
				if command.Name != tc.names[i] {
					t.Fatalf("want %q, got %q", tc.names[i], command.Name)
				}
			}
		})
	}
}
