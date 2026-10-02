package main

import (
	"fmt"
	"tartarus/repositories"
)

func main() {
	fmt.Println("Welcome to Tartarus CLI")

	cmd := repositories.Command{}
	cmd.Name = "Cmd 1"

	if err := cmd.Insert(); err != nil {
		fmt.Println(err)
	}
}
