package completer

import (
	"strings"
	"unicode"

	"github.com/alex-ermolaxe/qdata/internal/schema"
	"github.com/chzyer/readline"
)

// keywords — список доступных команд
var keywords = []string{
	"WHERE", "SELECT", "EXCLUDE", "SORT", "SHOW",
	"SAVE", "RESET", "COUNT", "SCHEMA", "EXIT",
}

// sortDirections — направления сортировки
var sortDirections = []string{"ASC", "DESC"}

// saveFormats — доступные форматы сохранения
var saveFormats = []string{"json", "xml", "csv"}

// Completer — реализует readline.AutoCompleter
type Completer struct {
	schema *schema.Schema
}

// New создаёт новый Completer на основе схемы
func New(s *schema.Schema) *Completer {
	return &Completer{schema: s}
}

// Do реализует интерфейс readline.AutoCompleter
func (c *Completer) Do(line []rune, pos int) (newLine [][]rune, length int) {
	original := string(line[:pos])
	input := strings.ToUpper(original)
	trimmed := strings.TrimLeft(input, " ")
	parts := strings.Fields(trimmed)

	if len(parts) == 0 {
		return toRunes(keywords), 0
	}

	// lastWord берём из оригинала чтобы сохранить регистр
	originalTrimmed := strings.TrimLeft(original, " ")
	originalParts := strings.Fields(originalTrimmed)
	lastWord := ""
	if !strings.HasSuffix(original, " ") && len(originalParts) > 0 {
		lastWord = originalParts[len(originalParts)-1]
	}

	if len(parts) == 1 && !strings.HasSuffix(input, " ") {
		return completeFrom(keywords, lastWord)
	}

	command := parts[0]

	switch command {
	case "WHERE", "SELECT", "EXCLUDE":
		return c.completeFields(lastWord)
	case "SORT":
		if len(parts) >= 2 && strings.HasSuffix(input, " ") {
			return completeFrom(sortDirections, "")
		}
		if len(parts) == 3 {
			return completeFrom(sortDirections, lastWord)
		}
		return c.completeFields(lastWord)
	case "SAVE":
		if len(parts) >= 2 && strings.ToUpper(parts[len(parts)-1]) == "FORMAT" {
			return completeFrom(saveFormats, "")
		}
		if len(parts) >= 3 && strings.ToUpper(parts[len(parts)-2]) == "FORMAT" {
			return completeFrom(saveFormats, lastWord)
		}
	}

	return nil, 0
}

// completeFields предлагает поля схемы с учётом вложенности
func (c *Completer) completeFields(prefix string) ([][]rune, int) {
	allPaths := c.schema.AllPaths()

	// Если есть точка — ищем вложенные поля
	if dotIdx := strings.LastIndex(prefix, "."); dotIdx >= 0 {
		parentPath := prefix[:dotIdx]
		childPrefix := prefix[dotIdx+1:]
		_ = childPrefix

		var matches []string
		for _, path := range allPaths {
			if strings.HasPrefix(path, parentPath+".") {
				rest := path[len(parentPath)+1:]
				if !strings.Contains(rest, ".") {
					matches = append(matches, path)
				}
			}
		}
		return completeFrom(matches, prefix)
	}

	return completeFrom(allPaths, prefix)
}

// completeFrom фильтрует список по префиксу и возвращает подходящие варианты
func completeFrom(options []string, prefix string) ([][]rune, int) {
	upperPrefix := strings.ToUpper(prefix)
	var matches []string

	for _, opt := range options {
		if strings.HasPrefix(opt, upperPrefix) {
			matches = append(matches, opt)
		}
	}

	if len(matches) == 0 {
		return nil, 0
	}

	suffixes := make([][]rune, len(matches))
	for i, match := range matches {
		suffix := match[len(prefix):]

		if isLower(prefix) {
			suffix = strings.ToLower(suffix)
		}

		suffixes[i] = []rune(suffix)
	}

	return suffixes, 0
}

// toRunes конвертирует []string в [][]rune для readline
func toRunes(options []string) [][]rune {
	result := make([][]rune, len(options))
	for i, opt := range options {
		result[i] = []rune(opt)
	}
	return result
}

func isLower(s string) bool {
	return unicode.IsLower(rune(s[0]))
}

// Newreadline создаёт конфигурацию readline с подключенным автодополнением
func NewReadline(c *Completer) *readline.Config {
	return &readline.Config{
		Prompt:          "> ",
		AutoComplete:    c,
		HistoryLimit:    100,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	}
}
