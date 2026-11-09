package engine

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/alex-ermolaxe/qdata/internal/completer"
	"github.com/alex-ermolaxe/qdata/internal/executor"
	"github.com/alex-ermolaxe/qdata/internal/format"
	"github.com/alex-ermolaxe/qdata/internal/schema"
	"github.com/alex-ermolaxe/qdata/internal/version"
	"github.com/chzyer/readline"
)

// Engine - main application engine
type Engine struct {
	session          *Session
	saveConfirmation func(string) bool
}

// New creates a new Engine, loads a file and builds a session
func New(filePath string, formatName string) (*Engine, error) {
	return NewFromFile(filePath, formatName)
}

// NewFromFile creates an Engine from a file path and optional format name.
func NewFromFile(filePath string, formatName string) (*Engine, error) {
	// Determine format
	var f format.Format
	var err error

	if formatName != "" {
		f, err = format.Get(formatName)
	} else {
		f, err = format.Detect(filePath)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to determine format: %w", err)
	}

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Decode data
	records, err := f.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode file: %w", err)
	}

	// Create session
	session := NewSession(filePath, f, records)

	return &Engine{session: session}, nil
}

// RecordCount returns the number of records in the current result.
func (e *Engine) RecordCount() int {
	return e.session.TotalRecords()
}

// FilePath returns the source file path for this engine.
func (e *Engine) FilePath() string {
	return e.session.FilePath
}

// Run starts an interactive REPL session
func (e *Engine) Run() error {
	e.saveConfirmation = executor.ConfirmOverwrite
	specs := CommandSpecs()
	completionCommands := make([]completer.CommandCompletion, 0, len(specs))
	for _, spec := range specs {
		completionCommands = append(completionCommands, completer.CommandCompletion{
			Name:     spec.Name,
			Complete: spec.Completion,
		})
	}

	c := completer.New(
		func() *schema.Schema {
			return e.session.Schema
		},
		completionCommands,
	)
	config := completer.NewReadline(c)

	// Initialize readline
	rl, err := readline.NewEx(config)
	if err != nil {
		return fmt.Errorf("failed to initialize readline: %w", err)
	}
	defer rl.Close()

	// Welcome message
	fileName := filepath.Base(e.session.FilePath)
	fmt.Printf("qdata v%s | file: %s | records: %d\n\n",
		version.Version,
		fileName,
		e.session.OriginalRecords(),
	)

	// REPL loop
	for {
		line, err := rl.Readline()
		if err != nil {
			// Ctrl+D or Ctrl+C
			if err == readline.ErrInterrupt || err == io.EOF {
				fmt.Println("\nBye!")
				return nil
			}
			return err
		}

		input := strings.TrimSpace(line)
		if input == "" {
			continue
		}

		// Process command
		if e.handleCommand(input) {
			return nil
		}
	}
}

// handleCommand processes an interactive command.
func (e *Engine) handleCommand(input string) bool {
	result, err := e.ExecuteCommand(input)
	if err != nil {
		fmt.Printf("error: %s\n", err)
		return false
	}
	fmt.Print(result.Output)
	return result.Exit
}

// ExecuteCommand runs a command without writing command output to stdout.
func (e *Engine) ExecuteCommand(input string) (CommandResult, error) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return CommandResult{}, nil
	}

	command := strings.ToUpper(parts[0])
	args := strings.TrimSpace(input[len(parts[0]):])

	for _, spec := range CommandSpecs() {
		if spec.Name != command {
			continue
		}
		if spec.Run == nil {
			return CommandResult{}, fmt.Errorf("command %s has no handler", command)
		}
		return spec.Run(e, args)
	}
	return CommandResult{}, fmt.Errorf("unknown command: %s", command)
}

// splitFields splits field string by comma
func splitFields(args string) []string {
	parts := strings.Split(args, ",")
	fields := make([]string, 0, len(parts))
	for _, p := range parts {
		field := strings.TrimSpace(p)
		if field != "" {
			fields = append(fields, field)
		}
	}
	return fields
}
