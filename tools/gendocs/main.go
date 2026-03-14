package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/0xveya/gns3util/internal/cli"
)

func main() {
	path := filepath.Join("documentation", "cli")
	err := os.MkdirAll(path, 0o750)
	if err != nil {
		log.Fatal(err)
	}
	cli.GenrateDocs(path)

	log.Println("CLI docs generated in ", path)
}
