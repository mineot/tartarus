package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"tartarus/repositories"
	"tartarus/store"
	"tartarus/themes"
	"tartarus/views"
)

func main() {
	context := context.Background()

	s, err := store.New(context)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Tartarus failed initialize store: %v\n", err)
	}

	defer func() {
		err = errors.Join(
			err,
			s.Close(),
		)
	}()

	if err := s.RunMigrations(); err != nil {
		fmt.Fprintf(os.Stderr, "Tartarus failed run migrations: %v\n", err)
	}

	r := repositories.New(s)

	themes.RenderLogo()

	if err := views.New(r).ExecuteContext(context); err != nil {
		fmt.Fprintf(os.Stderr, "Tartarus failed execute views: %v\n", err)
		os.Exit(1)
	}
}
