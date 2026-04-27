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
	"github.com/chzyer/readline"
)

const version = "1.0"

// Engine — основной движок приложения
type Engine struct {
	session *Session
}

// New создаёт новый Engine, загружает файл и строит сессию
func New(filePath string, formatName string) (*Engine, error) {
	// Определяем формат
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

	// Открываем файл
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Декодируем данные
	records, err := f.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode file: %w", err)
	}

	// Создаём сессию
	session := NewSession(filePath, f, records)

	return &Engine{session: session}, nil
}

// Run запускает интерактивную REPL сессию
func (e *Engine) Run() error {
	// Создаём автодополнение
	c := completer.New(e.session.Schema)
	config := completer.NewReadline(c)

	// Инициализируем readline
	rl, err := readline.NewEx(config)
	if err != nil {
		return fmt.Errorf("failed to initialize readline: %w", err)
	}
	defer rl.Close()

	// Приветственное сообщение
	fileName := filepath.Base(e.session.FilePath)
	fmt.Printf("qdata v%s | file: %s | records: %d\n\n",
		version,
		fileName,
		e.session.OriginalRecords(),
	)

	// REPL цикл
	for {
		line, err := rl.Readline()
		if err != nil {
			// Ctrl+D или Ctrl+C
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

		// Обрабатываем команду
		if err := e.handleCommand(input); err != nil {
			fmt.Printf("error: %s\n", err)
		}
	}
}

// handleCommand обрабатывает введённую команду
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
	fmt.Printf("✓ Applied SELECT to %d records (was: %d)\n", e.session.TotalRecords(), was)

	return nil
}

// splitFields разбивает строку полей через запятую
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

	// Разбираем аргументы SHOW LIMIT 10 OFFSET 20
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

	// Разбираем аргументы: SAVE AS <filename> FORMAT <format>
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

	// Определяем формат
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
