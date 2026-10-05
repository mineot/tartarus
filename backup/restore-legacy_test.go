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

func legacyPath(t *testing.T) string {
	t.Helper()

	return filepath.Join("testdata", "legacy.json")
}

func legacyFixture(t *testing.T, r *repositories.Repos) {
	t.Helper()

	if err := RestoreFromLegacy(r, legacyPath(t)); err != nil {
		t.Fatalf("RestoreFromLegacy: %v", err)
	}
}

func itemScripts(t *testing.T, r *repositories.Repos, name string) []string {
	t.Helper()

	commands, err := r.GetCommands()

	if err != nil {
		t.Fatalf("GetCommands: %v", err)
	}

	for _, c := range commands {
		if c.Name != name {
			continue
		}

		items, err := r.GetCommandItems(c.ID)

		if err != nil {
			t.Fatalf("GetCommandItems: %v", err)
		}

		scripts := make([]string, 0, len(items))

		for _, item := range items {
			scripts = append(scripts, item.Script)
		}

		return scripts
	}

	t.Fatalf("command %q not found", name)

	return nil
}

func commandNames(t *testing.T, r *repositories.Repos) []string {
	t.Helper()

	commands, err := r.GetCommands()

	if err != nil {
		t.Fatalf("GetCommands: %v", err)
	}

	names := make([]string, 0, len(commands))

	for _, c := range commands {
		names = append(names, c.Name)
	}

	return names
}

func manualNames(t *testing.T, r *repositories.Repos) []string {
	t.Helper()

	manuals, err := r.GetManuals()

	if err != nil {
		t.Fatalf("GetManuals: %v", err)
	}

	names := make([]string, 0, len(manuals))

	for _, m := range manuals {
		names = append(names, m.Name)
	}

	return names
}

func TestRestoreFromLegacyRejectsAMissingFile(t *testing.T) {
	s := newTestStore(t)

	err := RestoreFromLegacy(repositories.New(s), filepath.Join(t.TempDir(), "missing.json"))

	if err == nil {
		t.Fatalf("want an error reading a missing file")
	}

	if !strings.Contains(err.Error(), "backup: reading") {
		t.Fatalf("want a read error, got %v", err)
	}
}

func TestRestoreFromLegacyRejectsMalformedJSON(t *testing.T) {
	s := newTestStore(t)

	err := RestoreFromLegacy(repositories.New(s), write(t, `[{"_id": "cmd:build"`))

	if err == nil {
		t.Fatalf("want an error unmarshalling malformed json")
	}

	if !strings.Contains(err.Error(), "backup: unmarshalling") {
		t.Fatalf("want an unmarshal error, got %v", err)
	}
}

func TestRestoreFromLegacyRejectsAnObjectInsteadOfAnArray(t *testing.T) {
	s := newTestStore(t)

	err := RestoreFromLegacy(repositories.New(s), write(t, `{"commands": []}`))

	if err == nil {
		t.Fatalf("want an error unmarshalling an object into an array")
	}

	if !strings.Contains(err.Error(), "backup: unmarshalling") {
		t.Fatalf("want an unmarshal error, got %v", err)
	}
}

func TestRestoreFromLegacyImportsCommandsWithTheirInstructions(t *testing.T) {
	s := newTestStore(t)
	r := repositories.New(s)

	legacyFixture(t, r)

	commands, err := r.GetCommands()

	if err != nil {
		t.Fatalf("GetCommands: %v", err)
	}

	if len(commands) != 4 {
		t.Fatalf("want four commands, got %d: %v", len(commands), commandNames(t, r))
	}

	want := []string{
		"gdrive-refresh-connection",
		"linux_upgrade",
		"vpn_gr",
		"vpn_us",
	}

	got := commandNames(t, r)

	for i, name := range want {
		if got[i] != name {
			t.Fatalf("want command %d to be %q, got %q", i, name, got[i])
		}
	}

	if commands[1].Description != "Atualização geral do Linux" {
		t.Fatalf("want the legacy description, got %q", commands[1].Description)
	}

	if commands[0].Description != "" {
		t.Fatalf("want an empty description for a document without one, got %q", commands[0].Description)
	}
}

func TestRestoreFromLegacyKeepsTheOrderOfTheInstructions(t *testing.T) {
	s := newTestStore(t)
	r := repositories.New(s)

	legacyFixture(t, r)

	want := []string{
		"sudo apt update",
		"sudo apt upgrade",
		"sudo apt dist-upgrade",
		"sudo apt full-upgrade",
		"sudo apt autoremove",
		"sudo snap refresh",
		"composer global update",
	}

	got := itemScripts(t, r, "linux_upgrade")

	if len(got) != len(want) {
		t.Fatalf("want %d items, got %d: %v", len(want), len(got), got)
	}

	for i, script := range want {
		if got[i] != script {
			t.Fatalf("want item %d to be %q, got %q", i, script, got[i])
		}
	}
}

func TestRestoreFromLegacyAttachesEachItemToItsOwnCommand(t *testing.T) {
	s := newTestStore(t)
	r := repositories.New(s)

	legacyFixture(t, r)

	commands, err := r.GetCommands()

	if err != nil {
		t.Fatalf("GetCommands: %v", err)
	}

	byName := make(map[string]uint64, len(commands))

	for _, c := range commands {
		byName[c.Name] = c.ID
	}

	for _, name := range []string{"gdrive-refresh-connection", "linux_upgrade", "vpn_gr", "vpn_us"} {
		items, err := r.GetCommandItems(byName[name])

		if err != nil {
			t.Fatalf("GetCommandItems: %v", err)
		}

		if len(items) == 0 {
			t.Fatalf("want items on command %q", name)
		}

		for _, item := range items {
			if item.CommandID != byName[name] {
				t.Fatalf("want item %q on command %q, got command id %d",
					item.Script, name, item.CommandID)
			}
		}
	}
}

func TestRestoreFromLegacyImportsManuals(t *testing.T) {
	s := newTestStore(t)
	r := repositories.New(s)

	legacyFixture(t, r)

	manuals, err := r.GetManuals()

	if err != nil {
		t.Fatalf("GetManuals: %v", err)
	}

	if len(manuals) != 1 {
		t.Fatalf("want one manual, got %d", len(manuals))
	}

	if manuals[0].Name != "teste" {
		t.Fatalf("want the name after the prefix, got %q", manuals[0].Name)
	}

	if manuals[0].Body != "teste\n\nteste\n\nteste\n" {
		t.Fatalf("want the content verbatim, got %q", manuals[0].Body)
	}
}

func TestRestoreFromLegacySkipsConfigEditor(t *testing.T) {
	s := newTestStore(t)
	r := repositories.New(s)

	legacyFixture(t, r)

	for _, name := range commandNames(t, r) {
		if strings.HasPrefix(name, "config") {
			t.Fatalf("want config:editor left out, got command %q", name)
		}
	}

	if len(manualNames(t, r)) != 1 {
		t.Fatalf("want config:editor to become neither a command nor a manual")
	}
}

func TestRestoreFromLegacySkipsUnknownPrefixes(t *testing.T) {
	s := newTestStore(t)
	r := repositories.New(s)

	path := write(t, `[
		{"_id": "cmd:build", "instructions": ["go build ./..."]},
		{"_id": "config:editor", "data": "nano"},
		{"_id": "view:commands", "map": "function(doc) {}"},
		{"_id": "_design/commands"},
		{"_id": "_local/thing"}
	]`)

	if err := RestoreFromLegacy(r, path); err != nil {
		t.Fatalf("RestoreFromLegacy: %v", err)
	}

	if got := commandNames(t, r); len(got) != 1 || got[0] != "build" {
		t.Fatalf("want only the command document, got %v", got)
	}

	if got := manualNames(t, r); len(got) != 0 {
		t.Fatalf("want no manuals, got %v", got)
	}
}

func TestRestoreFromLegacyIsANoOpOnASecondRun(t *testing.T) {
	s := newTestStore(t)
	r := repositories.New(s)

	legacyFixture(t, r)

	before, err := r.GetCommands()

	if err != nil {
		t.Fatalf("GetCommands: %v", err)
	}

	legacyFixture(t, r)

	after, err := r.GetCommands()

	if err != nil {
		t.Fatalf("GetCommands: %v", err)
	}

	if len(after) != len(before) {
		t.Fatalf("want %d commands after a second run, got %d", len(before), len(after))
	}

	for i, c := range after {
		if c.ID != before[i].ID {
			t.Fatalf("want command %d to keep id %d, got %d", i, before[i].ID, c.ID)
		}

		items, err := r.GetCommandItems(c.ID)

		if err != nil {
			t.Fatalf("GetCommandItems: %v", err)
		}

		beforeItems, err := r.GetCommandItems(before[i].ID)

		if err != nil {
			t.Fatalf("GetCommandItems: %v", err)
		}

		if len(items) != len(beforeItems) {
			t.Fatalf("want command %q to keep %d items, got %d", c.Name, len(beforeItems), len(items))
		}
	}

	if got := len(manualNames(t, r)); got != 1 {
		t.Fatalf("want one manual after a second run, got %d", got)
	}
}

func TestRestoreFromLegacyCompletesAPartiallyImportedCommand(t *testing.T) {
	s := newTestStore(t)
	r := repositories.New(s)

	path := write(t, `[{"_id": "cmd:linux_upgrade", "instructions": ["sudo apt update", "sudo apt upgrade"]}]`)

	if err := RestoreFromLegacy(r, path); err != nil {
		t.Fatalf("RestoreFromLegacy: %v", err)
	}

	commands, err := r.GetCommands()

	if err != nil {
		t.Fatalf("GetCommands: %v", err)
	}

	items, err := r.GetCommandItems(commands[0].ID)

	if err != nil {
		t.Fatalf("GetCommandItems: %v", err)
	}

	if err = r.RemoveCommandItem(commands[0].ID, items[1].ID); err != nil {
		t.Fatalf("RemoveCommandItem: %v", err)
	}

	if got := itemScripts(t, r, "linux_upgrade"); len(got) != 1 {
		t.Fatalf("want one item left, got %v", got)
	}

	legacyFixture(t, r)

	got := itemScripts(t, r, "linux_upgrade")

	if len(got) != 7 {
		t.Fatalf("want the missing instructions to be added, got %v", got)
	}

	commands, err = r.GetCommands()

	if err != nil {
		t.Fatalf("GetCommands: %v", err)
	}

	if len(commands) != 4 {
		t.Fatalf("want the existing command reused rather than a second one, got %v", commandNames(t, r))
	}
}

func TestRestoreFromLegacyLeavesAnExistingManualAlone(t *testing.T) {
	s := newTestStore(t)
	r := repositories.New(s)

	existing := &repositories.Manual{Name: "teste", Body: "o que esta aqui agora"}
	if err := r.InsertManual(existing); err != nil {
		t.Fatalf("InsertManual: %v", err)
	}

	legacyFixture(t, r)

	manuals, err := r.GetManuals()

	if err != nil {
		t.Fatalf("GetManuals: %v", err)
	}

	if len(manuals) != 1 {
		t.Fatalf("want one manual, got %d", len(manuals))
	}

	if manuals[0].Body != "o que esta aqui agora" {
		t.Fatalf("want the existing body untouched, got %q", manuals[0].Body)
	}
}

func TestRestoreFromLegacySkipsAnItemWhoseScriptBelongsToAnotherCommand(t *testing.T) {
	s := newTestStore(t)
	r := repositories.New(s)

	existing := &repositories.Command{Name: "build", Description: "compila"}
	if err := r.InsertCommand(existing); err != nil {
		t.Fatalf("InsertCommand: %v", err)
	}

	item := &repositories.CommandItem{Script: "sudo apt update"}
	if err := r.AppendCommandItem(existing.ID, item); err != nil {
		t.Fatalf("AppendCommandItem: %v", err)
	}

	legacyFixture(t, r)

	got := itemScripts(t, r, "linux_upgrade")

	if len(got) != 6 {
		t.Fatalf("want the conflicting item skipped, got %d items: %v", len(got), got)
	}

	for _, script := range got {
		if script == "sudo apt update" {
			t.Fatalf("want no copy of the conflicting script, got %v", got)
		}
	}

	buildScripts := itemScripts(t, r, "build")

	if len(buildScripts) != 1 || buildScripts[0] != "sudo apt update" {
		t.Fatalf("want the original command untouched, got %v", buildScripts)
	}
}

func TestRestoreFromLegacyRejectsAnInvalidFileWithoutTouchingTheDatabase(t *testing.T) {
	cases := []struct {
		name string
		body string
		want string
	}{
		{
			name: "empty command name",
			body: `[{"_id": "cmd:", "instructions": ["ls"]}]`,
			want: "has no name",
		},
		{
			name: "empty script",
			body: `[{"_id": "cmd:build", "instructions": [""]}]`,
			want: "no script",
		},
		{
			name: "duplicate command name",
			body: `[{"_id": "cmd:build", "instructions": ["ls"]}, {"_id": "cmd:build", "instructions": ["pwd"]}]`,
			want: "appears more than once",
		},
		{
			name: "duplicate script within one command",
			body: `[{"_id": "cmd:build", "instructions": ["ls", "ls"]}]`,
			want: "appears more than once",
		},
		{
			name: "duplicate script across commands",
			body: `[{"_id": "cmd:build", "instructions": ["ls"]}, {"_id": "cmd:test", "instructions": ["ls"]}]`,
			want: "unique across all commands",
		},
		{
			name: "empty manual name",
			body: `[{"_id": "manual:", "content": "corpo"}]`,
			want: "has no name",
		},
		{
			name: "empty manual content",
			body: `[{"_id": "manual:guia", "content": ""}]`,
			want: "no content",
		},
		{
			name: "duplicate manual name",
			body: `[{"_id": "manual:guia", "content": "a"}, {"_id": "manual:guia", "content": "b"}]`,
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

			err := RestoreFromLegacy(r, write(t, tc.body))

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

			if len(commands) != 1 || commands[0].ID != existing.ID ||
				commands[0].Description != "deve sobreviver" {
				t.Fatalf("want the existing command untouched, got %+v", commands)
			}
		})
	}
}

func TestRestoreFromLegacyOnAFileWithNothingInIt(t *testing.T) {
	s := newTestStore(t)
	r := repositories.New(s)

	if err := RestoreFromLegacy(r, write(t, `[]`)); err != nil {
		t.Fatalf("RestoreFromLegacy: %v", err)
	}

	if got := commandNames(t, r); len(got) != 0 {
		t.Fatalf("want zero commands, got %v", got)
	}

	if got := manualNames(t, r); len(got) != 0 {
		t.Fatalf("want zero manuals, got %v", got)
	}
}

func TestRestoreFromLegacyPropagatesAClosedStore(t *testing.T) {
	s := newTestStore(t)

	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	err := RestoreFromLegacy(repositories.New(s), legacyPath(t))

	if !errors.Is(err, store.ErrClosed) {
		t.Fatalf("want store.ErrClosed, got %v", err)
	}
}

func TestRestoreFromLegacyIsRejectedInsideATransaction(t *testing.T) {
	s := newTestStore(t)
	r := repositories.New(s)

	err := s.WithTx(func(tx *store.Tx) error {
		return RestoreFromLegacy(r, legacyPath(t))
	})

	if !errors.Is(err, store.ErrUseTx) {
		t.Fatalf("want store.ErrUseTx, got %v", err)
	}
}

func TestRestoreFromLegacyDoesNotWriteTheFixture(t *testing.T) {
	before, err := os.ReadFile(legacyPath(t))

	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	s := newTestStore(t)
	r := repositories.New(s)

	legacyFixture(t, r)

	after, err := os.ReadFile(legacyPath(t))

	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if string(before) != string(after) {
		t.Fatalf("want the export file untouched by a restore")
	}
}
