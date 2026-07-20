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
	"github.com/alex-ermolaxe/qdata/internal/parser"
	"github.com/alex-ermolaxe/qdata/internal/schema"
	"github.com/alex-ermolaxe/qdata/internal/version"
	"github.com/chzyer/readline"
)

// Engine - main application engine
type Engine struct {
	session *Session
}

// New creates a new Engine, loads a file and builds a session
func New(filePath string, formatName string) (*Engine, error) {
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

// Run starts an interactive REPL session
func (e *Engine) Run() error {
	// Create autocomplete with a function that returns the current schema
	c := completer.New(func() *schema.Schema {
		return e.session.Schema
	})
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
		if err := e.handleCommand(input); err != nil {
			fmt.Printf("error: %s\n", err)
		}
	}
}

// handleCommand processes the entered command
func (e *Engine) handleCommand(input string) error {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil
	}

	command := strings.ToUpper(parts[0])
	args := strings.TrimSpace(input[len(parts[0]):])

	switch command {
	case "EXIT":
		fmt.Println("Bye!")
		os.Exit(0)
	case "SCHEMA":
		e.session.Schema.Print()
	case "COUNT":
		fmt.Printf("%d records\n", e.session.TotalRecords())
	case "RESET":
		e.session.Reset()
		fmt.Printf("✓ Reset to original: %d records\n", e.session.TotalRecords())
	case "WHERE":
		if args == "" {
			return fmt.Errorf("WHERE requires a condition")
		}
		return e.handleWhere(args)
	case "SELECT":
		if args == "" {
			return fmt.Errorf("SELECT requires field names")
		}
		return e.handleSelect(args)
	case "SHOW":
		return e.handleShow(args)
	case "SAVE":
		return e.handleSave(args)
	case "EXCLUDE":
		if args == "" {
			return fmt.Errorf("EXCLUDE requires field names")
		}
		return e.handleExclude(args)
	case "SORT":
		if args == "" {
			return fmt.Errorf("SORT requires a field name")
		}
		return e.handleSort(args)
	case "FLATTEN":
		if args == "" {
			return fmt.Errorf("FLATTEN requires a field name")
		}
		return e.handleFlatten(args)
	default:
		fmt.Printf("unknown command: %s\n", command)
	}

	return nil
}

func (e *Engine) handleWhere(args string) error {
	group, err := parser.Parse(args)
	if err != nil {
		return fmt.Errorf("failed to parse condition: %w", err)
	}

	was := e.session.TotalRecords()

	filtered, err := executor.Filter(e.session.Current, group)
	if err != nil {
		return fmt.Errorf("failed to apply filter: %w", err)
	}

	e.session.Current = filtered
	fmt.Printf("✓ Found: %d records (was: %d)\n", e.session.TotalRecords(), was)

	return nil
}

func (e *Engine) handleSelect(args string) error {
	fields := splitFields(args)
	was := e.session.TotalRecords()

	e.session.Current = executor.Select(e.session.Current, fields)
	e.session.UpdateSchema()
	fmt.Printf("✓ Applied SELECT to %d records (was: %d)\n", e.session.TotalRecords(), was)

	return nil
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

func (e *Engine) handleShow(args string) error {
	limit := 0
	offset := 0

	// Parse SHOW LIMIT 10 OFFSET 20 arguments
	parts := strings.Fields(strings.ToUpper(args))
	for i := 0; i < len(parts); i++ {
		switch parts[i] {
		case "LIMIT":
			if i+1 >= len(parts) {
				return fmt.Errorf("LIMIT requires a number")
			}
			if _, err := fmt.Sscanf(parts[i+1], "%d", &limit); err != nil {
				return fmt.Errorf("invalid LIMIT value: %s", parts[i+1])
			}
			i++
		case "OFFSET":
			if i+1 >= len(parts) {
				return fmt.Errorf("OFFSET requires a number")
			}
			if _, err := fmt.Sscanf(parts[i+1], "%d", &offset); err != nil {
				return fmt.Errorf("invalid OFFSET value: %s", parts[i+1])
			}
			i++
		}
	}

	return executor.Show(e.session.Current, limit, offset)
}

func (e *Engine) handleSave(args string) error {
	fileName := ""
	formatName := ""

	// Parse SAVE AS <filename> FORMAT <format> arguments
	parts := strings.Fields(args)
	for i := 0; i < len(parts); i++ {
		switch strings.ToUpper(parts[i]) {
		case "AS":
			if i+1 >= len(parts) {
				return fmt.Errorf("AS requires a filename")
			}
			fileName = parts[i+1]
			i++
		case "FORMAT":
			if i+1 >= len(parts) {
				return fmt.Errorf("FORMAT requires a format name")
			}
			formatName = parts[i+1]
			i++
		}
	}

	// Determine format
	f := e.session.Format
	if formatName != "" {
		var err error
		f, err = format.Get(formatName)
		if err != nil {
			return fmt.Errorf("unknown format: %s", formatName)
		}
	}

	return executor.Save(e.session.Current, e.session.FilePath, fileName, f)
}

func (e *Engine) handleExclude(args string) error {
	fields := splitFields(args)

	e.session.Current = executor.Exclude(e.session.Current, fields)
	e.session.UpdateSchema()
	fmt.Printf("✓ Applied EXCLUDE to %d records\n", e.session.TotalRecords())

	return nil
}

func (e *Engine) handleSort(args string) error {
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
			return fmt.Errorf("invalid sort direction: %s, expected ASC or DESC", parts[1])
		}
	}

	sorted, err := executor.Sort(e.session.Current, field, direction)
	if err != nil {
		return err
	}

	e.session.Current = sorted
	fmt.Printf("✓ Sorted %d records by %q %s\n", e.session.TotalRecords(), field, direction)

	return nil
}

func (e *Engine) handleFlatten(args string) error {
	fieldName := strings.TrimSpace(args)
	was := e.session.TotalRecords()

	flattened, err := executor.Flatten(e.session.Current, fieldName)
	if err != nil {
		return fmt.Errorf("failed to flatten field %q: %w", fieldName, err)
	}

	e.session.Current = flattened
	e.session.UpdateSchema()
	fmt.Printf("✓ Flattened field %q in %d records (was: %d)\n", fieldName, e.session.TotalRecords(), was)

	return nil
}
