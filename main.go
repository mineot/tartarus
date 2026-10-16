package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"tartarus/components"
	"tartarus/repositories"
	"tartarus/store"
	"tartarus/themes"
	"tartarus/views"
)

func main() {
	if err := run(); err != nil {
		if errors.Is(err, components.ErrInterrupted) {
			fmt.Fprintln(os.Stderr)
			os.Exit(130)
		}

		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	s, err := store.New(ctx)

	if err != nil {
		return fmt.Errorf("Tartarus failed initialize store: %w", err)
	}

	defer s.Close()

	if err := s.RunMigrations(); err != nil {
		return fmt.Errorf("Tartarus failed run migrations: %w", err)
	}

	r := repositories.New(s)

	themes.RenderLogo()

	return views.New(r).ExecuteContext(ctx)
}
