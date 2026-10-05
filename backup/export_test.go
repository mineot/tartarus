package backup

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"tartarus/helpers"
	"tartarus/repositories"
	"tartarus/store"
)

func newTestStore(t *testing.T) *store.Store {
	t.Helper()

	if err := helpers.SetDevStorePath(filepath.Join(t.TempDir(), "test.db")); err != nil {
		t.Fatalf("SetDevStorePath: %v", err)
	}

	s, err := store.New(context.Background())

	if err != nil {
		t.Fatalf("store.New: %v", err)
	}

	t.Cleanup(func() { s.Close() })

	if err = s.RunMigrations(); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	return s
}

func exportPath(t *testing.T) string {
	t.Helper()

	return filepath.Join(t.TempDir(), "export.json")
}

func readExport(t *testing.T, path string) jsonFile {
	t.Helper()

	data, err := os.ReadFile(path)

	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	var file jsonFile

	if err = json.Unmarshal(data, &file); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	return file
}

func TestExportOnEmptyDatabaseWritesEmptyArrays(t *testing.T) {
	s := newTestStore(t)
	path := exportPath(t)

	if err := Export(repositories.New(s), path); err != nil {
		t.Fatalf("Export: %v", err)
	}

	data, err := os.ReadFile(path)

	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	want := "{\n  \"commands\": [],\n  \"manuals\": []\n}"

	if string(data) != want {
		t.Fatalf("want %q, got %q", want, string(data))
	}
}

func TestExportWritesCommandsWithTheirItemsInOrder(t *testing.T) {
	s := newTestStore(t)
	r := repositories.New(s)

	first := &repositories.Command{Name: "build", Description: "compila o projeto"}
	if err := r.InsertCommand(first); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	second := &repositories.Command{Name: "test", Description: "roda os testes"}
	if err := r.InsertCommand(second); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	firstItem := &repositories.CommandItem{Script: "go build ./...", Description: "compila"}
	if err := r.AppendCommandItem(first.ID, firstItem); err != nil {
		t.Fatalf("AppendCommandItem: %v", err)
	}

	secondItem := &repositories.CommandItem{Script: "go test ./...", Description: "testa"}
	if err := r.AppendCommandItem(first.ID, secondItem); err != nil {
		t.Fatalf("AppendCommandItem: %v", err)
	}

	otherItem := &repositories.CommandItem{Script: "make test", Description: "testa via make"}
	if err := r.AppendCommandItem(second.ID, otherItem); err != nil {
		t.Fatalf("AppendCommandItem: %v", err)
	}

	path := exportPath(t)

	if err := Export(r, path); err != nil {
		t.Fatalf("Export: %v", err)
	}

	file := readExport(t, path)

	if len(file.Commands) != 2 {
		t.Fatalf("want 2 commands, got %d", len(file.Commands))
	}

	if file.Commands[0].ID != first.ID || file.Commands[0].Name != "build" {
		t.Fatalf("first command not exported in order: %+v", file.Commands[0])
	}

	if file.Commands[0].Description != "compila o projeto" {
		t.Fatalf("want command description %q, got %q", "compila o projeto", file.Commands[0].Description)
	}

	if !file.Commands[0].CreatedAt.Equal(first.CreatedAt) || !file.Commands[0].UpdatedAt.Equal(first.UpdatedAt) {
		t.Fatalf("command timestamps not exported: %+v", file.Commands[0])
	}

	if len(file.Commands[0].Items) != 2 {
		t.Fatalf("want 2 items, got %d", len(file.Commands[0].Items))
	}

	for index, want := range []*repositories.CommandItem{firstItem, secondItem} {
		got := file.Commands[0].Items[index]

		if got.ID != want.ID || got.CommandID != first.ID || got.Script != want.Script {
			t.Fatalf("item %d not exported correctly: %+v", index, got)
		}

		if got.Description != want.Description {
			t.Fatalf("want item description %q, got %q", want.Description, got.Description)
		}

		if !got.CreatedAt.Equal(want.CreatedAt) || !got.UpdatedAt.Equal(want.UpdatedAt) {
			t.Fatalf("item %d timestamps not exported: %+v", index, got)
		}
	}

	if len(file.Commands[1].Items) != 1 || file.Commands[1].Items[0].Script != "make test" {
		t.Fatalf("second command items not exported: %+v", file.Commands[1])
	}

	if file.Commands[1].Items[0].CommandID != second.ID {
		t.Fatalf("want command_id %d, got %d", second.ID, file.Commands[1].Items[0].CommandID)
	}
}

func TestExportWritesACommandWithoutItemsAsAnEmptyArray(t *testing.T) {
	s := newTestStore(t)
	r := repositories.New(s)

	if err := r.InsertCommand(&repositories.Command{Name: "vazio"}); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	path := exportPath(t)

	if err := Export(r, path); err != nil {
		t.Fatalf("Export: %v", err)
	}

	data, err := os.ReadFile(path)

	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if !strings.Contains(string(data), `"items": []`) {
		t.Fatalf("want an empty items array, got %q", string(data))
	}
}

func TestExportWritesManuals(t *testing.T) {
	s := newTestStore(t)
	r := repositories.New(s)

	first := &repositories.Manual{Name: "instalacao", Body: "passos", Description: "primeiro"}
	if err := r.InsertManual(first); err != nil {
		t.Fatalf("InsertManual: %v", err)
	}

	second := &repositories.Manual{Name: "uso", Body: "mais passos"}
	if err := r.InsertManual(second); err != nil {
		t.Fatalf("InsertManual: %v", err)
	}

	path := exportPath(t)

	if err := Export(r, path); err != nil {
		t.Fatalf("Export: %v", err)
	}

	file := readExport(t, path)

	if len(file.Manuals) != 2 {
		t.Fatalf("want 2 manuals, got %d", len(file.Manuals))
	}

	if file.Manuals[0].ID != first.ID || file.Manuals[0].Name != "instalacao" {
		t.Fatalf("first manual not exported in order: %+v", file.Manuals[0])
	}

	if file.Manuals[0].Body != "passos" || file.Manuals[0].Description != "primeiro" {
		t.Fatalf("manual body or description not exported: %+v", file.Manuals[0])
	}

	if !file.Manuals[0].CreatedAt.Equal(first.CreatedAt) || !file.Manuals[0].UpdatedAt.Equal(first.UpdatedAt) {
		t.Fatalf("manual timestamps not exported: %+v", file.Manuals[0])
	}

	if file.Manuals[1].ID != second.ID || file.Manuals[1].Body != "mais passos" {
		t.Fatalf("second manual not exported: %+v", file.Manuals[1])
	}
}

func TestExportWritesAnEmptyDescriptionAsAnEmptyString(t *testing.T) {
	s := newTestStore(t)
	r := repositories.New(s)

	c := &repositories.Command{Name: "sem descricao"}
	if err := r.InsertCommand(c); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	if err := r.InsertManual(&repositories.Manual{Name: "sem descricao", Body: "corpo"}); err != nil {
		t.Fatalf("InsertManual: %v", err)
	}

	if err := r.AppendCommandItem(c.ID, &repositories.CommandItem{Script: "ls"}); err != nil {
		t.Fatalf("AppendCommandItem: %v", err)
	}

	path := exportPath(t)

	if err := Export(r, path); err != nil {
		t.Fatalf("Export: %v", err)
	}

	data, err := os.ReadFile(path)

	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if got := strings.Count(string(data), `"description": ""`); got != 3 {
		t.Fatalf("want 3 empty descriptions, got %d in %q", got, string(data))
	}

	file := readExport(t, path)

	if file.Commands[0].Description != "" || file.Manuals[0].Description != "" {
		t.Fatalf("want empty descriptions: %+v %+v", file.Commands[0], file.Manuals[0])
	}

	if file.Commands[0].Items[0].Description != "" {
		t.Fatalf("want empty item description: %+v", file.Commands[0].Items[0])
	}
}

func TestExportReadsADescriptionStoredAsNull(t *testing.T) {
	s := newTestStore(t)

	now := time.Now().UTC()

	_, err := s.Exec(`INSERT INTO commands (name, description, created_at, updated_at) VALUES (?, NULL, ?, ?)`,
		"nulo", now, now)

	if err != nil {
		t.Fatalf("Exec: %v", err)
	}

	path := exportPath(t)

	if err = Export(repositories.New(s), path); err != nil {
		t.Fatalf("Export: %v", err)
	}

	data, err := os.ReadFile(path)

	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if !strings.Contains(string(data), `"description": ""`) {
		t.Fatalf("want a NULL description exported as empty, got %q", string(data))
	}
}

func TestExportPropagatesAClosedStore(t *testing.T) {
	s := newTestStore(t)
	r := repositories.New(s)

	if err := r.InsertManual(&repositories.Manual{Name: "guia", Body: "passos"}); err != nil {
		t.Fatalf("InsertManual: %v", err)
	}

	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	err := Export(r, exportPath(t))

	if !errors.Is(err, store.ErrClosed) {
		t.Fatalf("want store.ErrClosed, got %v", err)
	}
}

func TestExportIsRejectedInsideATransaction(t *testing.T) {
	s := newTestStore(t)
	r := repositories.New(s)

	err := s.WithTx(func(tx *store.Tx) error {
		return Export(r, exportPath(t))
	})

	if !errors.Is(err, store.ErrUseTx) {
		t.Fatalf("want store.ErrUseTx, got %v", err)
	}
}

func TestExportFailsOnAnUnwritablePath(t *testing.T) {
	s := newTestStore(t)

	path := filepath.Join(t.TempDir(), "missing", "export.json")

	err := Export(repositories.New(s), path)

	if err == nil {
		t.Fatalf("want an error writing to %s", path)
	}

	if !strings.Contains(err.Error(), "backup: writing") {
		t.Fatalf("want a write error, got %v", err)
	}
}
