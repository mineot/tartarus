package backup

import (
	"encoding/json"
	"os"
	"tartarus/repositories"
	"tartarus/store"
)

func Import(path string) error {
	var err error

	var dataFile []byte

	if dataFile, err = os.ReadFile(path); err != nil {
		return err
	}

	var file jsonFile

	if err = json.Unmarshal(dataFile, &file); err != nil {
		return err
	}

	st := store.Store{}

	if err := st.Open(); err != nil {
		return err
	}

	defer st.Close()

	if err := st.ResetMigrations(); err != nil {
		return err
	}

	for _, cmd := range file.Commands {
		var command = repositories.Command{
			Name:      cmd.Name,
			CreatedAt: cmd.CreatedAt,
			UpdatedAt: cmd.UpdatedAt,
		}

		if err := command.Insert(); err != nil {
			return err
		}

		for _, item := range cmd.Items {
			var commandItem = repositories.CommandItem{
				CommandID: cmd.ID,
				Script:    item.Script,
				CreatedAt: item.CreatedAt,
				UpdatedAt: item.UpdatedAt,
			}

			if err := command.AppendItem(commandItem); err != nil {
				return err
			}
		}
	}

	for _, man := range file.Manuals {
		var manual = repositories.Manual{
			Name:      man.Name,
			Body:      man.Body,
			CreatedAt: man.CreatedAt,
			UpdatedAt: man.UpdatedAt,
		}

		if err := manual.Insert(); err != nil {
			return err
		}
	}

	return nil
}
