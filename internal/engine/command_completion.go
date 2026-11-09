package engine

import (
	"strings"

	"github.com/alex-ermolaxe/qdata/internal/schema"
)

func completeFieldPaths(input string, currentSchema *schema.Schema) []string {
	if currentSchema == nil {
		return nil
	}

	trimmed := strings.TrimLeft(input, " ")
	parts := strings.Fields(trimmed)
	prefix := ""
	if !strings.HasSuffix(input, " ") && len(parts) > 1 {
		prefix = parts[len(parts)-1]
	}

	allPaths := currentSchema.AllPaths()
	if strings.Contains(prefix, ".") {
		return allPaths
	}

	topLevel := make([]string, 0, len(allPaths))
	for _, path := range allPaths {
		if !strings.Contains(path, ".") {
			topLevel = append(topLevel, path)
		}
	}
	return topLevel
}
