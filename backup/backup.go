package backup

import (
	"encoding/json"
	"os"
	"tartarus/repositories"
	"time"
)

type commandItem struct {
	ID        uint64    `json:"id"`
	CommandID uint64    `json:"command_id"`
	Script    string    `json:"script"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type command struct {
	ID        uint64        `json:"id"`
	Name      string        `json:"name"`
	Items     []commandItem `json:"items"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

type manual struct {
	ID        uint64    `json:"id"`
	Name      string    `json:"name"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type jsonFile struct {
	Commands []command `json:"commands"`
	Manuals  []manual  `json:"manuals"`
}

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

	// TODO implementar a gravação dos dados de file dentro do banco de dados, o caminho reverso do Export, deve-se verificar o dado já existe antes de tentar gravar

	return nil
}

func ImportFromLegacy() {}
