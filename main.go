package main

import (
	"github.com/vampire/syn/cmd"
	"github.com/vampire/syn/internal/config"
)

func main() {
	config.SetDefaults()
	cmd.Execute()
}
