package engine

var schemaCommand = CommandSpec{
	Name:        "SCHEMA",
	Usage:       "SCHEMA",
	Description: "Show fields and their inferred types.",
	Run: func(e *Engine, _ string) (CommandResult, error) {
		return CommandResult{Output: e.session.Schema.Render()}, nil
	},
	MCP: &MCPToolSpec{
		Name:        "schema",
		Description: "Show the fields and inferred types available in the session's current result. Use this before constructing field paths and filter conditions.",
	},
}
