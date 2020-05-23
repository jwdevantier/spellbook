package main

import (
	"github.com/jwdevantier/spellbook/cmd"
	"os"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}