package repositories

import (
	"errors"
	"testing"
	"time"

	"tartarus/store"
)

func insertManual(t *testing.T, s *store.Store, name string, body string, description string) (uint64, time.Time) {
	t.Helper()

	now := time.Now().UTC()

	var id int64

	err := s.WithTx(func(tx *store.Tx) error {
		result, err := tx.Exec(`INSERT INTO manuals (name, body, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`, name, body, nullable(description), now, now)

		if err != nil {
			return err
		}

		id, err = result.LastInsertId()

		return err
	})

	if err != nil {
		t.Fatalf("inserting manual %q: %v", name, err)
	}

	return uint64(id), now
}

func TestGetManualsOnEmptyDatabase(t *testing.T) {
	s := newTestStore(t)

	r := New(s)
	manuals, err := r.GetManuals()

	if err != nil {
		t.Fatalf("GetManuals: %v", err)
	}

	if len(manuals) != 0 {
		t.Fatalf("want zero manuals, got %d", len(manuals))
	}
}

func TestGetManualsReturnsEveryManualInCreationOrder(t *testing.T) {
	s := newTestStore(t)

	type inserted struct {
		id          uint64
		name        string
		body        string
		description string
		at          time.Time
	}

	want := make([]inserted, 0, 3)

	for _, m := range []struct{ name, body, description string }{
		{"zebra", "last alphabetically", "ultimo alfabeticamente"},
		{"alpha", "first alphabetically", ""},
		{"mango", "in between", "no meio"},
	} {
		id, at := insertManual(t, s, m.name, m.body, m.description)
		want = append(want, inserted{id: id, name: m.name, body: m.body, description: m.description, at: at})
	}

	r := New(s)
	manuals, err := r.GetManuals()

	if err != nil {
		t.Fatalf("GetManuals: %v", err)
	}

	if len(manuals) != len(want) {
		t.Fatalf("want %d manuals, got %d", len(want), len(manuals))
	}

	for i, w := range want {
		got := manuals[i]

		if got.ID != w.id {
			t.Fatalf("position %d: want id %d, got %d", i, w.id, got.ID)
		}
		if got.Name != w.name {
			t.Fatalf("position %d: want name %q, got %q", i, w.name, got.Name)
		}
		if got.Body != w.body {
			t.Fatalf("position %d: want body %q, got %q", i, w.body, got.Body)
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

func TestGetManualsIsRejectedInsideATransaction(t *testing.T) {
	s := newTestStore(t)

	err := s.WithTx(func(tx *store.Tx) error {
		r := New(s)
		_, err := r.GetManuals()
		return err
	})

	if !errors.Is(err, store.ErrUseTx) {
		t.Fatalf("want store.ErrUseTx, got %v", err)
	}
}

func TestGetManualsPropagatesAClosedStore(t *testing.T) {
	s := newTestStore(t)

	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	r := New(s)
	_, err := r.GetManuals()

	if !errors.Is(err, store.ErrClosed) {
		t.Fatalf("want store.ErrClosed, got %v", err)
	}
}

func TestGetManualFound(t *testing.T) {
	s := newTestStore(t)
	r := New(s)
	id, at := insertManual(t, s, "readme", "conteudo", "como usar o comando")

	got, err := r.GetManual(id)
	if err != nil {
		t.Fatalf("GetManual: %v", err)
	}

	if got.ID != id {
		t.Fatalf("want ID %d, got %d", id, got.ID)
	}
	if got.Name != "readme" {
		t.Fatalf("want name %q, got %q", "readme", got.Name)
	}
	if got.Body != "conteudo" {
		t.Fatalf("want body %q, got %q", "conteudo", got.Body)
	}
	if got.Description != "como usar o comando" {
		t.Fatalf("want description %q, got %q", "como usar o comando", got.Description)
	}
	if got.CreatedAt.Unix() != at.Unix() {
		t.Fatalf("want created_at %d, got %d", at.Unix(), got.CreatedAt.Unix())
	}
	if got.UpdatedAt.Unix() != at.Unix() {
		t.Fatalf("want updated_at %d, got %d", at.Unix(), got.UpdatedAt.Unix())
	}
}

func TestGetManualReadsANullDescriptionAsEmpty(t *testing.T) {
	s := newTestStore(t)
	r := New(s)
	id, _ := insertManual(t, s, "legacy", "conteudo", "")

	got, err := r.GetManual(id)
	if err != nil {
		t.Fatalf("GetManual: %v", err)
	}

	if got.Description != "" {
		t.Fatalf("want an empty description for NULL, got %q", got.Description)
	}
}

func TestGetManualNotFound(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	got, err := r.GetManual(999)
	if err != nil {
		t.Fatalf("GetManual: %v", err)
	}
	if got != (Manual{}) {
		t.Fatalf("want an empty manual, got %+v", got)
	}
}

func TestGetManualPropagatesAClosedStore(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	_, err := r.GetManual(1)
	if !errors.Is(err, store.ErrClosed) {
		t.Fatalf("want store.ErrClosed, got %v", err)
	}
}

func TestGetManualIsRejectedInsideATransaction(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	err := s.WithTx(func(tx *store.Tx) error {
		_, err := r.GetManual(1)
		return err
	})
	if !errors.Is(err, store.ErrUseTx) {
		t.Fatalf("want store.ErrUseTx, got %v", err)
	}
}

func TestInsertManualSuccess(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	m := &Manual{Name: "guia", Body: "passos", Description: "guia de instalacao"}
	if err := r.InsertManual(m); err != nil {
		t.Fatalf("InsertManual: %v", err)
	}

	if m.ID == 0 {
		t.Fatalf("want ID populated")
	}
	if m.Description != "guia de instalacao" {
		t.Fatalf("want description %q, got %q", "guia de instalacao", m.Description)
	}
	if m.CreatedAt.IsZero() || m.UpdatedAt.IsZero() {
		t.Fatalf("want timestamps populated")
	}

	got, err := r.GetManual(m.ID)
	if err != nil {
		t.Fatalf("GetManual: %v", err)
	}
	if got.ID != m.ID || got.Name != "guia" || got.Body != "passos" || got.Description != "guia de instalacao" {
		t.Fatalf("data not persisted correctly: %+v", got)
	}
}

func TestInsertManualWithoutDescription(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	m := &Manual{Name: "guia", Body: "passos"}
	if err := r.InsertManual(m); err != nil {
		t.Fatalf("InsertManual: %v", err)
	}

	if m.Description != "" {
		t.Fatalf("want an empty description, got %q", m.Description)
	}

	got, err := r.GetManual(m.ID)
	if err != nil {
		t.Fatalf("GetManual: %v", err)
	}
	if got.Description != "" {
		t.Fatalf("want an empty description persisted, got %q", got.Description)
	}
}

func TestInsertManualMissingName(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	err := r.InsertManual(&Manual{Body: "x"})
	if err == nil || !errors.Is(err, ErrManualNameRequired) {
		t.Fatalf("want ErrManualNameRequired, got %v", err)
	}
}

func TestInsertManualMissingBody(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	err := r.InsertManual(&Manual{Name: "x"})
	if err == nil || !errors.Is(err, ErrManualBodyRequired) {
		t.Fatalf("want ErrManualBodyRequired, got %v", err)
	}
}

func TestInsertManualPropagatesAClosedStore(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	err := r.InsertManual(&Manual{Name: "a", Body: "b"})
	if !errors.Is(err, store.ErrClosed) {
		t.Fatalf("want store.ErrClosed, got %v", err)
	}
}

func TestInsertManualIsRejectedInsideATransaction(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	err := s.WithTx(func(tx *store.Tx) error {
		return r.InsertManual(&Manual{Name: "a", Body: "b"})
	})
	if !errors.Is(err, store.ErrTxActive) {
		t.Fatalf("want store.ErrTxActive, got %v", err)
	}
}

func TestUpdateManualSuccess(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	m := &Manual{Name: "old", Body: "old", Description: "descricao antiga"}
	if err := r.InsertManual(m); err != nil {
		t.Fatalf("InsertManual: %v", err)
	}

	created := m.CreatedAt
	m.Name = "new"
	m.Body = "new"
	m.Description = "descricao nova"
	if err := r.UpdateManual(m); err != nil {
		t.Fatalf("UpdateManual: %v", err)
	}

	if m.CreatedAt.Unix() != created.Unix() {
		t.Fatalf("CreatedAt must not change")
	}
	if m.Description != "descricao nova" {
		t.Fatalf("want description %q, got %q", "descricao nova", m.Description)
	}

	got, err := r.GetManual(m.ID)
	if err != nil {
		t.Fatalf("GetManual: %v", err)
	}
	if got.Name != "new" || got.Body != "new" || got.Description != "descricao nova" {
		t.Fatalf("update did not persist: %+v", got)
	}
}

func TestUpdateManualClearsTheDescription(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	m := &Manual{Name: "guia", Body: "passos", Description: "descricao que sera removida"}
	if err := r.InsertManual(m); err != nil {
		t.Fatalf("InsertManual: %v", err)
	}

	m.Description = ""
	if err := r.UpdateManual(m); err != nil {
		t.Fatalf("UpdateManual: %v", err)
	}
	if m.Description != "" {
		t.Fatalf("want an empty description on the struct, got %q", m.Description)
	}

	got, err := r.GetManual(m.ID)
	if err != nil {
		t.Fatalf("GetManual: %v", err)
	}
	if got.Description != "" {
		t.Fatalf("want an empty description persisted, got %q", got.Description)
	}
}

func TestUpdateManualIDRequired(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	err := r.UpdateManual(&Manual{Name: "a", Body: "b"})
	if err == nil || !errors.Is(err, ErrManualIDRequired) {
		t.Fatalf("want ErrManualIDRequired, got %v", err)
	}
}

func TestUpdateManualMissingName(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	err := r.UpdateManual(&Manual{ID: 1, Body: "b"})
	if err == nil || !errors.Is(err, ErrManualNameRequired) {
		t.Fatalf("want ErrManualNameRequired, got %v", err)
	}
}

func TestUpdateManualMissingBody(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	err := r.UpdateManual(&Manual{ID: 1, Name: "a"})
	if err == nil || !errors.Is(err, ErrManualBodyRequired) {
		t.Fatalf("want ErrManualBodyRequired, got %v", err)
	}
}

func TestUpdateManualNotFound(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	err := r.UpdateManual(&Manual{ID: 123, Name: "a", Body: "b"})
	if err == nil || !errors.Is(err, ErrManualNotFound) {
		t.Fatalf("want ErrManualNotFound, got %v", err)
	}
}

func TestUpdateManualPropagatesAClosedStore(t *testing.T) {
	s := newTestStore(t)
	r := New(s)
	m := &Manual{Name: "a", Body: "b"}
	if err := r.InsertManual(m); err != nil {
		t.Fatalf("InsertManual: %v", err)
	}

	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	m.Name = "x"
	err := r.UpdateManual(m)
	if !errors.Is(err, store.ErrClosed) {
		t.Fatalf("want store.ErrClosed, got %v", err)
	}
}

func TestUpdateManualIsRejectedInsideATransaction(t *testing.T) {
	s := newTestStore(t)
	r := New(s)
	m := &Manual{Name: "a", Body: "b"}
	if err := r.InsertManual(m); err != nil {
		t.Fatalf("InsertManual: %v", err)
	}

	m.Name = "x"

	err := s.WithTx(func(tx *store.Tx) error {
		return r.UpdateManual(m)
	})
	if !errors.Is(err, store.ErrTxActive) {
		t.Fatalf("want store.ErrTxActive, got %v", err)
	}
}

func TestDeleteManualSuccess(t *testing.T) {
	s := newTestStore(t)
	r := New(s)
	id, _ := insertManual(t, s, "del", "body", "")

	if err := r.DeleteManual(id); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	got, err := r.GetManual(id)
	if err != nil {
		t.Fatalf("GetManual: %v", err)
	}
	if got != (Manual{}) {
		t.Fatalf("want the manual removed, got %+v", got)
	}
}

func TestDeleteManualIDRequired(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	err := r.DeleteManual(0)
	if err == nil || !errors.Is(err, ErrManualIDRequired) {
		t.Fatalf("want ErrManualIDRequired, got %v", err)
	}
}

func TestDeleteManualNotFound(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	if err := r.DeleteManual(999); err != nil {
		t.Fatalf("deleting a nonexistent ID should return nil, got %v", err)
	}
}

func TestDeleteManualPropagatesAClosedStore(t *testing.T) {
	s := newTestStore(t)
	r := New(s)
	id, _ := insertManual(t, s, "del", "body", "")

	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	err := r.DeleteManual(id)
	if !errors.Is(err, store.ErrClosed) {
		t.Fatalf("want store.ErrClosed, got %v", err)
	}
}

func TestDeleteManualIsRejectedInsideATransaction(t *testing.T) {
	s := newTestStore(t)
	r := New(s)
	id, _ := insertManual(t, s, "del", "body", "")

	err := s.WithTx(func(tx *store.Tx) error {
		return r.DeleteManual(id)
	})
	if !errors.Is(err, store.ErrTxActive) {
		t.Fatalf("want store.ErrTxActive, got %v", err)
	}
}
