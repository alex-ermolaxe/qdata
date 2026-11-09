package engine

import (
	"fmt"
	"strings"
)

func infoCommandSpec() CommandSpec {
	return CommandSpec{
		Name:        "INFO",
		Usage:       "INFO",
		Description: "Show this command list.",
		Run:         runInfoCommand,
	}
}

func runInfoCommand(_ *Engine, _ string) (CommandResult, error) {
	return CommandResult{Output: renderInfo()}, nil
}

func renderInfo() string {
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
