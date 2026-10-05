package main

import (
	"context"
	"fmt"
	"log"

	"tartarus/backup"
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

	backup.Export(r, "./teste2.json")
}
