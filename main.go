package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"tartarus/repositories"
	"tartarus/store"
	"tartarus/views"
)

func run() (err error) {
	s, err := store.New(context.Background())

	if err != nil {
		return err
	}

	defer func() {
		err = errors.Join(
			err,
			s.Close(),
		)
	}()

	if err := s.RunMigrations(); err != nil {
		return err
	}

	// r := repositories.New(s)
	// for i := range 50 {
	// 	cmd := repositories.Command{
	// 		Name:        "Comando " + fmt.Sprint(i),
	// 		Description: "Descrição do Commando: " + fmt.Sprint(i),
	// 	}
	// 	r.InsertCommand(&cmd)
	// 	for j := range 50 {
	// 		r.AppendCommandItem(cmd.ID, &repositories.CommandItem{
	// 			Script:      "Script " + fmt.Sprint(i) + "." + fmt.Sprint(j),
	// 			Description: "Descrição do Script: " + fmt.Sprint(i) + "." + fmt.Sprint(j),
	// 		})
	// 	}
	// }

	return views.RunView(repositories.New(s))
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Tartarus failed: %v\n", err)
		os.Exit(1)
	}
}
