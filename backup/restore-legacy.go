package backup

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"tartarus/repositories"
)

// legacyDoc is one document of a legacy v1.x export, the file the old
// `tartarus db export` wrote. The format is a flat array with no type field, so
// the prefix of _id is the only thing that says what a document is.
//
// _rev, updatedAt and data are deliberately absent. Ids and timestamps are
// ignored the same way Import ignores them, and data only ever held the editor
// name, which has no destination in the new schema.
type legacyDoc struct {
	ID           string   `json:"_id"`
	Instructions []string `json:"instructions"`
	Description  string   `json:"description"`
	Content      string   `json:"content"`
}

const (
	legacyCommandPrefix = "cmd:"
	legacyManualPrefix  = "manual:"
)

// legacyCommand is a command already in the database, with the scripts it
// holds. The set is what makes a second run converge instead of colliding.
type legacyCommand struct {
	id      uint64
	scripts map[string]struct{}
}

// legacyState is what the database already holds, read once before anything is
// written.
type legacyState struct {
	commands    map[string]legacyCommand
	scriptOwner map[string]string
	manuals     map[string]struct{}
}

// legacyReport counts what a run did, so the caller can be told without having
// to take a second look.
type legacyReport struct {
	commands int
	items    int
	manuals  int
	skipped  int
	notes    []string
}

// RestoreFromLegacy writes the contents of a legacy v1.x export into the
// database: every cmd: document becomes a command with one command item per
// entry of instructions, and every manual: document becomes a manual whose body
// is the document's content. Any other document, config:editor among them, has
// no equivalent here and is skipped.
//
// It is a merge, not a replacement, and unlike Import it resets nothing. The
// legacy folder is never deleted, so the export can be produced again and the
// run repeated; what makes repeating it safe is that this reads the database
// first and only writes what is missing. A command that is already there keeps
// its row and gains the instructions it is missing, which is also what finishes
// a run that was interrupted halfway through a command. A manual that is
// already there is left alone: deciding which of two bodies wins would mean
// comparing timestamps, and Import ignores those too.
//
// The file is validated before the first write. Every check is one the database
// would refuse anyway — commands.name and manuals.name are UNIQUE, and so is
// command_items.script, globally — and the file belongs to the user, not to us
// to rewrite. Reporting it beats half a restore followed by a constraint error.
//
// Nothing here is atomic: every repositories write opens its own transaction,
// and there is no InsertCommandTx to compose them into one. That is tolerable
// precisely because of the convergence above.
//
// It must not be called from inside a store.WithTx callback. The reads come
// first and go through Store.Query, so it fails with store.ErrUseTx, the same
// sentinel Export runs into, rather than the store.ErrTxActive Import gets from
// ResetMigrations.
//
// The report is printed to stdout: the signature returns only an error, and a
// restore the user ran by hand is worth being told about in words.
func RestoreFromLegacy(r *repositories.Repos, path string) error {
	data, err := os.ReadFile(path)

	if err != nil {
		return fmt.Errorf("backup: reading %s: %w", path, err)
	}

	var docs []legacyDoc

	if err = json.Unmarshal(data, &docs); err != nil {
		return fmt.Errorf("backup: unmarshalling %s: %w", path, err)
	}

	if err = validateLegacy(docs); err != nil {
		return fmt.Errorf("backup: validating %s: %w", path, err)
	}

	state, err := readLegacyState(r)

	if err != nil {
		return err
	}

	var report legacyReport

	for _, doc := range docs {
		var docErr error

		switch {
		case strings.HasPrefix(doc.ID, legacyCommandPrefix):
			docErr = restoreLegacyCommand(r, doc, state, &report)
		case strings.HasPrefix(doc.ID, legacyManualPrefix):
			docErr = restoreLegacyManual(r, doc, state, &report)
		default:
			report.skipped++
		}

		if docErr != nil {
			return docErr
		}
	}

	report.print()

	return nil
}

// restoreLegacyCommand inserts the command named by the document, or reuses the
// one already there, and then appends the instructions it does not have yet.
//
// The command id handed to AppendCommandItem is the one InsertCommand assigned,
// or the one the existing row carries. It is never anything from the file: a
// legacy document carries no ids at all.
func restoreLegacyCommand(r *repositories.Repos, doc legacyDoc, state *legacyState, report *legacyReport) error {
	name := strings.TrimPrefix(doc.ID, legacyCommandPrefix)

	command, ok := state.commands[name]

	if !ok {
		inserted := &repositories.Command{
			Name:        name,
			Description: doc.Description,
		}

		if err := r.InsertCommand(inserted); err != nil {
			return fmt.Errorf("backup: inserting command %q: %w", name, err)
		}

		command = legacyCommand{id: inserted.ID, scripts: make(map[string]struct{})}
		state.commands[name] = command
		report.commands++
	}

	for _, script := range doc.Instructions {
		if _, ok := command.scripts[script]; ok {
			continue
		}

		// A script the database already holds under another name cannot be
		// inserted at all, since the constraint covers the whole table. Skipping
		// it and saying so beats failing the command over one line.
		if owner, ok := state.scriptOwner[script]; ok {
			report.note("item %q of command %q was skipped: the script already belongs to command %q",
				script, name, owner)
			continue
		}

		item := &repositories.CommandItem{Script: script}

		if err := r.AppendCommandItem(command.id, item); err != nil {
			return fmt.Errorf("backup: appending item %q of command %q: %w", script, name, err)
		}

		command.scripts[script] = struct{}{}
		state.scriptOwner[script] = name
		report.items++
	}

	return nil
}

// restoreLegacyManual inserts the manual named by the document, unless one of
// that name is already there.
func restoreLegacyManual(r *repositories.Repos, doc legacyDoc, state *legacyState, report *legacyReport) error {
	name := strings.TrimPrefix(doc.ID, legacyManualPrefix)

	if _, ok := state.manuals[name]; ok {
		report.note("manual %q was skipped: it already exists", name)
		return nil
	}

	manual := &repositories.Manual{
		Name:        name,
		Body:        doc.Content,
		Description: doc.Description,
	}

	if err := r.InsertManual(manual); err != nil {
		return fmt.Errorf("backup: inserting manual %q: %w", name, err)
	}

	state.manuals[name] = struct{}{}
	report.manuals++

	return nil
}

// validateLegacy rejects a file the repositories or the schema would refuse
// anyway, so the refusal happens before the first write rather than halfway
// through.
//
// It reports the first problem rather than collecting them: nothing has been
// touched at this point, so one clear error is enough. Documents with an
// unrecognised prefix are not an error at all — they are exactly what
// RestoreFromLegacy skips, and config:editor is the common case.
func validateLegacy(docs []legacyDoc) error {
	commands := make(map[string]struct{}, len(docs))
	scripts := make(map[string]string)
	manuals := make(map[string]struct{}, len(docs))

	for _, doc := range docs {
		switch {
		case strings.HasPrefix(doc.ID, legacyCommandPrefix):
			name := strings.TrimPrefix(doc.ID, legacyCommandPrefix)

			if name == "" {
				return fmt.Errorf("a command document has no name")
			}

			if _, ok := commands[name]; ok {
				return fmt.Errorf("command name %q appears more than once", name)
			}

			commands[name] = struct{}{}

			for _, script := range doc.Instructions {
				if script == "" {
					return fmt.Errorf("command %q has an instruction with no script", name)
				}

				if owner, ok := scripts[script]; ok {
					return fmt.Errorf(
						"script %q appears more than once, in %q and %q, and it is unique across all commands",
						script, owner, name,
					)
				}

				scripts[script] = name
			}

		case strings.HasPrefix(doc.ID, legacyManualPrefix):
			name := strings.TrimPrefix(doc.ID, legacyManualPrefix)

			if name == "" {
				return fmt.Errorf("a manual document has no name")
			}

			if doc.Content == "" {
				return fmt.Errorf("manual %q has no content", name)
			}

			if _, ok := manuals[name]; ok {
				return fmt.Errorf("manual name %q appears more than once", name)
			}

			manuals[name] = struct{}{}
		}
	}

	return nil
}

// readLegacyState reads what the database already holds, in one pass before
// anything is written.
//
// It runs one query per command for that command's items, the same N+1 Export
// runs. The connection is free again between reads, so this costs queries, not
// correctness, and it buys the convergence the whole function rests on.
func readLegacyState(r *repositories.Repos) (*legacyState, error) {
	commands, err := r.GetCommands()

	if err != nil {
		return nil, fmt.Errorf("backup: listing commands: %w", err)
	}

	manuals, err := r.GetManuals()

	if err != nil {
		return nil, fmt.Errorf("backup: listing manuals: %w", err)
	}

	state := &legacyState{
		commands:    make(map[string]legacyCommand, len(commands)),
		scriptOwner: make(map[string]string),
		manuals:     make(map[string]struct{}, len(manuals)),
	}

	for _, c := range commands {
		items, err := r.GetCommandItems(c.ID)

		if err != nil {
			return nil, fmt.Errorf("backup: listing items of command %d: %w", c.ID, err)
		}

		command := legacyCommand{id: c.ID, scripts: make(map[string]struct{}, len(items))}

		for _, item := range items {
			command.scripts[item.Script] = struct{}{}
			state.scriptOwner[item.Script] = c.Name
		}

		state.commands[c.Name] = command
	}

	for _, m := range manuals {
		state.manuals[m.Name] = struct{}{}
	}

	return state, nil
}

// note records something the user may want to know and counts it as skipped.
// It is for the cases that are not errors but are not silent successes either:
// an item lost to the global script constraint, or a manual already present.
func (rep *legacyReport) note(format string, args ...any) {
	rep.skipped++
	rep.notes = append(rep.notes, fmt.Sprintf(format, args...))
}

// print reports what the run did.
func (rep *legacyReport) print() {
	fmt.Printf("legacy: %d commands, %d items and %d manuals imported\n",
		rep.commands, rep.items, rep.manuals)

	if rep.skipped == 0 {
		return
	}

	fmt.Printf("legacy: %d skipped\n", rep.skipped)

	for _, note := range rep.notes {
		fmt.Printf("legacy: %s\n", note)
	}
}
