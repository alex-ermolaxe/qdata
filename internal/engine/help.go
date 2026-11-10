package engine

import (
	"fmt"
	"strings"
)

func helpCommandSpec() CommandSpec {
	return CommandSpec{
		Name:        "HELP",
		Usage:       "HELP",
		Description: "Show this command list.",
		Run:         runHelpCommand,
		MCP: &MCPToolSpec{
			Name:               "help",
			Description:        "List available qdata commands and their usage.",
			SessionIndependent: true,
		},
	}
}

func runHelpCommand(_ *Engine, _ string) (CommandResult, error) {
	return CommandResult{Output: renderHelp()}, nil
}

func renderHelp() string {
	commands := CommandSpecs()
	maxUsageWidth := 0
	for _, command := range commands {
		if len(command.Usage) > maxUsageWidth {
			maxUsageWidth = len(command.Usage)
		}
	}

	var output strings.Builder
	output.WriteString("Available commands:\n")
	for _, command := range commands {
		fmt.Fprintf(&output, "  %-*s  %s\n", maxUsageWidth, command.Usage, command.Description)
	}

	return output.String()
}
