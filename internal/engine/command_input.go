package engine

import "strings"

type parsedCommand struct {
	name string
	args string
}

func parseCommandInput(input string) (parsedCommand, bool) {
	trimmed := strings.TrimSpace(input)
	parts := strings.Fields(trimmed)
	if len(parts) == 0 {
		return parsedCommand{}, false
	}

	name := parts[0]
	args := strings.TrimSpace(trimmed[len(name):])

	return parsedCommand{
		name: strings.ToUpper(name),
		args: args,
	}, true
}
