package engine

import (
	"fmt"
	"strings"

	"github.com/alex-ermolaxe/qdata/internal/completer"
	"github.com/alex-ermolaxe/qdata/internal/executor"
)

var transformCommand = CommandSpec{
	Name:        "TRANSFORM",
	Usage:       "TRANSFORM TO-MAP KEY <path> VALUE <path>",
	Description: "Transform records into maps using selected key and value paths.",
	Run:         runTransformCommand,
	Completion:  completeTransform,
	MCP: &MCPToolSpec{
		Name:        "transform",
		Description: "Transform each record into a one-key object. The KEY path supplies the dynamic property name; VALUE supplies the property value. Paths through arrays of objects extract the value from each element. Example: TO-MAP KEY sap_order_id VALUE assets.name.",
		Arguments: []MCPArgumentSpec{
			{
				Name:        "spec",
				Description: "Transform specification, for example TO-MAP KEY sap_order_id VALUE assets.name.",
				Required:    true,
			},
		},
	},
}

func runTransformCommand(e *Engine, args string) (CommandResult, error) {
	keyPath, valuePath, err := parseTransformArgs(args)
	if err != nil {
		return CommandResult{}, err
	}
	output, err := e.handleTransformToMap(keyPath, valuePath)
	if err != nil {
		return CommandResult{}, err
	}

	return CommandResult{Output: output}, nil
}

func completeTransform(context completer.CompletionContext) []string {
	arguments := context.Arguments
	switch {
	case len(arguments) == 0:
		return []string{"TO-MAP"}
	case len(arguments) == 1:
		if context.HasTrailingSpace {
			return []string{"KEY"}
		}
		return []string{"TO-MAP"}
	case len(arguments) == 2 && strings.EqualFold(arguments[1], "KEY"):
		return completeFieldPaths(context)
	case len(arguments) == 3:
		if context.HasTrailingSpace {
			return []string{"VALUE"}
		}
		return completeFieldPaths(context)
	case len(arguments) == 4 && strings.EqualFold(arguments[3], "VALUE"):
		return completeFieldPaths(context)
	case len(arguments) == 5 && strings.EqualFold(arguments[3], "VALUE"):
		return completeFieldPaths(context)
	default:
		return nil
	}
}

func parseTransformArgs(args string) (string, string, error) {
	parts := strings.Fields(args)
	if len(parts) != 5 ||
		!strings.EqualFold(parts[0], "TO-MAP") ||
		!strings.EqualFold(parts[1], "KEY") ||
		!strings.EqualFold(parts[3], "VALUE") ||
		parts[2] == "" ||
		parts[4] == "" {
		return "", "", fmt.Errorf("TRANSFORM syntax: TRANSFORM TO-MAP KEY <key_field> VALUE <value_field>")
	}

	return parts[2], parts[4], nil
}

func (e *Engine) handleTransformToMap(keyPath, valuePath string) (string, error) {
	transformed, err := executor.TransformToMap(e.session.Current, keyPath, valuePath)
	if err != nil {
		return "", fmt.Errorf("failed to transform records: %w", err)
	}

	e.session.Current = transformed
	e.session.UpdateSchema()
	recordCount := e.session.TotalRecords()
	output := fmt.Sprintf("✓ Transformed %d records to maps\n", recordCount)

	return output, nil
}
