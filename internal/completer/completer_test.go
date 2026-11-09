package completer

import (
	"strings"
	"testing"

	"github.com/alex-ermolaxe/qdata/internal/schema"
)

func TestCommandCompletionStrategy(t *testing.T) {
	completer := New(
		func() *schema.Schema { return nil },
		[]CommandCompletion{
			{Name: "WHERE", Complete: func(string, *schema.Schema) []string {
				return []string{"status"}
			}},
		},
	)

	got, _ := completer.Do([]rune("WHERE st"), len([]rune("WHERE st")))
	if len(got) != 1 || string(got[0]) != "atus" {
		t.Errorf("command-specific completion = %q, want [atus]", suggestions(got))
	}
}

func suggestions(options [][]rune) string {
	values := make([]string, len(options))
	for i, option := range options {
		values[i] = string(option)
	}
	return strings.Join(values, ", ")
}
