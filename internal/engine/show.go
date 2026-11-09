package engine

import (
	"fmt"
	"strings"

	"github.com/alex-ermolaxe/qdata/internal/completer"
	"github.com/alex-ermolaxe/qdata/internal/executor"
)

var showCommand = CommandSpec{
	Name:        "SHOW",
	Usage:       "SHOW [LIMIT <n>] [OFFSET <n>]",
	Description: "Display the current records.",
	Run:         runShowCommand,
	Completion:  completeShow,
	MCP: &MCPToolSpec{
		Name:        "show",
		Description: "Display the current records as formatted JSON. Options may include LIMIT <n> and OFFSET <n>; provide an empty options string to show all records.",
		Arguments: []MCPArgumentSpec{
			{
				Name:        "options",
				Description: "Optional display options such as LIMIT 20 OFFSET 10; use an empty string for all records.",
				Required:    true,
			},
		},
	},
}

func runShowCommand(e *Engine, args string) (CommandResult, error) {
	output, err := e.handleShow(args)
	if err != nil {
		return CommandResult{}, err
	}

	return CommandResult{Output: output}, nil
}

func completeShow(context completer.CompletionContext) []string {
	if len(context.Arguments) >= 1 && context.HasTrailingSpace {
		lastArgument := context.Arguments[len(context.Arguments)-1]
		if strings.EqualFold(lastArgument, "LIMIT") || strings.EqualFold(lastArgument, "OFFSET") {
			return nil
		}
	}

	return []string{"LIMIT", "OFFSET"}
}

func (e *Engine) handleShow(args string) (string, error) {
	limit := 0
	offset := 0

	// Parse SHOW LIMIT 10 OFFSET 20 arguments
	parts := strings.Fields(strings.ToUpper(args))
	for i := 0; i < len(parts); i++ {
		switch parts[i] {
		case "LIMIT":
			if i+1 >= len(parts) {
				return "", fmt.Errorf("LIMIT requires a number")
			}
			if _, err := fmt.Sscanf(parts[i+1], "%d", &limit); err != nil {
				return "", fmt.Errorf("invalid LIMIT value: %s", parts[i+1])
			}
			i++
		case "OFFSET":
			if i+1 >= len(parts) {
				return "", fmt.Errorf("OFFSET requires a number")
			}
			if _, err := fmt.Sscanf(parts[i+1], "%d", &offset); err != nil {
				return "", fmt.Errorf("invalid OFFSET value: %s", parts[i+1])
			}
			i++
		}
	}

	return executor.RenderShow(e.session.Current, limit, offset)
}
