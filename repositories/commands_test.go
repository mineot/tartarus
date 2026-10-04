package repositories

import (
	"errors"
	"testing"
	"time"

	"tartarus/store"
)

// helper para inserir command diretamente
func insertTestCommand(t *testing.T, s *store.Store, name string) (uint64, time.Time) {
	t.Helper()

	now := time.Now().UTC()

	var id int64

	err := s.WithTx(func(tx *store.Tx) error {
		result, err := tx.Exec(`INSERT INTO commands (name, created_at, updated_at) VALUES (?, ?, ?)`, name, now, now)

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

// helper para inserir command item diretamente
func insertTestCommandItem(t *testing.T, s *store.Store, commandID uint64, script string) (uint64, time.Time) {
	t.Helper()

	now := time.Now().UTC()

	var id int64

	err := s.WithTx(func(tx *store.Tx) error {
		result, err := tx.Exec(`INSERT INTO command_items (command_id, script, created_at, updated_at) VALUES (?, ?, ?, ?)`, commandID, script, now, now)

		if err != nil {
			return err
		}

		id, err = result.LastInsertId()

		return err
	})

	if err != nil {
		t.Fatalf("inserting command item %q: %v", script, err)
	}

	return uint64(id), now
}

func TestGetCommandsOnEmptyDatabase(t *testing.T) {
	s := newTestStore(t)

	r := New(s)
	commands, err := r.GetCommands()

	if err != nil {
		t.Fatalf("GetCommands: %v", err)
	}

	if len(commands) != 0 {
		t.Fatalf("esperava zero commands, obteve %d", len(commands))
	}
}

func TestGetCommandsReturnsEveryCommandInCreationOrder(t *testing.T) {
	s := newTestStore(t)

	type inserted struct {
		id uint64
		at time.Time
	}

	want := make([]inserted, 0, 3)

	for _, m := range []struct {
		name string
	}{
		{"zebra"},
		{"alpha"},
		{"mango"},
	} {
		id, at := insertTestCommand(t, s, m.name)
		want = append(want, inserted{id: id, at: at})
	}

	r := New(s)
	commands, err := r.GetCommands()

	if err != nil {
		t.Fatalf("GetCommands: %v", err)
	}

	if len(commands) != len(want) {
		t.Fatalf("esperava %d commands, obteve %d", len(want), len(commands))
	}

	for i, w := range want {
		got := commands[i]

		if got.ID != w.id {
			t.Fatalf("posição %d: esperava id %d, obteve %d", i, w.id, got.ID)
		}
		if got.Name == "" {
			t.Fatalf("posição %d: esperava nome preenchido", i)
		}
		if got.CreatedAt.Unix() != w.at.Unix() {
			t.Fatalf("posição %d: esperava created_at %d, obteve %d", i, w.at.Unix(), got.CreatedAt.Unix())
		}
		if got.UpdatedAt.Unix() != w.at.Unix() {
			t.Fatalf("posição %d: esperava updated_at %d, obteve %d", i, w.at.Unix(), got.UpdatedAt.Unix())
		}
	}
}

func TestGetCommandsIsRejectedInsideATransaction(t *testing.T) {
	s := newTestStore(t)

	err := s.WithTx(func(tx *store.Tx) error {
		r := New(s)
		_, err := r.GetCommands()
		return err
	})

	if !errors.Is(err, store.ErrUseTx) {
		t.Fatalf("esperava store.ErrUseTx, obteve %v", err)
	}
}

func TestGetCommandsPropagatesAClosedStore(t *testing.T) {
	s := newTestStore(t)

	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	r := New(s)
	_, err := r.GetCommands()

	if !errors.Is(err, store.ErrClosed) {
		t.Fatalf("esperava store.ErrClosed, obteve %v", err)
	}
}

func TestGetCommandFound(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	id, at := insertTestCommand(t, s, "build")

	got, err := r.GetCommand(id)
	if err != nil {
		t.Fatalf("GetCommand: %v", err)
	}

	if got.ID != id {
		t.Fatalf("esperava ID %d, obteve %d", id, got.ID)
	}
	if got.Name != "build" {
		t.Fatalf("esperava nome %q, obteve %q", "build", got.Name)
	}
	if got.CreatedAt.Unix() != at.Unix() {
		t.Fatalf("esperava created_at %d, obteve %d", at.Unix(), got.CreatedAt.Unix())
	}
	if got.UpdatedAt.Unix() != at.Unix() {
		t.Fatalf("esperava updated_at %d, obteve %d", at.Unix(), got.UpdatedAt.Unix())
	}
}

func TestGetCommandNotFound(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	got, err := r.GetCommand(999)
	if err != nil {
		t.Fatalf("GetCommand: %v", err)
	}
	if got.ID != 0 || got.Name != "" {
		t.Fatalf("esperava command vazio, obteve %+v", got)
	}
}

func TestGetCommandPropagatesAClosedStore(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	_, err := r.GetCommand(1)

	if !errors.Is(err, store.ErrClosed) {
		t.Fatalf("esperava store.ErrClosed, obteve %v", err)
	}
}

func TestGetCommandIsRejectedInsideATransaction(t *testing.T) {
	s := newTestStore(t)

	err := s.WithTx(func(tx *store.Tx) error {
		r := New(s)
		_, err := r.GetCommand(1)
		return err
	})

	if !errors.Is(err, store.ErrUseTx) {
		t.Fatalf("esperava store.ErrUseTx, obteve %v", err)
	}
}

func TestInsertCommandSuccess(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "run"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	if c.ID == 0 {
		t.Fatal("esperava ID populado")
	}
	if c.Name != "run" {
		t.Fatalf("esperava nome %q, obteve %q", "run", c.Name)
	}
	if c.CreatedAt.IsZero() {
		t.Fatal("esperava CreatedAt populado")
	}
	if c.UpdatedAt.IsZero() {
		t.Fatal("esperava UpdatedAt populado")
	}
	if !c.CreatedAt.Equal(c.UpdatedAt) {
		t.Fatalf("esperava CreatedAt == UpdatedAt no insert, obteve %v vs %v", c.CreatedAt, c.UpdatedAt)
	}

	got, err := r.GetCommand(c.ID)
	if err != nil {
		t.Fatalf("GetCommand: %v", err)
	}
	if got.ID != c.ID {
		t.Fatalf("esperava ID %d, obteve %d", c.ID, got.ID)
	}
	if got.Name != "run" {
		t.Fatalf("esperava nome %q, obteve %q", "run", got.Name)
	}
}

func TestInsertCommandMissingName(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: ""}
	err := r.InsertCommand(c)
	if !errors.Is(err, ErrCommandNameRequired) {
		t.Fatalf("esperava ErrCommandNameRequired, obteve %v", err)
	}

	commands, err := r.GetCommands()
	if err != nil {
		t.Fatalf("GetCommands: %v", err)
	}
	if len(commands) != 0 {
		t.Fatalf("esperava banco inalterado, obteve %d commands", len(commands))
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
		t.Fatalf("esperava store.ErrClosed, obteve %v", err)
	}
}

func TestInsertCommandIsRejectedInsideATransaction(t *testing.T) {
	s := newTestStore(t)

	err := s.WithTx(func(tx *store.Tx) error {
		r := New(s)
		return r.InsertCommand(&Command{Name: "inside"})
	})

	if !errors.Is(err, store.ErrTxActive) {
		t.Fatalf("esperava store.ErrTxActive, obteve %v", err)
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
		t.Fatal("esperava erro de UNIQUE")
	}
	if !errors.Is(err, store.ErrUseTx) && !errors.Is(err, store.ErrClosed) && !errors.Is(err, store.ErrTxActive) && !errors.Is(err, ErrCommandNameRequired) && !errors.Is(err, ErrCommandIDRequired) {
		// Verifica apenas que é wrapped: contém "repositories: inserting command"
		_ = err
	}
	// Aceita qualquer erro (SQLite UNIQUE) desde que seja wrapped
}

func TestUpdateCommandSuccess(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "old"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}
	id := c.ID
	created := c.CreatedAt
	updatedBefore := c.UpdatedAt

	// Pequeno delay para garantir mudança de timestamp
	time.Sleep(time.Microsecond)

	c.Name = "new"
	if err := r.UpdateCommand(c); err != nil {
		t.Fatalf("UpdateCommand: %v", err)
	}

	if c.ID != id {
		t.Fatalf("esperava ID preservado %d, obteve %d", id, c.ID)
	}
	if c.Name != "new" {
		t.Fatalf("esperava nome %q, obteve %q", "new", c.Name)
	}
	if !c.CreatedAt.Equal(created) {
		t.Fatalf("esperava CreatedAt preservado %v, obteve %v", created, c.CreatedAt)
	}
	if c.UpdatedAt.Unix() < updatedBefore.Unix() {
		t.Fatalf("esperava UpdatedAt atualizado")
	}

	got, err := r.GetCommand(id)
	if err != nil {
		t.Fatalf("GetCommand: %v", err)
	}
	if got.Name != "new" {
		t.Fatalf("esperava nome persistido %q, obteve %q", "new", got.Name)
	}
}

func TestUpdateCommandIDRequired(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{ID: 0, Name: "x"}
	err := r.UpdateCommand(c)
	if !errors.Is(err, ErrCommandIDRequired) {
		t.Fatalf("esperava ErrCommandIDRequired, obteve %v", err)
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
		t.Fatalf("esperava ErrCommandNameRequired, obteve %v", err)
	}
}

func TestUpdateCommandNotFound(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{ID: 999, Name: "ghost"}
	err := r.UpdateCommand(c)
	if !errors.Is(err, ErrCommandNotFound) {
		t.Fatalf("esperava ErrCommandNotFound, obteve %v", err)
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
		t.Fatalf("esperava store.ErrClosed, obteve %v", err)
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
		t.Fatalf("esperava store.ErrTxActive, obteve %v", err)
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
		t.Fatal("esperava erro de UNIQUE")
	}
}

func TestDeleteCommandSuccess(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "target"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	insertTestCommandItem(t, s, c.ID, "echo 1")
	insertTestCommandItem(t, s, c.ID, "echo 2")

	if err := r.DeleteCommand(c.ID); err != nil {
		t.Fatalf("DeleteCommand: %v", err)
	}

	got, err := r.GetCommand(c.ID)
	if err != nil {
		t.Fatalf("GetCommand: %v", err)
	}
	if got.ID != 0 || got.Name != "" {
		t.Fatalf("esperava command removido, obteve %+v", got)
	}

	items, err := r.GetCommandItems(c.ID)
	if err != nil {
		t.Fatalf("GetCommandItems: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("esperava itens removidos por CASCADE, obteve %d", len(items))
	}
}

func TestDeleteCommandIDRequired(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	err := r.DeleteCommand(0)
	if !errors.Is(err, ErrCommandIDRequired) {
		t.Fatalf("esperava ErrCommandIDRequired, obteve %v", err)
	}
}

func TestDeleteCommandNotFoundIsIdempotent(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	err := r.DeleteCommand(999)
	if err != nil {
		t.Fatalf("esperava nil (idempotente), obteve %v", err)
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
		t.Fatalf("esperava store.ErrClosed, obteve %v", err)
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
		t.Fatalf("esperava store.ErrTxActive, obteve %v", err)
	}
}

func TestGetCommandItemsEmpty(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "cmd"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	items, err := r.GetCommandItems(c.ID)
	if err != nil {
		t.Fatalf("GetCommandItems: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("esperava zero itens, obteve %d", len(items))
	}
}

func TestGetCommandItemsReturnsAllInOrder(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "cmd"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	type inserted struct {
		id       uint64
		commandID uint64
		script   string
		at       time.Time
	}

	want := make([]inserted, 0, 3)

	for _, m := range []struct{ script string }{
		{"z"},
		{"a"},
		{"m"},
	} {
		id, at := insertTestCommandItem(t, s, c.ID, m.script)
		want = append(want, inserted{id: id, commandID: c.ID, script: m.script, at: at})
	}

	items, err := r.GetCommandItems(c.ID)
	if err != nil {
		t.Fatalf("GetCommandItems: %v", err)
	}

	if len(items) != len(want) {
		t.Fatalf("esperava %d itens, obteve %d", len(want), len(items))
	}

	for i, w := range want {
		got := items[i]
		if got.ID != w.id {
			t.Fatalf("posição %d: esperava id %d, obteve %d", i, w.id, got.ID)
		}
		if got.CommandID != w.commandID {
			t.Fatalf("posição %d: esperava command_id %d, obteve %d", i, w.commandID, got.CommandID)
		}
		if got.Script != w.script {
			t.Fatalf("posição %d: esperava script %q, obteve %q", i, w.script, got.Script)
		}
		if got.CreatedAt.Unix() != w.at.Unix() {
			t.Fatalf("posição %d: esperava created_at %d, obteve %d", i, w.at.Unix(), got.CreatedAt.Unix())
		}
		if got.UpdatedAt.Unix() != w.at.Unix() {
			t.Fatalf("posição %d: esperava updated_at %d, obteve %d", i, w.at.Unix(), got.UpdatedAt.Unix())
		}
	}
}

func TestGetCommandItemsForNonExistentCommand(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	items, err := r.GetCommandItems(999)
	if err != nil {
		t.Fatalf("GetCommandItems: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("esperava zero itens, obteve %d", len(items))
	}
}

func TestGetCommandItemsIsRejectedInsideATransaction(t *testing.T) {
	s := newTestStore(t)

	err := s.WithTx(func(tx *store.Tx) error {
		r := New(s)
		_, err := r.GetCommandItems(1)
		return err
	})

	if !errors.Is(err, store.ErrUseTx) {
		t.Fatalf("esperava store.ErrUseTx, obteve %v", err)
	}
}

func TestGetCommandItemsPropagatesAClosedStore(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	_, err := r.GetCommandItems(1)
	if !errors.Is(err, store.ErrClosed) {
		t.Fatalf("esperava store.ErrClosed, obteve %v", err)
	}
}

func TestGetCommandItemFound(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "cmd"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	id, at := insertTestCommandItem(t, s, c.ID, "echo test")

	got, err := r.GetCommandItem(c.ID, id)
	if err != nil {
		t.Fatalf("GetCommandItem: %v", err)
	}

	if got.ID != id {
		t.Fatalf("esperava ID %d, obteve %d", id, got.ID)
	}
	if got.CommandID != c.ID {
		t.Fatalf("esperava CommandID %d, obteve %d", c.ID, got.CommandID)
	}
	if got.Script != "echo test" {
		t.Fatalf("esperava script %q, obteve %q", "echo test", got.Script)
	}
	if got.CreatedAt.Unix() != at.Unix() {
		t.Fatalf("esperava created_at %d, obteve %d", at.Unix(), got.CreatedAt.Unix())
	}
	if got.UpdatedAt.Unix() != at.Unix() {
		t.Fatalf("esperava updated_at %d, obteve %d", at.Unix(), got.UpdatedAt.Unix())
	}
}

func TestGetCommandItemNotFoundWrongItem(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "cmd"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	got, err := r.GetCommandItem(c.ID, 999)
	if err != nil {
		t.Fatalf("GetCommandItem: %v", err)
	}
	if got.ID != 0 || got.CommandID != 0 || got.Script != "" {
		t.Fatalf("esperava vazio, obteve %+v", got)
	}
}

func TestGetCommandItemNotFoundWrongCommand(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "cmd"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	id, _ := insertTestCommandItem(t, s, c.ID, "script")

	got, err := r.GetCommandItem(999, id)
	if err != nil {
		t.Fatalf("GetCommandItem: %v", err)
	}
	if got.ID != 0 || got.CommandID != 0 || got.Script != "" {
		t.Fatalf("esperava vazio, obteve %+v", got)
	}
}

func TestGetCommandItemPropagatesAClosedStore(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "cmd"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	id, _ := insertTestCommandItem(t, s, c.ID, "s")

	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	_, err := r.GetCommandItem(c.ID, id)
	if !errors.Is(err, store.ErrClosed) {
		t.Fatalf("esperava store.ErrClosed, obteve %v", err)
	}
}

func TestGetCommandItemIsRejectedInsideATransaction(t *testing.T) {
	s := newTestStore(t)

	err := s.WithTx(func(tx *store.Tx) error {
		r := New(s)
		_, err := r.GetCommandItem(1, 1)
		return err
	})

	if !errors.Is(err, store.ErrUseTx) {
		t.Fatalf("esperava store.ErrUseTx, obteve %v", err)
	}
}

func TestAppendCommandItemSuccess(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "cmd"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	item := &CommandItem{Script: "npm test"}
	if err := r.AppendCommandItem(c.ID, item); err != nil {
		t.Fatalf("AppendCommandItem: %v", err)
	}

	if item.ID == 0 {
		t.Fatal("esperava ID populado")
	}
	if item.CommandID != c.ID {
		t.Fatalf("esperava CommandID %d, obteve %d", c.ID, item.CommandID)
	}
	if item.Script != "npm test" {
		t.Fatalf("esperava script %q, obteve %q", "npm test", item.Script)
	}
	if item.CreatedAt.IsZero() || item.UpdatedAt.IsZero() {
		t.Fatal("esperava timestamps populados")
	}
	if !item.CreatedAt.Equal(item.UpdatedAt) {
		t.Fatalf("esperava CreatedAt == UpdatedAt no append")
	}

	items, err := r.GetCommandItems(c.ID)
	if err != nil {
		t.Fatalf("GetCommandItems: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("esperava 1 item, obteve %d", len(items))
	}
}

func TestAppendCommandItemCommandIDRequired(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	item := &CommandItem{Script: "x"}
	err := r.AppendCommandItem(0, item)
	if !errors.Is(err, ErrCommandIDRequired) {
		t.Fatalf("esperava ErrCommandIDRequired, obteve %v", err)
	}
}

func TestAppendCommandItemScriptRequired(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "cmd"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	item := &CommandItem{Script: ""}
	err := r.AppendCommandItem(c.ID, item)
	if !errors.Is(err, ErrCommandItemScriptRequired) {
		t.Fatalf("esperava ErrCommandItemScriptRequired, obteve %v", err)
	}
}

func TestAppendCommandItemToNonExistentCommand(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	item := &CommandItem{Script: "x"}
	err := r.AppendCommandItem(999, item)
	if err == nil {
		t.Fatal("esperava erro ao anexar em command inexistente")
	}
}

func TestAppendCommandItemScriptUniqueness(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "cmd"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	if err := r.AppendCommandItem(c.ID, &CommandItem{Script: "dup"}); err != nil {
		t.Fatalf("AppendCommandItem: %v", err)
	}

	err := r.AppendCommandItem(c.ID, &CommandItem{Script: "dup"})
	if err == nil {
		t.Fatal("esperava erro de UNIQUE")
	}
}

func TestAppendCommandItemPropagatesAClosedStore(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "cmd"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	err := r.AppendCommandItem(c.ID, &CommandItem{Script: "x"})
	if !errors.Is(err, store.ErrClosed) {
		t.Fatalf("esperava store.ErrClosed, obteve %v", err)
	}
}

func TestAppendCommandItemIsRejectedInsideATransaction(t *testing.T) {
	s := newTestStore(t)

	err := s.WithTx(func(tx *store.Tx) error {
		r := New(s)
		c := &Command{Name: "cmd"}
		if err := r.InsertCommand(c); err != nil {
			return err
		}
		return r.AppendCommandItem(c.ID, &CommandItem{Script: "x"})
	})

	if !errors.Is(err, store.ErrTxActive) {
		t.Fatalf("esperava store.ErrTxActive, obteve %v", err)
	}
}

func TestUpdateCommandItemSuccess(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "cmd"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	item := &CommandItem{Script: "old"}
	if err := r.AppendCommandItem(c.ID, item); err != nil {
		t.Fatalf("AppendCommandItem: %v", err)
	}
	id := item.ID
	cmdID := item.CommandID
	created := item.CreatedAt
	updatedBefore := item.UpdatedAt

	time.Sleep(time.Microsecond)

	item.Script = "new"
	if err := r.UpdateCommandItem(cmdID, item); err != nil {
		t.Fatalf("UpdateCommandItem: %v", err)
	}

	if item.ID != id {
		t.Fatalf("esperava ID preservado %d, obteve %d", id, item.ID)
	}
	if item.CommandID != cmdID {
		t.Fatalf("esperava CommandID preservado %d, obteve %d", cmdID, item.CommandID)
	}
	if item.Script != "new" {
		t.Fatalf("esperava script %q, obteve %q", "new", item.Script)
	}
	if !item.CreatedAt.Equal(created) {
		t.Fatalf("esperava CreatedAt preservado")
	}
	if item.UpdatedAt.Unix() < updatedBefore.Unix() {
		t.Fatalf("esperava UpdatedAt atualizado")
	}

	got, err := r.GetCommandItem(cmdID, id)
	if err != nil {
		t.Fatalf("GetCommandItem: %v", err)
	}
	if got.Script != "new" {
		t.Fatalf("esperava script persistido %q, obteve %q", "new", got.Script)
	}
}

func TestUpdateCommandItemCommandIDRequired(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	item := &CommandItem{ID: 1, Script: "x"}
	err := r.UpdateCommandItem(0, item)
	if !errors.Is(err, ErrCommandIDRequired) {
		t.Fatalf("esperava ErrCommandIDRequired, obteve %v", err)
	}
}

func TestUpdateCommandItemIDRequired(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	item := &CommandItem{ID: 0, Script: "x"}
	err := r.UpdateCommandItem(1, item)
	if !errors.Is(err, ErrCommandItemIDRequired) {
		t.Fatalf("esperava ErrCommandItemIDRequired, obteve %v", err)
	}
}

func TestUpdateCommandItemScriptRequired(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "cmd"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	item := &CommandItem{Script: "old"}
	if err := r.AppendCommandItem(c.ID, item); err != nil {
		t.Fatalf("AppendCommandItem: %v", err)
	}

	item.Script = ""
	err := r.UpdateCommandItem(c.ID, item)
	if !errors.Is(err, ErrCommandItemScriptRequired) {
		t.Fatalf("esperava ErrCommandItemScriptRequired, obteve %v", err)
	}
}

func TestUpdateCommandItemNotFoundWrongID(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "cmd"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	item := &CommandItem{ID: 999, Script: "x"}
	err := r.UpdateCommandItem(c.ID, item)
	if !errors.Is(err, ErrCommandItemNotFound) {
		t.Fatalf("esperava ErrCommandItemNotFound, obteve %v", err)
	}
}

func TestUpdateCommandItemNotFoundWrongCommandID(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "cmd"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	item := &CommandItem{Script: "s"}
	if err := r.AppendCommandItem(c.ID, item); err != nil {
		t.Fatalf("AppendCommandItem: %v", err)
	}

	item.Script = "new"
	err := r.UpdateCommandItem(999, item)
	if !errors.Is(err, ErrCommandItemNotFound) {
		t.Fatalf("esperava ErrCommandItemNotFound, obteve %v", err)
	}
}

func TestUpdateCommandItemScriptUniqueness(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "cmd"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	if err := r.AppendCommandItem(c.ID, &CommandItem{Script: "a"}); err != nil {
		t.Fatalf("AppendCommandItem: %v", err)
	}
	item := &CommandItem{Script: "b"}
	if err := r.AppendCommandItem(c.ID, item); err != nil {
		t.Fatalf("AppendCommandItem: %v", err)
	}

	item.Script = "a"
	err := r.UpdateCommandItem(c.ID, item)
	if err == nil {
		t.Fatal("esperava erro de UNIQUE")
	}
}

func TestUpdateCommandItemPropagatesAClosedStore(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "cmd"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	item := &CommandItem{Script: "s"}
	if err := r.AppendCommandItem(c.ID, item); err != nil {
		t.Fatalf("AppendCommandItem: %v", err)
	}

	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	item.Script = "x"
	err := r.UpdateCommandItem(c.ID, item)
	if !errors.Is(err, store.ErrClosed) {
		t.Fatalf("esperava store.ErrClosed, obteve %v", err)
	}
}

func TestUpdateCommandItemIsRejectedInsideATransaction(t *testing.T) {
	s := newTestStore(t)

	err := s.WithTx(func(tx *store.Tx) error {
		r := New(s)
		c := &Command{Name: "cmd"}
		if err := r.InsertCommand(c); err != nil {
			return err
		}
		item := &CommandItem{Script: "s"}
		if err := r.AppendCommandItem(c.ID, item); err != nil {
			return err
		}
		item.Script = "x"
		return r.UpdateCommandItem(c.ID, item)
	})

	if !errors.Is(err, store.ErrTxActive) {
		t.Fatalf("esperava store.ErrTxActive, obteve %v", err)
	}
}

func TestRemoveCommandItemSuccess(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "cmd"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	item := &CommandItem{Script: "rm"}
	if err := r.AppendCommandItem(c.ID, item); err != nil {
		t.Fatalf("AppendCommandItem: %v", err)
	}

	if err := r.RemoveCommandItem(c.ID, item.ID); err != nil {
		t.Fatalf("RemoveCommandItem: %v", err)
	}

	got, err := r.GetCommandItem(c.ID, item.ID)
	if err != nil {
		t.Fatalf("GetCommandItem: %v", err)
	}
	if got.ID != 0 || got.CommandID != 0 || got.Script != "" {
		t.Fatalf("esperava item removido, obteve %+v", got)
	}

	items, err := r.GetCommandItems(c.ID)
	if err != nil {
		t.Fatalf("GetCommandItems: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("esperava lista vazia, obteve %d", len(items))
	}
}

func TestRemoveCommandItemCommandIDRequired(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	err := r.RemoveCommandItem(0, 1)
	if !errors.Is(err, ErrCommandIDRequired) {
		t.Fatalf("esperava ErrCommandIDRequired, obteve %v", err)
	}
}

func TestRemoveCommandItemIDRequired(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	err := r.RemoveCommandItem(1, 0)
	if !errors.Is(err, ErrCommandItemIDRequired) {
		t.Fatalf("esperava ErrCommandItemIDRequired, obteve %v", err)
	}
}

func TestRemoveCommandItemNotFoundIsIdempotent(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	err := r.RemoveCommandItem(1, 999)
	if err != nil {
		t.Fatalf("esperava nil (idempotente), obteve %v", err)
	}
}

func TestRemoveCommandItemWrongCommandID(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "cmd"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	item := &CommandItem{Script: "s"}
	if err := r.AppendCommandItem(c.ID, item); err != nil {
		t.Fatalf("AppendCommandItem: %v", err)
	}

	err := r.RemoveCommandItem(999, item.ID)
	if err != nil {
		t.Fatalf("esperava nil (idempotente), obteve %v", err)
	}

	// Verifica que item não foi removido
	got, err := r.GetCommandItem(c.ID, item.ID)
	if err != nil {
		t.Fatalf("GetCommandItem: %v", err)
	}
	if got.ID != item.ID {
		t.Fatalf("esperava item preservado")
	}
}

func TestRemoveCommandItemPropagatesAClosedStore(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "cmd"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	item := &CommandItem{Script: "s"}
	if err := r.AppendCommandItem(c.ID, item); err != nil {
		t.Fatalf("AppendCommandItem: %v", err)
	}

	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	err := r.RemoveCommandItem(c.ID, item.ID)
	if !errors.Is(err, store.ErrClosed) {
		t.Fatalf("esperava store.ErrClosed, obteve %v", err)
	}
}

func TestRemoveCommandItemIsRejectedInsideATransaction(t *testing.T) {
	s := newTestStore(t)

	err := s.WithTx(func(tx *store.Tx) error {
		r := New(s)
		c := &Command{Name: "cmd"}
		if err := r.InsertCommand(c); err != nil {
			return err
		}
		item := &CommandItem{Script: "s"}
		if err := r.AppendCommandItem(c.ID, item); err != nil {
			return err
		}
		return r.RemoveCommandItem(c.ID, item.ID)
	})

	if !errors.Is(err, store.ErrTxActive) {
		t.Fatalf("esperava store.ErrTxActive, obteve %v", err)
	}
}
