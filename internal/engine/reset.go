package engine

import "fmt"

var resetCommand = CommandSpec{
	Name:        "RESET",
	Usage:       "RESET",
	Description: "Restore the original records.",
	Run:         runResetCommand,
	MCP: &MCPToolSpec{
		Name:        "reset",
		Description: "Restore the session's current records to the original loaded dataset.",
	},
}

func runResetCommand(e *Engine, _ string) (CommandResult, error) {
	e.session.Reset()
	recordCount := e.session.TotalRecords()
	output := fmt.Sprintf("✓ Reset to original: %d records\n", recordCount)

	return CommandResult{Output: output}, nil
}
