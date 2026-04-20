package engine

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/alex-ermolaxe/qdata/internal/completer"
	"github.com/alex-ermolaxe/qdata/internal/format"
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

	default:
		fmt.Printf("unknown command: %s\n", command)
	}

	return nil
}
