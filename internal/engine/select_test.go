package engine

import (
	"testing"

	"github.com/alex-ermolaxe/qdata/internal/completer"
	"github.com/alex-ermolaxe/qdata/internal/format"
	"github.com/alex-ermolaxe/qdata/internal/schema"
	"github.com/iancoleman/orderedmap"
)

func TestSelectCommandSpec(t *testing.T) {
	record := format.NewRecord()
	record.Set("id", 1)
	record.Set("name", "Alice")
	e := &Engine{session: NewSession("input.json", nil, []format.Record{record})}

	if _, err := e.ExecuteCommand("SELECT name"); err != nil {
		t.Fatalf("ExecuteCommand() error = %v", err)
	}
	if _, ok := e.session.Current[0].Get("id"); ok {
		t.Error("SELECT retained an unrequested field")
	}
	if value, ok := e.session.Current[0].Get("name"); !ok || value != "Alice" {
		t.Errorf("selected name = %v, exists=%v; want Alice", value, ok)
	}

	if _, err := e.ExecuteCommand("SELECT "); err == nil {
		t.Error("SELECT with no fields did not return an error")
	}
}

func TestSelectCommandCompletion(t *testing.T) {
	address := orderedmap.New()
	address.Set("city", "Moscow")
	record := format.NewRecord()
	record.Set("address", address)
	dataSchema := schema.Infer([]format.Record{record}, 100)

	got := selectCommand.Completion(completer.CompletionContext{
		Arguments:   []string{"address."},
		CurrentWord: "address.",
		Schema:      dataSchema,
	})
	if !containsString(got, "address.city") {
		t.Errorf("SELECT completions = %v, want address.city", got)
	}
}
