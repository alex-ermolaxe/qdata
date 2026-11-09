package engine

import "fmt"

var countCommand = CommandSpec{
	Name:        "COUNT",
	Usage:       "COUNT",
	Description: "Count records in the current result.",
	Run:         runCountCommand,
	MCP: &MCPToolSpec{
		Name:        "count",
		Description: "Return the number of records in the session's current result.",
	},
}

func runCountCommand(e *Engine, _ string) (CommandResult, error) {
	return CommandResult{Output: fmt.Sprintf("%d records\n", e.session.TotalRecords())}, nil
}
