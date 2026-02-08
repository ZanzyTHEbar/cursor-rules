package main

import (
	"github.com/ZanzyTHEbar/cursor-rules/internal/cli"
	"github.com/ZanzyTHEbar/cursor-rules/internal/cli/commands"
)

func main() {
	commands.RegisterAll()
	cli.Execute()
}
