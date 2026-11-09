package schema

import (
	"strings"
	"testing"

	"github.com/alex-ermolaxe/qdata/internal/format"
	"github.com/iancoleman/orderedmap"
)

func TestInferPrimitiveArrayElementType(t *testing.T) {
	emptyTags := format.NewRecord()
	emptyTags.Set("tags", []any{})
	populatedTags := format.NewRecord()
	populatedTags.Set("tags", []any{"admin", "user"})

	got := Infer([]format.Record{emptyTags, populatedTags}, 100)
	field := got.index["tags"]
	if field == nil {
		t.Fatal("schema is missing tags field")
	}
	if field.Type != TypeArray {
		t.Errorf("tags type = %q, want %q", field.Type, TypeArray)
	}
	if field.ElementType != TypeString {
		t.Errorf("tags element type = %q, want %q", field.ElementType, TypeString)
	}
	if !strings.Contains(got.Render(), "array<string>") {
		t.Errorf("rendered schema = %q, want array<string>", got.Render())
	}
}

func TestInferMixedArrayElementTypeAsUnknown(t *testing.T) {
	record := format.NewRecord()
	record.Set("tags", []any{"admin", 1})

	got := Infer([]format.Record{record}, 100)
	field := got.index["tags"]
	if field == nil {
		t.Fatal("schema is missing tags field")
	}
	if field.ElementType != TypeUnknown {
		t.Errorf("tags element type = %q, want %q", field.ElementType, TypeUnknown)
	}
}

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

	var field *Field
	for _, topLevelField := range got.Fields {
		if topLevelField.Path == "items" {
			field = topLevelField
			break
		}
	}
	if field == nil {
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
