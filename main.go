package main

import (
	"fmt"
	"log"
	"tartarus/store"
)

func main() {
	fmt.Println("Welcome to Tartarus CLI")

	s := store.Store{}

	if err := s.Open(); err != nil {
		log.Fatalf("unable to open database: %v", err)
	}

	if err := s.RunMigrations(); err != nil {
		log.Fatalf("unable run migrations: %v", err)
	}

	defer s.Close()
}
