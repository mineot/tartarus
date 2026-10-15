package main

import (
	"context"
	"fmt"
	"os"
	"tartarus/repositories"
	"tartarus/store"
	"tartarus/themes"
	"tartarus/views"
)

func main() {
	if err := run(); err != nil {
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

	// for i := 1; i <= 10; i++ {
	// 	cmd := repositories.Command{
	// 		Name:        fmt.Sprintf("command%d", i),
	// 		Description: fmt.Sprintf("Description command %d", i),
	// 	}

	// 	if err := r.InsertCommand(&cmd); err != nil {
	// 		return fmt.Errorf("fail to create command: %w", err)
	// 	}

	// 	for j := 1; j <= 10; j++ {
	// 		script := repositories.CommandScript{
	// 			Script:      fmt.Sprintf("echo %d.%d", i, j),
	// 			Description: fmt.Sprintf("Description command %d script %d", i, j),
	// 		}

	// 		if err := r.AppendCommandScript(cmd.ID, &script); err != nil {
	// 			return fmt.Errorf("fail to create command script: %w", err)
	// 		}
	// 	}
	// }

	themes.RenderLogo()

	return views.New(r).ExecuteContext(ctx)
}
