package main

import (
	"log"

	"github.com/netlify/gotell/cmd"
)

func main() {
	if err := cmd.RootCommand().Execute(); err != nil {
		log.Fatal(err)
	}
}
