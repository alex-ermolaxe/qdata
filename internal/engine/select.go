package engine

import (
	"fmt"
	"strings"

	"github.com/alex-ermolaxe/qdata/internal/executor"
)

var selectCommand = CommandSpec{
	Name:        "SELECT",
	Usage:       "SELECT <field>[, <field> ...] | *",
	Description: "Keep only the specified fields in each record.",
	Run:         runSelectCommand,
	Completion:  completeFieldPaths,
	MCP: &MCPToolSpec{
		Name:        "select",
		Description: "Keep only the specified comma-separated fields in each record. Supports nested dot paths and fields inside arrays of objects, such as items.name. Use * to retain all fields.",
		Arguments: []MCPArgumentSpec{
			{
				Name:        "fields",
				Description: "Comma-separated field paths, for example id, name, items.name.",
				Required:    true,
			},
		},
	},
}

func runSelectCommand(e *Engine, args string) (CommandResult, error) {
	if strings.TrimSpace(args) == "" {
		return CommandResult{}, fmt.Errorf("SELECT requires field names")
	}

	output, err := e.handleSelect(args)
	if err != nil {
		return CommandResult{}, err
	}

	return CommandResult{Output: output}, nil
}

func (e *Engine) handleSelect(args string) (string, error) {
	fields := splitFields(args)
	was := e.session.TotalRecords()

	e.session.Current = executor.Select(e.session.Current, fields)
	e.session.UpdateSchema()

	recordCount := e.session.TotalRecords()
	output := fmt.Sprintf("✓ Applied SELECT to %d records (was: %d)\n", recordCount, was)

	return output, nil
}
