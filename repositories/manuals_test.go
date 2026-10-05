package repositories

import (
	"errors"
	"testing"
	"time"

	"tartarus/store"
)

// insertManual writes a manual and returns the timestamp it was stamped with,
// which the caller needs in order to compare what comes back out.
//
// The description is written as SQL NULL when empty, which is the state of any
// row that predates the column. The repository itself always sends a plain
// string, so an empty description reaches the database as an empty string.
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
		t.Fatalf("esperava zero manuais, obteve %d", len(manuals))
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
		t.Fatalf("esperava %d manuais, obteve %d", len(want), len(manuals))
	}

	for i, w := range want {
		got := manuals[i]

		if got.ID != w.id {
			t.Fatalf("posição %d: esperava o id %d, obteve %d", i, w.id, got.ID)
		}
		if got.Name != w.name {
			t.Fatalf("posição %d: esperava o nome %q, obteve %q", i, w.name, got.Name)
		}
		if got.Body != w.body {
			t.Fatalf("posição %d: esperava o corpo %q, obteve %q", i, w.body, got.Body)
		}
		if got.Description != w.description {
			t.Fatalf("posição %d: esperava a description %q, obteve %q", i, w.description, got.Description)
		}
		if got.CreatedAt.Unix() != w.at.Unix() {
			t.Fatalf("posição %d: esperava created_at %d, obteve %d", i, w.at.Unix(), got.CreatedAt.Unix())
		}
		if got.UpdatedAt.Unix() != w.at.Unix() {
			t.Fatalf("posição %d: esperava updated_at %d, obteve %d", i, w.at.Unix(), got.UpdatedAt.Unix())
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
		t.Fatalf("esperava store.ErrUseTx, obteve %v", err)
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
		t.Fatalf("esperava store.ErrClosed, obteve %v", err)
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
		t.Fatalf("esperava ID %d, obteve %d", id, got.ID)
	}
	if got.Name != "readme" {
		t.Fatalf("esperava nome %q, obteve %q", "readme", got.Name)
	}
	if got.Body != "conteudo" {
		t.Fatalf("esperava corpo %q, obteve %q", "conteudo", got.Body)
	}
	if got.Description != "como usar o comando" {
		t.Fatalf("esperava description %q, obteve %q", "como usar o comando", got.Description)
	}
	if got.CreatedAt.Unix() != at.Unix() {
		t.Fatalf("esperava created_at %d, obteve %d", at.Unix(), got.CreatedAt.Unix())
	}
	if got.UpdatedAt.Unix() != at.Unix() {
		t.Fatalf("esperava updated_at %d, obteve %d", at.Unix(), got.UpdatedAt.Unix())
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
		t.Fatalf("esperava description vazia para NULL, obteve %q", got.Description)
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
		t.Fatalf("esperava manual vazio, obteve %+v", got)
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
		t.Fatalf("esperava store.ErrClosed, obteve %v", err)
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
		t.Fatalf("esperava store.ErrUseTx, obteve %v", err)
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
		t.Fatalf("esperava ID populado")
	}
	if m.Description != "guia de instalacao" {
		t.Fatalf("esperava description %q, obteve %q", "guia de instalacao", m.Description)
	}
	if m.CreatedAt.IsZero() || m.UpdatedAt.IsZero() {
		t.Fatalf("esperava timestamps populados")
	}

	got, err := r.GetManual(m.ID)
	if err != nil {
		t.Fatalf("GetManual: %v", err)
	}
	if got.ID != m.ID || got.Name != "guia" || got.Body != "passos" || got.Description != "guia de instalacao" {
		t.Fatalf("dados não persistidos corretamente: %+v", got)
	}
}

func TestInsertManualWithoutDescription(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	// A description e nullable e nao tem validacao: um manual sem descricao e
	// valido, e deve sair com o campo vazio em vez de erro.
	m := &Manual{Name: "guia", Body: "passos"}
	if err := r.InsertManual(m); err != nil {
		t.Fatalf("InsertManual: %v", err)
	}

	if m.Description != "" {
		t.Fatalf("esperava description vazia, obteve %q", m.Description)
	}

	got, err := r.GetManual(m.ID)
	if err != nil {
		t.Fatalf("GetManual: %v", err)
	}
	if got.Description != "" {
		t.Fatalf("esperava description vazia persistida, obteve %q", got.Description)
	}
}

func TestInsertManualMissingName(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	err := r.InsertManual(&Manual{Body: "x"})
	if err == nil || !errors.Is(err, ErrManualNameRequired) {
		t.Fatalf("esperava ErrManualNameRequired, obteve %v", err)
	}
}

func TestInsertManualMissingBody(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	err := r.InsertManual(&Manual{Name: "x"})
	if err == nil || !errors.Is(err, ErrManualBodyRequired) {
		t.Fatalf("esperava ErrManualBodyRequired, obteve %v", err)
	}
}

func TestInsertManualPropagatesAClosedStore(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	err := r.InsertManual(&Manual{Name: "a", Body: "b"})
	if err == nil {
		t.Fatalf("esperava erro ao inserir em store fechado")
	}
}

func TestInsertManualIsRejectedInsideATransaction(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	err := s.WithTx(func(tx *store.Tx) error {
		return r.InsertManual(&Manual{Name: "a", Body: "b"})
	})
	if !errors.Is(err, store.ErrTxActive) {
		t.Fatalf("esperava store.ErrTxActive, obteve %v", err)
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
		t.Fatalf("CreatedAt não deve mudar")
	}
	if m.Description != "descricao nova" {
		t.Fatalf("esperava description %q, obteve %q", "descricao nova", m.Description)
	}

	got, err := r.GetManual(m.ID)
	if err != nil {
		t.Fatalf("GetManual: %v", err)
	}
	if got.Name != "new" || got.Body != "new" || got.Description != "descricao nova" {
		t.Fatalf("atualização não persistiu: %+v", got)
	}
}

func TestUpdateManualClearsTheDescription(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	m := &Manual{Name: "guia", Body: "passos", Description: "descricao que sera removida"}
	if err := r.InsertManual(m); err != nil {
		t.Fatalf("InsertManual: %v", err)
	}

	// Description nao tem validacao, entao limpar e um update legitimo.
	m.Description = ""
	if err := r.UpdateManual(m); err != nil {
		t.Fatalf("UpdateManual: %v", err)
	}
	if m.Description != "" {
		t.Fatalf("esperava description vazia na struct, obteve %q", m.Description)
	}

	got, err := r.GetManual(m.ID)
	if err != nil {
		t.Fatalf("GetManual: %v", err)
	}
	if got.Description != "" {
		t.Fatalf("esperava description vazia persistida, obteve %q", got.Description)
	}
}

func TestUpdateManualIDRequired(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	err := r.UpdateManual(&Manual{Name: "a", Body: "b"})
	if err == nil || !errors.Is(err, ErrManualIDRequired) {
		t.Fatalf("esperava ErrManualIDRequired, obteve %v", err)
	}
}

func TestUpdateManualMissingName(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	err := r.UpdateManual(&Manual{ID: 1, Body: "b"})
	if err == nil || !errors.Is(err, ErrManualNameRequired) {
		t.Fatalf("esperava ErrManualNameRequired, obteve %v", err)
	}
}

func TestUpdateManualMissingBody(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	err := r.UpdateManual(&Manual{ID: 1, Name: "a"})
	if err == nil || !errors.Is(err, ErrManualBodyRequired) {
		t.Fatalf("esperava ErrManualBodyRequired, obteve %v", err)
	}
}

func TestUpdateManualNotFound(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	err := r.UpdateManual(&Manual{ID: 123, Name: "a", Body: "b"})
	if err == nil || !errors.Is(err, ErrManualNotFound) {
		t.Fatalf("esperava ErrManualNotFound, obteve %v", err)
	}
}

func TestUpdateManualPropagatesAClosedStore(t *testing.T) {
	s := newTestStore(t)
	r := New(s)
	m := &Manual{Name: "a", Body: "b"}
	_ = r.InsertManual(m)

	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	m.Name = "x"
	err := r.UpdateManual(m)
	if err == nil {
		t.Fatalf("esperava erro ao atualizar em store fechado")
	}
}

func TestUpdateManualIsRejectedInsideATransaction(t *testing.T) {
	s := newTestStore(t)
	r := New(s)
	m := &Manual{Name: "a", Body: "b"}
	_ = r.InsertManual(m)
	m.Name = "x"

	err := s.WithTx(func(tx *store.Tx) error {
		return r.UpdateManual(m)
	})
	if !errors.Is(err, store.ErrTxActive) {
		t.Fatalf("esperava store.ErrTxActive, obteve %v", err)
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
		t.Fatalf("esperava manual removido, obteve %+v", got)
	}
}

func TestDeleteManualIDRequired(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	err := r.DeleteManual(0)
	if err == nil || !errors.Is(err, ErrManualIDRequired) {
		t.Fatalf("esperava ErrManualIDRequired, obteve %v", err)
	}
}

func TestDeleteManualNotFound(t *testing.T) {
	s := newTestStore(t)
	r := New(s)

	if err := r.DeleteManual(999); err != nil {
		t.Fatalf("Delete de ID inexistente deve retornar nil, obteve %v", err)
	}
}

func TestDeleteManualPropagatesAClosedStore(t *testing.T) {
	s := newTestStore(t)
	r := New(s)
	id, _ := insertManual(t, s, "del", "body", "")

	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if err := r.DeleteManual(id); err == nil {
		t.Fatalf("esperava erro ao deletar em store fechado")
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
		t.Fatalf("esperava store.ErrTxActive, obteve %v", err)
	}
}
