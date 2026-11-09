package engine

import (
	"fmt"
	"strings"

	"github.com/alex-ermolaxe/qdata/internal/executor"
	"github.com/alex-ermolaxe/qdata/internal/schema"
)

var sortCommand = CommandSpec{
	Name:        "SORT",
	Usage:       "SORT <field> [ASC|DESC]",
	Description: "Sort records by a field.",
	Run:         runSortCommand,
	Completion:  completeSort,
	MCP: &MCPToolSpec{
		Name:        "sort",
		Description: "Sort the current records by a field path. The optional direction is ASC or DESC; it defaults to ASC.",
		Arguments: []MCPArgumentSpec{
			{
				Name:        "spec",
				Description: "Field path and optional direction, for example created_at DESC.",
				Required:    true,
			},
		},
	},
}

var sortDirections = []string{"ASC", "DESC"}

func runSortCommand(e *Engine, args string) (CommandResult, error) {
	if strings.TrimSpace(args) == "" {
		return CommandResult{}, fmt.Errorf("SORT requires a field name")
	}

	output, err := e.handleSort(args)
	if err != nil {
		return CommandResult{}, err
	}
	return CommandResult{Output: output}, nil
}

func completeSort(input string, currentSchema *schema.Schema) []string {
	upperInput := strings.ToUpper(input)
	parts := strings.Fields(strings.TrimLeft(upperInput, " "))

	if len(parts) >= 2 && strings.HasSuffix(input, " ") {
		return sortDirections
	}
	if len(parts) == 3 {
		return sortDirections
	}
	return completeFieldPaths(input, currentSchema)
}

func (e *Engine) handleSort(args string) (string, error) {
	parts := strings.Fields(args)

	field := parts[0]
	direction := executor.SortAsc

	if len(parts) > 1 {
		switch strings.ToUpper(parts[1]) {
		case "ASC":
			direction = executor.SortAsc
		case "DESC":
			direction = executor.SortDesc
		default:
			return "", fmt.Errorf("invalid sort direction: %s, expected ASC or DESC", parts[1])
		}
	}

	sorted, err := executor.Sort(e.session.Current, field, direction)
	if err != nil {
		return "", err
	}

	e.session.Current = sorted

	return fmt.Sprintf("✓ Sorted %d records by %q %s\n", e.session.TotalRecords(), field, direction), nil
}
