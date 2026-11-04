package schema

import (
	"testing"

	"github.com/alex-ermolaxe/qdata/internal/format"
	"github.com/iancoleman/orderedmap"
)

func TestInferFieldsFromObjectArray(t *testing.T) {
	first := orderedmap.New()
	first.Set("name", "keyboard")
	first.Set("details", map[string]any{
		"color": "black",
	})
	second := orderedmap.New()
	second.Set("name", "mouse")
	second.Set("details", map[string]any{
		"color":    "white",
		"wireless": true,
	})

	record := format.NewRecord()
	record.Set("items", []any{first, second})

	got := Infer([]format.Record{record}, 100)
	paths := make(map[string]bool)
	for _, path := range got.AllPaths() {
		paths[path] = true
	}

	for _, want := range []string{"items", "items.name", "items.details", "items.details.color", "items.details.wireless"} {
		if !paths[want] {
			t.Errorf("schema is missing path %q", want)
		}
	}

	field, ok := got.Get("items")
	if !ok {
		t.Fatal("schema is missing top-level items field")
	}
	if field.Type != TypeArray {
		t.Errorf("items type = %q, want %q", field.Type, TypeArray)
	}

	topLevel := make(map[string]bool)
	for _, field := range got.Fields {
		topLevel[field.Path] = true
	}
	if topLevel["items.name"] {
		t.Error("nested array fields should not be listed as top-level fields")
	}
}
