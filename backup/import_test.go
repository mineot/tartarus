package backup

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tartarus/repositories"
	"tartarus/store"
)

func write(t *testing.T, body string) string {
	t.Helper()

	path := exportPath(t)

	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	return path
}

func TestImportRejectsAMissingFile(t *testing.T) {
	s := newTestStore(t)

	err := Import(repositories.New(s), filepath.Join(t.TempDir(), "missing.json"))

	if err == nil {
		t.Fatalf("want an error reading a missing file")
	}

	if !strings.Contains(err.Error(), "backup: reading") {
		t.Fatalf("want a read error, got %v", err)
	}
}

func TestImportRejectsMalformedJSON(t *testing.T) {
	s := newTestStore(t)

	path := write(t, `{"commands": [`)

	err := Import(repositories.New(s), path)

	if err == nil {
		t.Fatalf("want an error unmarshalling malformed json")
	}

	if !strings.Contains(err.Error(), "backup: unmarshalling") {
		t.Fatalf("want an unmarshal error, got %v", err)
	}
}

func TestImportOnAFileWithNothingInIt(t *testing.T) {
	s := newTestStore(t)
	r := repositories.New(s)

	path := write(t, `{"commands": [], "manuals": []}`)

	if err := Import(r, path); err != nil {
		t.Fatalf("Import: %v", err)
	}

	commands, err := r.GetCommands()

	if err != nil {
		t.Fatalf("GetCommands: %v", err)
	}

	if len(commands) != 0 {
		t.Fatalf("want zero commands, got %d", len(commands))
	}

	manuals, err := r.GetManuals()

	if err != nil {
		t.Fatalf("GetManuals: %v", err)
	}

	if len(manuals) != 0 {
		t.Fatalf("want zero manuals, got %d", len(manuals))
	}
}

func TestImportInsertsCommandsWithTheirItems(t *testing.T) {
	s := newTestStore(t)
	r := repositories.New(s)

	path := write(t, `{
		"commands": [
			{
				"id": 7,
				"name": "build",
				"description": "compila o projeto",
				"items": [
					{"id": 20, "command_id": 7, "script": "go build ./...", "description": "compila"},
					{"id": 21, "command_id": 7, "script": "make build", "description": "compila via make"}
				],
				"created_at": "2020-01-01T00:00:00Z",
				"updated_at": "2020-01-01T00:00:00Z"
			},
			{
				"id": 9,
				"name": "test",
				"items": [
					{"id": 22, "command_id": 9, "script": "go test ./...", "description": "testa"}
				],
				"created_at": "2020-01-01T00:00:00Z",
				"updated_at": "2020-01-01T00:00:00Z"
			}
		],
		"manuals": []
	}`)

	if err := Import(r, path); err != nil {
		t.Fatalf("Import: %v", err)
	}

	commands, err := r.GetCommands()

	if err != nil {
		t.Fatalf("GetCommands: %v", err)
	}

	if len(commands) != 2 {
		t.Fatalf("want 2 commands, got %d", len(commands))
	}

	first := commands[0]

	if first.Name != "build" || first.Description != "compila o projeto" {
		t.Fatalf("command not imported: %+v", first)
	}

	items, err := r.GetCommandItems(first.ID)

	if err != nil {
		t.Fatalf("GetCommandItems: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("want 2 items, got %d", len(items))
	}

	for i, want := range []struct{ script, description string }{
		{"go build ./...", "compila"},
		{"make build", "compila via make"},
	} {
		if items[i].Script != want.script || items[i].Description != want.description {
			t.Fatalf("item %d not imported: %+v", i, items[i])
		}

		if items[i].CommandID != first.ID {
			t.Fatalf("want command_id %d, got %d", first.ID, items[i].CommandID)
		}
	}

	second := commands[1]

	if second.Name != "test" || second.Description != "" {
		t.Fatalf("second command not imported: %+v", second)
	}

	secondItems, err := r.GetCommandItems(second.ID)

	if err != nil {
		t.Fatalf("GetCommandItems: %v", err)
	}

	if len(secondItems) != 1 || secondItems[0].Script != "go test ./..." {
		t.Fatalf("second command items not imported: %+v", secondItems)
	}

	if second.ID == first.ID {
		t.Fatalf("want two distinct command ids, both are %d", first.ID)
	}
}

func TestImportInsertsManuals(t *testing.T) {
	s := newTestStore(t)
	r := repositories.New(s)

	path := write(t, `{
		"commands": [],
		"manuals": [
			{
				"id": 3,
				"name": "instalacao",
				"body": "passos",
				"description": "primeiro",
				"created_at": "2020-01-01T00:00:00Z",
				"updated_at": "2020-01-01T00:00:00Z"
			},
			{
				"id": 4,
				"name": "uso",
				"body": "mais passos",
				"created_at": "2020-01-01T00:00:00Z",
				"updated_at": "2020-01-01T00:00:00Z"
			}
		]
	}`)

	if err := Import(r, path); err != nil {
		t.Fatalf("Import: %v", err)
	}

	manuals, err := r.GetManuals()

	if err != nil {
		t.Fatalf("GetManuals: %v", err)
	}

	if len(manuals) != 2 {
		t.Fatalf("want 2 manuals, got %d", len(manuals))
	}

	if manuals[0].Name != "instalacao" || manuals[0].Body != "passos" || manuals[0].Description != "primeiro" {
		t.Fatalf("manual not imported: %+v", manuals[0])
	}

	if manuals[1].Name != "uso" || manuals[1].Body != "mais passos" || manuals[1].Description != "" {
		t.Fatalf("second manual not imported: %+v", manuals[1])
	}
}

func TestImportAssignsItsOwnIdsAndTimestamps(t *testing.T) {
	s := newTestStore(t)
	r := repositories.New(s)

	path := write(t, `{
		"commands": [
			{
				"id": 7,
				"name": "build",
				"created_at": "2020-01-01T00:00:00Z",
				"updated_at": "2020-01-01T00:00:00Z"
			}
		],
		"manuals": []
	}`)

	if err := Import(r, path); err != nil {
		t.Fatalf("Import: %v", err)
	}

	commands, err := r.GetCommands()

	if err != nil {
		t.Fatalf("GetCommands: %v", err)
	}

	if len(commands) != 1 {
		t.Fatalf("want 1 command, got %d", len(commands))
	}

	if commands[0].ID != 1 {
		t.Fatalf("want id 1 from the reset database, got %d", commands[0].ID)
	}

	if commands[0].CreatedAt.Year() == 2020 {
		t.Fatalf("want an import time timestamp, got %v", commands[0].CreatedAt)
	}

	if commands[0].CreatedAt.IsZero() || commands[0].UpdatedAt.IsZero() {
		t.Fatalf("want timestamps populated: %+v", commands[0])
	}
}

func TestImportReplacesTheExistingDatabase(t *testing.T) {
	s := newTestStore(t)
	r := repositories.New(s)

	if err := r.InsertCommand(&repositories.Command{Name: "antigo", Description: "sera apagado"}); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	if err := r.InsertManual(&repositories.Manual{Name: "antigo", Body: "sera apagado"}); err != nil {
		t.Fatalf("InsertManual: %v", err)
	}

	path := write(t, `{
		"commands": [{"id": 1, "name": "novo", "items": []}],
		"manuals": []
	}`)

	if err := Import(r, path); err != nil {
		t.Fatalf("Import: %v", err)
	}

	commands, err := r.GetCommands()

	if err != nil {
		t.Fatalf("GetCommands: %v", err)
	}

	if len(commands) != 1 || commands[0].Name != "novo" {
		t.Fatalf("want only the imported command, got %+v", commands)
	}

	manuals, err := r.GetManuals()

	if err != nil {
		t.Fatalf("GetManuals: %v", err)
	}

	if len(manuals) != 0 {
		t.Fatalf("want the existing manuals gone, got %+v", manuals)
	}
}

func TestImportRejectsAnInvalidFileWithoutTouchingTheDatabase(t *testing.T) {
	cases := []struct {
		name string
		body string
		want string
	}{
		{
			name: "empty command name",
			body: `{"commands": [{"id": 1, "name": "", "items": []}], "manuals": []}`,
			want: "no name",
		},
		{
			name: "empty script",
			body: `{"commands": [{"id": 1, "name": "build", "items": [{"id": 1, "script": ""}]}], "manuals": []}`,
			want: "no script",
		},
		{
			name: "duplicate command name",
			body: `{"commands": [{"id": 1, "name": "build", "items": []}, {"id": 2, "name": "build", "items": []}], "manuals": []}`,
			want: "appears more than once",
		},
		{
			name: "duplicate script across commands",
			body: `{"commands": [{"id": 1, "name": "build", "items": [{"id": 1, "script": "ls"}]}, {"id": 2, "name": "test", "items": [{"id": 2, "script": "ls"}]}], "manuals": []}`,
			want: "unique across all commands",
		},
		{
			name: "empty manual name",
			body: `{"commands": [], "manuals": [{"id": 1, "name": "", "body": "corpo"}]}`,
			want: "no name",
		},
		{
			name: "empty manual body",
			body: `{"commands": [], "manuals": [{"id": 1, "name": "guia", "body": ""}]}`,
			want: "no body",
		},
		{
			name: "duplicate manual name",
			body: `{"commands": [], "manuals": [{"id": 1, "name": "guia", "body": "a"}, {"id": 2, "name": "guia", "body": "b"}]}`,
			want: "appears more than once",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := newTestStore(t)
			r := repositories.New(s)

			existing := &repositories.Command{Name: "existente", Description: "deve sobreviver"}
			if err := r.InsertCommand(existing); err != nil {
				t.Fatalf("InsertCommand: %v", err)
			}

			err := Import(r, write(t, tc.body))

			if err == nil {
				t.Fatalf("want a validation error")
			}

			if !strings.Contains(err.Error(), "backup: validating") {
				t.Fatalf("want a validation error, got %v", err)
			}

			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("want an error mentioning %q, got %v", tc.want, err)
			}

			commands, err := r.GetCommands()

			if err != nil {
				t.Fatalf("GetCommands: %v", err)
			}

			if len(commands) != 1 || commands[0].ID != existing.ID || commands[0].Description != "deve sobreviver" {
				t.Fatalf("want the existing command untouched, got %+v", commands)
			}
		})
	}
}

func TestImportPropagatesAClosedStore(t *testing.T) {
	s := newTestStore(t)
	r := repositories.New(s)

	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	err := Import(r, write(t, `{"commands": [], "manuals": []}`))

	if !errors.Is(err, store.ErrClosed) {
		t.Fatalf("want store.ErrClosed, got %v", err)
	}
}

func TestImportIsRejectedInsideATransaction(t *testing.T) {
	s := newTestStore(t)
	r := repositories.New(s)

	err := s.WithTx(func(tx *store.Tx) error {
		return Import(r, write(t, `{"commands": [], "manuals": []}`))
	})

	if !errors.Is(err, store.ErrTxActive) {
		t.Fatalf("want store.ErrTxActive, got %v", err)
	}
}

func TestExportThenImportReproducesTheDatabase(t *testing.T) {
	source := newTestStore(t)
	src := repositories.New(source)

	build := &repositories.Command{Name: "build", Description: "compila o projeto"}
	if err := src.InsertCommand(build); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	if err := src.AppendCommandItem(build.ID, &repositories.CommandItem{
		Script:      "go build ./...",
		Description: "compila",
	}); err != nil {
		t.Fatalf("AppendCommandItem: %v", err)
	}

	if err := src.AppendCommandItem(build.ID, &repositories.CommandItem{Script: "make build"}); err != nil {
		t.Fatalf("AppendCommandItem: %v", err)
	}

	if err := src.InsertManual(&repositories.Manual{
		Name:        "instalacao",
		Body:        "passos",
		Description: "primeiro",
	}); err != nil {
		t.Fatalf("InsertManual: %v", err)
	}

	path := exportPath(t)

	if err := Export(src, path); err != nil {
		t.Fatalf("Export: %v", err)
	}

	target := newTestStore(t)
	dst := repositories.New(target)

	if err := Import(dst, path); err != nil {
		t.Fatalf("Import: %v", err)
	}

	before := readExport(t, path)

	afterPath := exportPath(t)

	if err := Export(dst, afterPath); err != nil {
		t.Fatalf("Export: %v", err)
	}

	after := readExport(t, afterPath)

	if len(after.Commands) != len(before.Commands) || len(after.Manuals) != len(before.Manuals) {
		t.Fatalf("counts differ: before %+v, after %+v", before, after)
	}

	for i, want := range before.Commands {
		got := after.Commands[i]

		if got.Name != want.Name || got.Description != want.Description {
			t.Fatalf("command %d differs: want %+v, got %+v", i, want, got)
		}

		if len(got.Items) != len(want.Items) {
			t.Fatalf("command %d item count differs: want %d, got %d", i, len(want.Items), len(got.Items))
		}

		for j, wantItem := range want.Items {
			gotItem := got.Items[j]

			if gotItem.Script != wantItem.Script || gotItem.Description != wantItem.Description {
				t.Fatalf("item %d/%d differs: want %+v, got %+v", i, j, wantItem, gotItem)
			}

			// The id is not compared against the source: Import assigns its own.
			// What has to hold is that the item points at the command it is
			// nested under, which is the link the source file only implied.
			if gotItem.CommandID != got.ID {
				t.Fatalf("item %d/%d points at command %d, nested under %d", i, j, gotItem.CommandID, got.ID)
			}
		}
	}

	for i, want := range before.Manuals {
		got := after.Manuals[i]

		if got.Name != want.Name || got.Body != want.Body || got.Description != want.Description {
			t.Fatalf("manual %d differs: want %+v, got %+v", i, want, got)
		}
	}
}
