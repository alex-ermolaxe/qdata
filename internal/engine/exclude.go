package engine

import (
	"fmt"
	"strings"

	"github.com/alex-ermolaxe/qdata/internal/executor"
	"github.com/alex-ermolaxe/qdata/internal/schema"
)

var excludeCommand = CommandSpec{
	Name:        "EXCLUDE",
	Usage:       "EXCLUDE <field>[, <field> ...]",
	Description: "Remove the specified fields from each record.",
	Run:         runExcludeCommand,
	Completion:  completeExclude,
	MCP: &MCPToolSpec{
		Name:        "exclude",
		Description: "Remove the specified comma-separated fields from each record. Nested fields can be addressed with dot paths.",
		Arguments: []MCPArgumentSpec{
			{
				Name:        "fields",
				Description: "Comma-separated field paths to remove.",
				Required:    true,
			},
		},
	},
}

func runExcludeCommand(e *Engine, args string) (CommandResult, error) {
	if strings.TrimSpace(args) == "" {
		return CommandResult{}, fmt.Errorf("EXCLUDE requires field names")
	}

	output, err := e.handleExclude(args)
	if err != nil {
		return CommandResult{}, err
	}
	return CommandResult{Output: output}, nil
}

func completeExclude(input string, currentSchema *schema.Schema) []string {
	return completeFieldPaths(input, currentSchema)
}

func (e *Engine) handleExclude(args string) (string, error) {
	fields := splitFields(args)

	e.session.Current = executor.Exclude(e.session.Current, fields)
	e.session.UpdateSchema()

	return fmt.Sprintf("✓ Applied EXCLUDE to %d records\n", e.session.TotalRecords()), nil
}
