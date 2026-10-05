package backup

import (
	"encoding/json"
	"fmt"
	"os"

	"tartarus/repositories"
)

// exportPerm is the permission bits of the exported file.
const exportPerm = 0644

// Export writes every command, command item and manual to path as JSON.
//
// It takes the repositories rather than opening anything: the store is opened
// once per process and handed in, so the file that gets exported is the file the
// caller migrated. Export never migrates either, for the same reason store.New
// does not.
//
// It reads through the repositories one read at a time and never starts a
// transaction, so it must not be called from inside a store.WithTx callback:
// those reads go through store.Store.Query, which returns store.ErrUseTx while a
// transaction is in progress.
func Export(r *repositories.Repos, path string) error {
	file := jsonFile{
		Commands: []command{},
		Manuals:  []manual{},
	}

	commands, err := r.GetCommands()

	if err != nil {
		return fmt.Errorf("backup: listing commands: %w", err)
	}

	for _, c := range commands {
		items, err := r.GetCommandItems(c.ID)

		if err != nil {
			return fmt.Errorf("backup: listing items of command %d: %w", c.ID, err)
		}

		entry := command{
			ID:          c.ID,
			Name:        c.Name,
			Description: c.Description,
			Items:       []commandItem{},
			CreatedAt:   c.CreatedAt,
			UpdatedAt:   c.UpdatedAt,
		}

		for _, item := range items {
			entry.Items = append(entry.Items, commandItem{
				ID:          item.ID,
				CommandID:   item.CommandID,
				Script:      item.Script,
				Description: item.Description,
				CreatedAt:   item.CreatedAt,
				UpdatedAt:   item.UpdatedAt,
			})
		}

		file.Commands = append(file.Commands, entry)
	}

	manuals, err := r.GetManuals()

	if err != nil {
		return fmt.Errorf("backup: listing manuals: %w", err)
	}

	for _, m := range manuals {
		file.Manuals = append(file.Manuals, manual{
			ID:          m.ID,
			Name:        m.Name,
			Body:        m.Body,
			Description: m.Description,
			CreatedAt:   m.CreatedAt,
			UpdatedAt:   m.UpdatedAt,
		})
	}

	data, err := json.MarshalIndent(file, "", "  ")

	if err != nil {
		return fmt.Errorf("backup: encoding json: %w", err)
	}

	if err = os.WriteFile(path, data, exportPerm); err != nil {
		return fmt.Errorf("backup: writing %s: %w", path, err)
	}

	return nil
}
