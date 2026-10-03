package main

import (
	"fmt"
	"tartarus/backup"
	"tartarus/helpers"
)

func main() {
	fmt.Println("Welcome to Tartarus CLI")

	path, err := helpers.GetDevelopmentStorePath("test.json")

	if err != nil {
		panic(err)
	}

	backup.Import(path)
}
