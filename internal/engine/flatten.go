package engine

import (
	"fmt"
	"strings"

	"github.com/alex-ermolaxe/qdata/internal/executor"
	"github.com/alex-ermolaxe/qdata/internal/schema"
)

var flattenCommand = CommandSpec{
	Name:        "FLATTEN",
	Usage:       "FLATTEN <object_field>",
	Description: "Flatten a nested object field.",
	Run:         runFlattenCommand,
	Completion:  completeFlatten,
	MCP: &MCPToolSpec{
		Name:        "flatten",
		Description: "Flatten a nested object field into its parent record. If any nested field collides with an existing top-level field, all fields from the object are prefixed with the object's name.",
		Arguments: []MCPArgumentSpec{
			{
				Name:        "field",
				Description: "Name of the object field to flatten.",
				Required:    true,
			},
		},
	},
}

func runFlattenCommand(e *Engine, args string) (CommandResult, error) {
	if strings.TrimSpace(args) == "" {
		return CommandResult{}, fmt.Errorf("FLATTEN requires a field name")
	}

	output, err := e.handleFlatten(args)
	if err != nil {
		return CommandResult{}, err
	}
	return CommandResult{Output: output}, nil
}

func completeFlatten(input string, currentSchema *schema.Schema) []string {
	return completeFieldPaths(input, currentSchema)
}

func (e *Engine) handleFlatten(args string) (string, error) {
	fieldName := strings.TrimSpace(args)
	was := e.session.TotalRecords()

	flattened, err := executor.Flatten(e.session.Current, fieldName)
	if err != nil {
		return "", fmt.Errorf("failed to flatten field %q: %w", fieldName, err)
	}

	e.session.Current = flattened
	e.session.UpdateSchema()

	return fmt.Sprintf("✓ Flattened field %q in %d records (was: %d)\n", fieldName, e.session.TotalRecords(), was), nil
}
