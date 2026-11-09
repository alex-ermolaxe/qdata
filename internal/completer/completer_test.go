package completer

import (
	"strings"
	"testing"

	"github.com/alex-ermolaxe/qdata/internal/schema"
)

func TestCommandCompletionStrategy(t *testing.T) {
	completer := New(
		func() *schema.Schema { return nil },
		[]CommandCompletion{{
			Name: "WHERE",
			Complete: func(CompletionContext) []string {
				return []string{"status"}
			},
		}},
	)

	got, _ := completer.Do([]rune("WHERE st"), len([]rune("WHERE st")))
	if len(got) != 1 || string(got[0]) != "atus" {
		t.Errorf("command-specific completion = %q, want [atus]", suggestions(got))
	}
}

func TestDoBuildsCompletionContext(t *testing.T) {
	currentSchema := &schema.Schema{}
	var got CompletionContext
	completer := New(
		func() *schema.Schema { return currentSchema },
		[]CommandCompletion{{
			Name: "WHERE",
			Complete: func(context CompletionContext) []string {
				got = context
				return nil
			},
		}},
	)

	input := []rune("  WHERE sta")
	completer.Do(input, len(input))
	if got.CurrentWord != "sta" || len(got.Arguments) != 1 || got.Arguments[0] != "sta" {
		t.Errorf("completion context for partial input = %#v, want current word and argument sta", got)
	}
	if got.HasTrailingSpace {
		t.Error("completion context for partial input has trailing space")
	}
	if got.Schema != currentSchema {
		t.Errorf("completion context schema = %p, want %p", got.Schema, currentSchema)
	}

	input = []rune("WHERE ")
	completer.Do(input, len(input))
	if got.CurrentWord != "" || len(got.Arguments) != 0 || !got.HasTrailingSpace {
		t.Errorf("completion context after command = %#v, want empty arguments and trailing space", got)
	}
}

func suggestions(options [][]rune) string {
	values := make([]string, len(options))
	for i, option := range options {
		values[i] = string(option)
	}

	return strings.Join(values, ", ")
}
