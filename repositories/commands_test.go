package repositories

// import (
// 	"errors"
// 	"testing"
// 	"time"

// 	"tartarus/store"
// )

// func insertCommand(t *testing.T, s *store.Store, name string) (uint64, time.Time) {
// 	t.Helper()
// 	now := time.Now().UTC()
// 	var id int64
// 	err := s.WithTx(func(tx *store.Tx) error {
// 		res, err := tx.Exec(`INSERT INTO commands (name, created_at, updated_at) VALUES (?, ?, ?)`, name, now, now)
// 		if err != nil {
// 			return err
// 		}
// 		id, err = res.LastInsertId()
// 		return err
// 	})
// 	if err != nil {
// 		t.Fatalf("inserting command %q: %v", name, err)
// 	}
// 	return uint64(id), now
// }

// func insertCommandItem(t *testing.T, s *store.Store, cmdID uint64, script string) (uint64, time.Time) {
// 	t.Helper()
// 	now := time.Now().UTC()
// 	var id int64
// 	err := s.WithTx(func(tx *store.Tx) error {
// 		res, err := tx.Exec(`INSERT INTO command_items (command_id, script, created_at, updated_at) VALUES (?, ?, ?, ?)`, cmdID, script, now, now)
// 		if err != nil {
// 			return err
// 		}
// 		id, err = res.LastInsertId()
// 		return err
// 	})
// 	if err != nil {
// 		t.Fatalf("inserting command item: %v", err)
// 	}
// 	return uint64(id), now
// }

// func TestGetCommandsOnEmptyDatabase(t *testing.T) {
// 	s := newTestStore(t)
// 	r := New(s)
// 	cmds, err := r.GetCommands()
// 	if err != nil {
// 		t.Fatalf("GetCommands: %v", err)
// 	}
// 	if len(cmds) != 0 {
// 		t.Fatalf("esperava 0, obteve %d", len(cmds))
// 	}
// }

// func TestGetCommandsReturnsEveryCommandInCreationOrder(t *testing.T) {
// 	s := newTestStore(t)
// 	type inserted struct {
// 		id   uint64
// 		name string
// 		at   time.Time
// 	}
// 	want := make([]inserted, 0, 3)
// 	for _, name := range []string{"zebra", "alpha", "mango"} {
// 		id, at := insertCommand(t, s, name)
// 		want = append(want, inserted{id: id, name: name, at: at})
// 	}
// 	r := New(s)
// 	cmds, err := r.GetCommands()
// 	if err != nil {
// 		t.Fatalf("GetCommands: %v", err)
// 	}
// 	if len(cmds) != len(want) {
// 		t.Fatalf("esperava %d, obteve %d", len(want), len(cmds))
// 	}
// 	for i, w := range want {
// 		got := cmds[i]
// 		if got.ID != w.id || got.Name != w.name || got.CreatedAt.Unix() != w.at.Unix() {
// 			t.Fatalf("posição %d incorreta: %+v vs %+v", i, got, w)
// 		}
// 	}
// }

// func TestGetCommandFound(t *testing.T) {
// 	s := newTestStore(t)
// 	r := New(s)
// 	id, at := insertCommand(t, s, "build")
// 	got, err := r.GetCommand(id)
// 	if err != nil {
// 		t.Fatalf("GetCommand: %v", err)
// 	}
// 	if got.ID != id || got.Name != "build" || got.CreatedAt.Unix() != at.Unix() {
// 		t.Fatalf("dados incorretos: %+v", got)
// 	}
// }

// func TestGetCommandNotFound(t *testing.T) {
// 	s := newTestStore(t)
// 	r := New(s)
// 	got, err := r.GetCommand(999)
// 	if err != nil {
// 		t.Fatalf("GetCommand: %v", err)
// 	}
// 	if got != (Command{}) {
// 		t.Fatalf("esperava vazio, obteve %+v", got)
// 	}
// }

// func TestGetCommandPropagatesAClosedStore(t *testing.T) {
// 	s := newTestStore(t)
// 	r := New(s)
// 	if err := s.Close(); err != nil {
// 		t.Fatalf("Close: %v", err)
// 	}
// 	_, err := r.GetCommand(1)
// 	if !errors.Is(err, store.ErrClosed) {
// 		t.Fatalf("esperava ErrClosed, obteve %v", err)
// 	}
// }

// func TestGetCommandIsRejectedInsideATransaction(t *testing.T) {
// 	s := newTestStore(t)
// 	r := New(s)
// 	err := s.WithTx(func(tx *store.Tx) error {
// 		_, err := r.GetCommand(1)
// 		return err
// 	})
// 	if !errors.Is(err, store.ErrUseTx) {
// 		t.Fatalf("esperava ErrUseTx, obteve %v", err)
// 	}
// }

// func TestInsertCommandSuccess(t *testing.T) {
// 	s := newTestStore(t)
// 	r := New(s)
// 	c := &Command{Name: "build"}
// 	if err := r.InsertCommand(c); err != nil {
// 		t.Fatalf("InsertCommand: %v", err)
// 	}
// 	if c.ID == 0 || c.CreatedAt.IsZero() {
// 		t.Fatalf("esperava campos populados")
// 	}
// 	got, err := r.GetCommand(c.ID)
// 	if err != nil {
// 		t.Fatalf("GetCommand: %v", err)
// 	}
// 	if got.Name != "build" {
// 		t.Fatalf("incorreto: %+v", got)
// 	}
// }

// func TestInsertCommandMissingName(t *testing.T) {
// 	s := newTestStore(t)
// 	r := New(s)
// 	err := r.InsertCommand(&Command{})
// 	if err == nil || !errors.Is(err, ErrCommandNameRequired) {
// 		t.Fatalf("esperava ErrCommandNameRequired, obteve %v", err)
// 	}
// }

// func TestInsertCommandPropagatesAClosedStore(t *testing.T) {
// 	s := newTestStore(t)
// 	r := New(s)
// 	if err := s.Close(); err != nil {
// 		t.Fatalf("Close: %v", err)
// 	}
// 	err := r.InsertCommand(&Command{Name: "a"})
// 	if err == nil {
// 		t.Fatalf("esperava erro")
// 	}
// }

// func TestInsertCommandIsRejectedInsideATransaction(t *testing.T) {
// 	s := newTestStore(t)
// 	r := New(s)
// 	err := s.WithTx(func(tx *store.Tx) error {
// 		return r.InsertCommand(&Command{Name: "a"})
// 	})
// 	if !errors.Is(err, store.ErrTxActive) {
// 		t.Fatalf("esperava ErrTxActive, obteve %v", err)
// 	}
// }

// func TestUpdateCommandSuccess(t *testing.T) {
// 	s := newTestStore(t)
// 	r := New(s)
// 	c := &Command{Name: "old"}
// 	_ = r.InsertCommand(c)
// 	created := c.CreatedAt
// 	c.Name = "new"
// 	if err := r.UpdateCommand(c); err != nil {
// 		t.Fatalf("UpdateCommand: %v", err)
// 	}
// 	if c.CreatedAt.Unix() != created.Unix() {
// 		t.Fatalf("CreatedAt não deve mudar")
// 	}
// 	got, _ := r.GetCommand(c.ID)
// 	if got.Name != "new" {
// 		t.Fatalf("não atualizou")
// 	}
// }

// func TestUpdateCommandIDRequired(t *testing.T) {
// 	s := newTestStore(t)
// 	r := New(s)
// 	err := r.UpdateCommand(&Command{Name: "a"})
// 	if err == nil || !errors.Is(err, ErrCommandIDRequired) {
// 		t.Fatalf("esperava ErrCommandIDRequired, obteve %v", err)
// 	}
// }

// func TestUpdateCommandMissingName(t *testing.T) {
// 	s := newTestStore(t)
// 	r := New(s)
// 	err := r.UpdateCommand(&Command{ID: 1})
// 	if err == nil || !errors.Is(err, ErrCommandNameRequired) {
// 		t.Fatalf("esperava ErrCommandNameRequired, obteve %v", err)
// 	}
// }

// func TestUpdateCommandNotFound(t *testing.T) {
// 	s := newTestStore(t)
// 	r := New(s)
// 	err := r.UpdateCommand(&Command{ID: 123, Name: "a"})
// 	if err == nil || !errors.Is(err, ErrCommandNotFound) {
// 		t.Fatalf("esperava ErrCommandNotFound, obteve %v", err)
// 	}
// }

// func TestUpdateCommandPropagatesAClosedStore(t *testing.T) {
// 	s := newTestStore(t)
// 	r := New(s)
// 	c := &Command{Name: "a"}
// 	_ = r.InsertCommand(c)
// 	if err := s.Close(); err != nil {
// 		t.Fatalf("Close: %v", err)
// 	}
// 	c.Name = "x"
// 	err := r.UpdateCommand(c)
// 	if err == nil {
// 		t.Fatalf("esperava erro")
// 	}
// }

// func TestUpdateCommandIsRejectedInsideATransaction(t *testing.T) {
// 	s := newTestStore(t)
// 	r := New(s)
// 	c := &Command{Name: "a"}
// 	_ = r.InsertCommand(c)
// 	c.Name = "x"
// 	err := s.WithTx(func(tx *store.Tx) error {
// 		return r.UpdateCommand(c)
// 	})
// 	if !errors.Is(err, store.ErrTxActive) {
// 		t.Fatalf("esperava ErrTxActive, obteve %v", err)
// 	}
// }

// func TestDeleteCommandSuccess(t *testing.T) {
// 	s := newTestStore(t)
// 	r := New(s)
// 	id, _ := insertCommand(t, s, "del")
// 	if err := r.DeleteCommand(id); err != nil {
// 		t.Fatalf("DeleteCommand: %v", err)
// 	}
// 	got, _ := r.GetCommand(id)
// 	if got.ID != 0 || got.Name != "" {
// 		t.Fatalf("deveria estar vazio")
// 	}
// }

// func TestDeleteCommandIDRequired(t *testing.T) {
// 	s := newTestStore(t)
// 	r := New(s)
// 	err := r.DeleteCommand(0)
// 	if err == nil || !errors.Is(err, ErrCommandIDRequired) {
// 		t.Fatalf("esperava ErrCommandIDRequired, obteve %v", err)
// 	}
// }

// func TestDeleteCommandNotFound(t *testing.T) {
// 	s := newTestStore(t)
// 	r := New(s)
// 	if err := r.DeleteCommand(999); err != nil {
// 		t.Fatalf("deveria retornar nil, obteve %v", err)
// 	}
// }

// func TestDeleteCommandPropagatesAClosedStore(t *testing.T) {
// 	s := newTestStore(t)
// 	r := New(s)
// 	id, _ := insertCommand(t, s, "del")
// 	if err := s.Close(); err != nil {
// 		t.Fatalf("Close: %v", err)
// 	}
// 	if err := r.DeleteCommand(id); err == nil {
// 		t.Fatalf("esperava erro")
// 	}
// }

// func TestDeleteCommandIsRejectedInsideATransaction(t *testing.T) {
// 	s := newTestStore(t)
// 	r := New(s)
// 	id, _ := insertCommand(t, s, "del")
// 	err := s.WithTx(func(tx *store.Tx) error {
// 		return r.DeleteCommand(id)
// 	})
// 	if !errors.Is(err, store.ErrTxActive) {
// 		t.Fatalf("esperava ErrTxActive, obteve %v", err)
// 	}
// }
