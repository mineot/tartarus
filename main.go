package main

import (
	"context"
	"fmt"
	"log"

	"tartarus/store"
)

func main() {
	fmt.Println("Welcome to Tartarus CLI")

	s, err := store.New(context.Background())

	if err != nil {
		log.Fatal(err)
	}

	defer s.Close()

	if err = s.RunMigrations(); err != nil {
		log.Fatal(err)
	}

	fmt.Println(s.Path())
}
