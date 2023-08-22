package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	args := os.Args

	if len(args) != 2 || (strings.ToLower(args[1]) != "backup" && strings.ToLower(args[1]) != "restore") {
		fmt.Printf("USAGE: %s [backup|restore]\n", args[0])
		os.Exit(1)
	}

	if strings.ToLower(args[1]) == "backup" {
		fmt.Println("Backing up...")
		Backup()
	} else {
		fmt.Println("Restoring...")
		Restore()
	}
}
