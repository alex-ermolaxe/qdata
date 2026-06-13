package completer

import (
	"strings"
	"unicode"

	"github.com/alex-ermolaxe/qdata/internal/schema"
	"github.com/chzyer/readline"
)

// keywords - list of available commands
var keywords = []string{
	"WHERE", "SELECT", "EXCLUDE", "SORT", "SHOW",
	"SAVE", "RESET", "COUNT", "SCHEMA", "EXIT",
}

// sortDirections - sort directions
var sortDirections = []string{"ASC", "DESC"}

// saveFormats - available save formats
var saveFormats = []string{"json", "xml", "csv"}

// Completer - implements readline.AutoCompleter
type Completer struct {
	schema *schema.Schema
}

// New creates a new Completer based on schema
func New(s *schema.Schema) *Completer {
	return &Completer{schema: s}
}

// Do implements readline.AutoCompleter interface
func (c *Completer) Do(line []rune, pos int) (newLine [][]rune, length int) {
	original := string(line[:pos])
	input := strings.ToUpper(original)
	trimmed := strings.TrimLeft(input, " ")
	parts := strings.Fields(trimmed)

	if len(parts) == 0 {
		return toRunes(keywords), 0
	}

	// lastWord is taken from original to preserve case
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
	case "SHOW":
		if len(parts) >= 2 && strings.HasSuffix(input, " ") {
			last := strings.ToUpper(parts[len(parts)-1])
			if last == "LIMIT" || last == "OFFSET" {
				return nil, 0 // after LIMIT/OFFSET we expect a number
			}
			return completeFrom([]string{"LIMIT", "OFFSET"}, "")
		}
		if !strings.HasSuffix(input, " ") && len(parts) >= 2 {
			return completeFrom([]string{"LIMIT", "OFFSET"}, lastWord)
		}
		return completeFrom([]string{"LIMIT", "OFFSET"}, "")
	}

	return nil, 0
}

// NewReadline creates a readline configuration with connected autocomplete
func NewReadline(c *Completer) *readline.Config {
	return &readline.Config{
		Prompt:          "> ",
		AutoComplete:    c,
		HistoryLimit:    100,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	}
}

// completeFields suggests schema fields considering nesting
func (c *Completer) completeFields(prefix string) ([][]rune, int) {
	allPaths := c.schema.AllPaths()
	upperPrefix := strings.ToUpper(prefix)

	// If there's a dot - suggest only nested fields of the appropriate parent
	if dotIdx := strings.LastIndex(upperPrefix, "."); dotIdx >= 0 {
		var matches []string
		for _, path := range allPaths {
			if strings.HasPrefix(strings.ToUpper(path), upperPrefix) {
				matches = append(matches, path)
			}
		}
		return completeFrom(matches, prefix)
	}

	// Otherwise suggest all top-level fields
	var topLevel []string
	for _, path := range allPaths {
		if !strings.Contains(path, ".") {
			topLevel = append(topLevel, path)
		}
	}

	return completeFrom(topLevel, prefix)
}

// completeFrom filters the list by prefix and returns suitable options
func completeFrom(options []string, prefix string) ([][]rune, int) {
	upperPrefix := strings.ToUpper(prefix)
	var matches []string

	for _, opt := range options {
		if strings.HasPrefix(strings.ToUpper(opt), upperPrefix) {
			matches = append(matches, opt)
		}
	}

	if len(matches) == 0 {
		return nil, 0
	}

	suffixes := make([][]rune, len(matches))
	for i, match := range matches {
		if match == strings.ToUpper(match) {
			suffix := match[len(prefix):]
			if isLower(prefix) {
				suffix = strings.ToLower(suffix)
			}
			suffixes[i] = []rune(suffix)
		} else {
			suffixes[i] = []rune(match[len(prefix):])
		}
	}

	return suffixes, 0
}

// toRunes converts []string to [][]rune for readline
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
