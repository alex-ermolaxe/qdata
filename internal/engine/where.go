package engine

import (
	"fmt"
	"strings"

	"github.com/alex-ermolaxe/qdata/internal/executor"
	"github.com/alex-ermolaxe/qdata/internal/parser"
)

var whereCommand = CommandSpec{
	Name:        "WHERE",
	Usage:       "WHERE <condition> [AND|OR <condition> ...]",
	Description: "Filter the current record set by a condition.",
	Run:         runWhereCommand,
	Completion:  completeFieldPaths,
	MCP: &MCPToolSpec{
		Name:        "where",
		Description: `Filter records in a session. Syntax: <field> <operator> <value>, optionally combined with AND or OR. Operators: =, !=, >, <, >=, <=, ~, !~, ^, $, IN, EXISTS. Nested fields use dot paths (for example address.city). Example: status = "active" AND age >= 18.`,
		Arguments: []MCPArgumentSpec{
			{
				Name:        "condition",
				Description: "Filter condition in qdata syntax; nested field paths use dots.",
				Required:    true,
			},
		},
	},
}

func runWhereCommand(e *Engine, args string) (CommandResult, error) {
	if strings.TrimSpace(args) == "" {
		return CommandResult{}, fmt.Errorf("WHERE requires a condition")
	}

	output, err := e.handleWhere(args)
	if err != nil {
		return CommandResult{}, err
	}

	return CommandResult{Output: output}, nil
}

func (e *Engine) handleWhere(args string) (string, error) {
	group, err := parser.Parse(args)
	if err != nil {
		return "", fmt.Errorf("failed to parse condition: %w", err)
	}

	was := e.session.TotalRecords()

	filtered, err := executor.Filter(e.session.Current, group)
	if err != nil {
		return "", fmt.Errorf("failed to apply filter: %w", err)
	}

	e.session.Current = filtered
	recordCount := e.session.TotalRecords()
	output := fmt.Sprintf("✓ Found: %d records (was: %d)\n", recordCount, was)

	return output, nil
}
