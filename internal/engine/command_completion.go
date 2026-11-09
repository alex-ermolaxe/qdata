package engine

import (
	"strings"

	"github.com/alex-ermolaxe/qdata/internal/completer"
)

func completeFieldPaths(context completer.CompletionContext) []string {
	currentSchema := context.Schema
	if currentSchema == nil {
		return nil
	}

	allPaths := currentSchema.AllPaths()
	if strings.Contains(context.CurrentWord, ".") {
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
