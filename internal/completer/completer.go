package completer

import (
	"strings"
	"unicode"

	"github.com/alex-ermolaxe/qdata/internal/schema"
	"github.com/chzyer/readline"
)

// CompletionFunc returns candidates for the input text before the cursor.
type CompletionFunc func(input string, schema *schema.Schema) []string

// CommandCompletion supplies a command name and its argument completion strategy.
type CommandCompletion struct {
	Name     string
	Complete CompletionFunc
}

// Completer - implements readline.AutoCompleter
type Completer struct {
	schemaGetter      func() *schema.Schema
	commandNames      []string
	commandCompletion map[string]CompletionFunc
}

// New creates a Completer using the schema getter and command-specific strategies.
func New(schemaGetter func() *schema.Schema, commands []CommandCompletion) *Completer {
	completer := &Completer{
		schemaGetter:      schemaGetter,
		commandNames:      make([]string, 0, len(commands)),
		commandCompletion: make(map[string]CompletionFunc, len(commands)),
	}
	for _, command := range commands {
		completer.commandNames = append(completer.commandNames, command.Name)
		if command.Complete != nil {
			completer.commandCompletion[command.Name] = command.Complete
		}
	}
	return completer
}

// Do implements readline.AutoCompleter interface
func (c *Completer) Do(line []rune, pos int) (newLine [][]rune, length int) {
	original := string(line[:pos])
	input := strings.ToUpper(original)
	trimmed := strings.TrimLeft(input, " ")
	parts := strings.Fields(trimmed)

	if len(parts) == 0 {
		return toRunes(c.commandNames), 0
	}

	// lastWord is taken from original to preserve case
	originalTrimmed := strings.TrimLeft(original, " ")
	originalParts := strings.Fields(originalTrimmed)
	lastWord := ""
	if !strings.HasSuffix(original, " ") && len(originalParts) > 0 {
		lastWord = originalParts[len(originalParts)-1]
	}

	if len(parts) == 1 && !strings.HasSuffix(input, " ") {
		return completeFrom(c.commandNames, lastWord)
	}

	command := parts[0]
	if completion := c.commandCompletion[command]; completion != nil {
		return completeFrom(completion(original, c.schemaGetter()), lastWord)
	}

	return nil, 0
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
	if s == "" {
		return false
	}
	return unicode.IsLower(rune(s[0]))
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
