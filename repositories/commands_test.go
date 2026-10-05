package repositories

import (
	"errors"
	"testing"
	"time"

	"tartarus/store"
)

// insertTestCommand writes a command straight to the database, bypassing the
// repository, so tests can set up rows the repository API would not allow.
// It returns the id and the timestamp the row was stamped with, which the
// caller needs in order to compare what comes back out.
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

// insertTestCommandItem writes a command item straight to the database, on the
// same terms as insertTestCommand.
func insertTestCommandItem(t *testing.T, s *store.Store, commandID uint64, script string, description string) (uint64, time.Time) {
	t.Helper()

	now := time.Now().UTC()

	var id int64

	err := s.WithTx(func(tx *store.Tx) error {
		result, err := tx.Exec(`INSERT INTO command_items (command_id, script, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`, commandID, script, nullable(description), now, now)

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

// nullable maps an empty string to a nil argument, so a test can write SQL NULL
// into a nullable column. The repository always sends the plain string, so an
// empty description reaches the database as an empty string rather than NULL.
// A NULL is only reachable by writing the column directly, which is what these
// helpers are for.
func nullable(s string) any {
	if s == "" {
		return nil
	}

	return s
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
		if got.Name != w.name {
			t.Fatalf("posição %d: esperava nome %q, obteve %q", i, w.name, got.Name)
		}
		if got.Description != w.description {
			t.Fatalf("posição %d: esperava description %q, obteve %q", i, w.description, got.Description)
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

	id, at := insertTestCommand(t, s, "build", "compila o projeto")

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
	if got.Description != "compila o projeto" {
		t.Fatalf("esperava description %q, obteve %q", "compila o projeto", got.Description)
	}
	if got.CreatedAt.Unix() != at.Unix() {
		t.Fatalf("esperava created_at %d, obteve %d", at.Unix(), got.CreatedAt.Unix())
	}
	if got.UpdatedAt.Unix() != at.Unix() {
		t.Fatalf("esperava updated_at %d, obteve %d", at.Unix(), got.UpdatedAt.Unix())
	}
}

func TestGetCommandReadsANullDescriptionAsEmpty(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	// O helper escreve SQL NULL quando a description vem vazia, que e o estado
	// de qualquer linha criada antes de a coluna existir.
	id, _ := insertTestCommand(t, s, "legacy", "")

	got, err := r.GetCommand(id)
	if err != nil {
		t.Fatalf("GetCommand: %v", err)
	}

	if got.Description != "" {
		t.Fatalf("esperava description vazia para NULL, obteve %q", got.Description)
	}
}

func TestGetCommandNotFound(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	got, err := r.GetCommand(999)
	if err != nil {
		t.Fatalf("GetCommand: %v", err)
	}
	if got.ID != 0 || got.Name != "" || got.Description != "" {
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

	c := &Command{Name: "run", Description: "executa o projeto"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	if c.ID == 0 {
		t.Fatal("esperava ID populado")
	}
	if c.Name != "run" {
		t.Fatalf("esperava nome %q, obteve %q", "run", c.Name)
	}
	if c.Description != "executa o projeto" {
		t.Fatalf("esperava description %q, obteve %q", "executa o projeto", c.Description)
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
	if got.Description != "executa o projeto" {
		t.Fatalf("esperava description %q, obteve %q", "executa o projeto", got.Description)
	}
}

func TestInsertCommandWithoutDescription(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	// A description e nullable e nao tem validacao: um command sem descricao e
	// valido, e deve sair com o campo vazio em vez de erro.
	c := &Command{Name: "run"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	if c.Description != "" {
		t.Fatalf("esperava description vazia, obteve %q", c.Description)
	}

	got, err := r.GetCommand(c.ID)
	if err != nil {
		t.Fatalf("GetCommand: %v", err)
	}
	if got.Description != "" {
		t.Fatalf("esperava description vazia persistida, obteve %q", got.Description)
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
	if errors.Is(err, ErrCommandNameRequired) {
		t.Fatalf("esperava violacao de UNIQUE, obteve %v", err)
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

	// Pequeno delay para garantir mudança de timestamp
	time.Sleep(time.Microsecond)

	c.Name = "new"
	c.Description = "descricao nova"
	if err := r.UpdateCommand(c); err != nil {
		t.Fatalf("UpdateCommand: %v", err)
	}

	if c.ID != id {
		t.Fatalf("esperava ID preservado %d, obteve %d", id, c.ID)
	}
	if c.Name != "new" {
		t.Fatalf("esperava nome %q, obteve %q", "new", c.Name)
	}
	if c.Description != "descricao nova" {
		t.Fatalf("esperava description %q, obteve %q", "descricao nova", c.Description)
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
	if got.Description != "descricao nova" {
		t.Fatalf("esperava description persistida %q, obteve %q", "descricao nova", got.Description)
	}
}

func TestUpdateCommandClearsTheDescription(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "run", Description: "descricao que sera removida"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	// Description nao tem validacao, entao limpar e um update legitimo.
	c.Description = ""
	if err := r.UpdateCommand(c); err != nil {
		t.Fatalf("UpdateCommand: %v", err)
	}
	if c.Description != "" {
		t.Fatalf("esperava description vazia na struct, obteve %q", c.Description)
	}

	got, err := r.GetCommand(c.ID)
	if err != nil {
		t.Fatalf("GetCommand: %v", err)
	}
	if got.Description != "" {
		t.Fatalf("esperava description vazia persistida, obteve %q", got.Description)
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

	insertTestCommandItem(t, s, c.ID, "echo 1", "primeiro item")
	insertTestCommandItem(t, s, c.ID, "echo 2", "segundo item")

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
		id, at := insertTestCommandItem(t, s, c.ID, m.script, m.description)
		want = append(want, inserted{id: id, commandID: c.ID, script: m.script, description: m.description, at: at})
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
		if got.Description != w.description {
			t.Fatalf("posição %d: esperava description %q, obteve %q", i, w.description, got.Description)
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

	id, at := insertTestCommandItem(t, s, c.ID, "echo test", "imprime test")

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
	if got.Description != "imprime test" {
		t.Fatalf("esperava description %q, obteve %q", "imprime test", got.Description)
	}
	if got.CreatedAt.Unix() != at.Unix() {
		t.Fatalf("esperava created_at %d, obteve %d", at.Unix(), got.CreatedAt.Unix())
	}
	if got.UpdatedAt.Unix() != at.Unix() {
		t.Fatalf("esperava updated_at %d, obteve %d", at.Unix(), got.UpdatedAt.Unix())
	}
}

func TestGetCommandItemReadsANullDescriptionAsEmpty(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "cmd"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	id, _ := insertTestCommandItem(t, s, c.ID, "legacy", "")

	got, err := r.GetCommandItem(c.ID, id)
	if err != nil {
		t.Fatalf("GetCommandItem: %v", err)
	}

	if got.Description != "" {
		t.Fatalf("esperava description vazia para NULL, obteve %q", got.Description)
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

	id, _ := insertTestCommandItem(t, s, c.ID, "script", "descricao do script")

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

	id, _ := insertTestCommandItem(t, s, c.ID, "s", "")

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

	item := &CommandItem{Script: "npm test", Description: "roda a suite de testes"}
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
	if item.Description != "roda a suite de testes" {
		t.Fatalf("esperava description %q, obteve %q", "roda a suite de testes", item.Description)
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
	if items[0].Description != "roda a suite de testes" {
		t.Fatalf("esperava description persistida %q, obteve %q", "roda a suite de testes", items[0].Description)
	}
}

func TestAppendCommandItemWithoutDescription(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "cmd"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	item := &CommandItem{Script: "ls"}
	if err := r.AppendCommandItem(c.ID, item); err != nil {
		t.Fatalf("AppendCommandItem: %v", err)
	}

	if item.Description != "" {
		t.Fatalf("esperava description vazia, obteve %q", item.Description)
	}

	got, err := r.GetCommandItem(c.ID, item.ID)
	if err != nil {
		t.Fatalf("GetCommandItem: %v", err)
	}
	if got.Description != "" {
		t.Fatalf("esperava description vazia persistida, obteve %q", got.Description)
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

	item := &CommandItem{Script: "old", Description: "descricao antiga"}
	if err := r.AppendCommandItem(c.ID, item); err != nil {
		t.Fatalf("AppendCommandItem: %v", err)
	}
	id := item.ID
	cmdID := item.CommandID
	created := item.CreatedAt
	updatedBefore := item.UpdatedAt

	time.Sleep(time.Microsecond)

	item.Script = "new"
	item.Description = "descricao nova"
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
	if item.Description != "descricao nova" {
		t.Fatalf("esperava description %q, obteve %q", "descricao nova", item.Description)
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
	if got.Description != "descricao nova" {
		t.Fatalf("esperava description persistida %q, obteve %q", "descricao nova", got.Description)
	}
}

func TestUpdateCommandItemClearsTheDescription(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	c := &Command{Name: "cmd"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	item := &CommandItem{Script: "s", Description: "descricao que sera removida"}
	if err := r.AppendCommandItem(c.ID, item); err != nil {
		t.Fatalf("AppendCommandItem: %v", err)
	}

	item.Description = ""
	if err := r.UpdateCommandItem(c.ID, item); err != nil {
		t.Fatalf("UpdateCommandItem: %v", err)
	}
	if item.Description != "" {
		t.Fatalf("esperava description vazia na struct, obteve %q", item.Description)
	}

	got, err := r.GetCommandItem(c.ID, item.ID)
	if err != nil {
		t.Fatalf("GetCommandItem: %v", err)
	}
	if got.Description != "" {
		t.Fatalf("esperava description vazia persistida, obteve %q", got.Description)
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
