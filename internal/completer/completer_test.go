package completer

import (
	"strings"
	"testing"

	"github.com/alex-ermolaxe/qdata/internal/format"
	"github.com/alex-ermolaxe/qdata/internal/schema"
	"github.com/iancoleman/orderedmap"
)

func TestTransformAutocomplete(t *testing.T) {
	asset := orderedmap.New()
	asset.Set("name", "K06")
	record := format.NewRecord()
	record.Set("sap_order_id", "order-1")
	record.Set("assets", []any{asset})
	dataSchema := schema.Infer([]format.Record{record}, 100)
	completer := New(func() *schema.Schema { return dataSchema })

	tests := []struct {
		input string
		want  string
	}{
		{input: "TRANSFORM ", want: "TO-MAP"},
		{input: "TRANSFORM TO-MAP ", want: "KEY"},
		{input: "TRANSFORM TO-MAP KEY ", want: "sap_order_id"},
		{input: "TRANSFORM TO-MAP KEY sap_order_id ", want: "VALUE"},
		{input: "TRANSFORM TO-MAP KEY sap_order_id VALUE assets.", want: "name"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, _ := completer.Do([]rune(tt.input), len([]rune(tt.input)))
			for _, suggestion := range got {
				if string(suggestion) == tt.want {
					return
				}
			}
			t.Errorf("autocomplete for %q = %q, want suggestion %q", tt.input, suggestions(got), tt.want)
		})
	}
}

func suggestions(options [][]rune) string {
	values := make([]string, len(options))
	for i, option := range options {
		values[i] = string(option)
	}
	return strings.Join(values, ", ")
}
