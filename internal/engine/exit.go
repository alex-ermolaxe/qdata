package engine

import (
	"fmt"
	"strings"
)

var exitCommand = CommandSpec{
	Name:        "EXIT",
	Usage:       "EXIT",
	Description: "Quit qdata.",
	Run:         runExitCommand,
}

func runExitCommand(_ *Engine, args string) (CommandResult, error) {
	if strings.TrimSpace(args) != "" {
		return CommandResult{}, fmt.Errorf("EXIT does not accept arguments")
	}
	return CommandResult{Output: "Bye!\n", Exit: true}, nil
}
