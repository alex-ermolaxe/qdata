package completer

import (
	"strings"
	"unicode"

	"github.com/alex-ermolaxe/qdata/internal/schema"
	"github.com/chzyer/readline"
)

// CompletionFunc returns candidates for the input text before the cursor.
type CompletionFunc func(context CompletionContext) []string

// CompletionContext contains the parsed input needed by a completion strategy.
type CompletionContext struct {
	Arguments        []string
	CurrentWord      string
	HasTrailingSpace bool
	Schema           *schema.Schema
}

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
	trimmed := strings.TrimLeft(original, " ")
	parts := strings.Fields(trimmed)

	if len(parts) == 0 {
		return toRunes(c.commandNames), 0
	}

	hasTrailingSpace := strings.HasSuffix(original, " ")
	currentWord := ""
	if !hasTrailingSpace {
		currentWord = parts[len(parts)-1]
	}

	if len(parts) == 1 && !hasTrailingSpace {
		return completeFrom(c.commandNames, currentWord)
	}

	command := strings.ToUpper(parts[0])
	if completion := c.commandCompletion[command]; completion != nil {
		context := CompletionContext{
			Arguments:        parts[1:],
			CurrentWord:      currentWord,
			HasTrailingSpace: hasTrailingSpace,
			Schema:           c.schemaGetter(),
		}
		options := completion(context)

		return completeFrom(options, context.CurrentWord)
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
