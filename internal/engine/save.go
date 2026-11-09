package engine

import (
	"fmt"
	"strings"

	"github.com/alex-ermolaxe/qdata/internal/executor"
	"github.com/alex-ermolaxe/qdata/internal/format"
	"github.com/alex-ermolaxe/qdata/internal/schema"
)

var saveCommand = CommandSpec{
	Name:        "SAVE",
	Usage:       "SAVE [AS <filename>] [FORMAT <format>]",
	Description: "Save the current records.",
	Run:         runSaveCommand,
	Completion:  completeSave,
	MCP: &MCPToolSpec{
		Name:            "save",
		Description:     "Save the current records beside the source file. Options may include AS <filename> and FORMAT <format>. If the target exists, this tool returns an error instead of prompting; choose a different filename.",
		DestructiveHint: boolPointer(true),
		Arguments: []MCPArgumentSpec{
			{
				Name:        "options",
				Description: "Save options such as AS filtered_users FORMAT json. Use an empty string for the default output name and format.",
				Required:    true,
			},
		},
	},
}

var saveFormats = []string{"json", "xml", "csv"}

func runSaveCommand(e *Engine, args string) (CommandResult, error) {
	output, err := e.handleSave(args)
	if err != nil {
		return CommandResult{}, err
	}
	return CommandResult{Output: output}, nil
}

func completeSave(input string, _ *schema.Schema) []string {
	parts := strings.Fields(input)
	if len(parts) >= 2 && strings.EqualFold(parts[len(parts)-1], "FORMAT") {
		return saveFormats
	}
	if len(parts) >= 3 && strings.EqualFold(parts[len(parts)-2], "FORMAT") {
		return saveFormats
	}
	return nil
}

func (e *Engine) handleSave(args string) (string, error) {
	fileName := ""
	formatName := ""

	// Parse SAVE AS <filename> FORMAT <format> arguments
	parts := strings.Fields(args)
	for i := 0; i < len(parts); i++ {
		switch strings.ToUpper(parts[i]) {
		case "AS":
			if i+1 >= len(parts) {
				return "", fmt.Errorf("AS requires a filename")
			}
			fileName = parts[i+1]
			i++
		case "FORMAT":
			if i+1 >= len(parts) {
				return "", fmt.Errorf("FORMAT requires a format name")
			}
			formatName = parts[i+1]
			i++
		}
	}

	f := e.session.Format
	if formatName != "" {
		var err error
		f, err = format.Get(formatName)
		if err != nil {
			return "", fmt.Errorf("unknown format: %s", formatName)
		}
	}

	return executor.Save(e.session.Current, e.session.FilePath, fileName, f, e.saveConfirmation)
}
