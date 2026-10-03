package backup

import (
	"encoding/json"
	"os"
	"tartarus/repositories"
)

func Export(path string) error {
	var err error

	var file jsonFile = jsonFile{
		Commands: []command{},
		Manuals:  []manual{},
	}

	var cmds []repositories.Command

	if cmds, err = repositories.GetCommands(); err != nil {
		return err
	}

	for index, cmd := range cmds {
		file.Commands = append(file.Commands, command{
			ID:        cmd.ID,
			Name:      cmd.Name,
			Items:     []commandItem{},
			CreatedAt: cmd.CreatedAt,
			UpdatedAt: cmd.UpdatedAt,
		})

		var cmdItems []repositories.CommandItem

		if cmdItems, err = cmd.GetItems(); err != nil {
			return err
		}

		for _, item := range cmdItems {
			file.Commands[index].Items = append(file.Commands[index].Items, commandItem{
				ID:        item.ID,
				CommandID: item.CommandID,
				Script:    item.Script,
				CreatedAt: item.CreatedAt,
				UpdatedAt: item.UpdatedAt,
			})
		}
	}

	var mans []repositories.Manual

	if mans, err = repositories.GetManuals(); err != nil {
		return err
	}

	for _, man := range mans {
		file.Manuals = append(file.Manuals, manual{
			ID:        man.ID,
			Name:      man.Name,
			Body:      man.Body,
			CreatedAt: man.CreatedAt,
			UpdatedAt: man.UpdatedAt,
		})
	}

	var dataFile []byte

	if dataFile, err = json.MarshalIndent(file, "", "  "); err != nil {
		return err
	}

	if err = os.WriteFile(path, dataFile, 0644); err != nil {
		return err
	}

	return nil
}
