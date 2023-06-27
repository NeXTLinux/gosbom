package main

import (
	"log"

	"github.com/nextlinux/gosbom/cmd/gosbom/cli"
	_ "modernc.org/sqlite"
)

func main() {
	cli, err := cli.New()
	if err != nil {
		log.Fatalf("error during command construction: %v", err)
	}

	if err := cli.Execute(); err != nil {
		log.Fatalf("error during command execution: %v", err)
	}
}
