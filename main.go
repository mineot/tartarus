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

	fmt.Println(s.Path())
}
