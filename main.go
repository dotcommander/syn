package main

import (
	"github.com/dotcommander/syn/cmd"
	"github.com/dotcommander/syn/internal/config"
)

func main() {
	config.SetDefaults()
	cmd.Execute()
}
