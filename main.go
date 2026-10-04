package main

import (
	"context"
	"fmt"
	"log"

	"tartarus/repositories"
	"tartarus/store"
)

func main() {
	fmt.Println("Welcome to Tartarus CLI")

	// One Store for the whole process. The repositories take it as an argument
	// instead of opening a connection per call, so this is the only place that
	// decides which database is used.
	s, err := store.New(context.Background())

	if err != nil {
		log.Fatal(err)
	}

	defer s.Close()

	if err = s.RunMigrations(); err != nil {
		log.Fatal(err)
	}

	fmt.Println(s.Path())

	r := repositories.New(s)

	listManuals(r)
}

// listManuals is an example of reading through a repository: a plain read, no
// transaction, straight after the migrations are done.
func listManuals(r *repositories.Repos) {
	manuals, err := r.GetManuals()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\n%d manual(s)\n", len(manuals))

	for _, m := range manuals {
		fmt.Printf("  %d\t%s\n", m.ID, m.Name)
	}
}
