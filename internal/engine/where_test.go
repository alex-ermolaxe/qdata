package engine

import (
	"strings"
	"testing"

	"github.com/alex-ermolaxe/qdata/internal/completer"
	"github.com/alex-ermolaxe/qdata/internal/format"
	"github.com/alex-ermolaxe/qdata/internal/schema"
	"github.com/iancoleman/orderedmap"
)

func TestWhereCommandSpec(t *testing.T) {
	active := format.NewRecord()
	active.Set("status", "active")
	inactive := format.NewRecord()
	inactive.Set("status", "inactive")
	e := &Engine{session: NewSession("input.json", nil, []format.Record{active, inactive})}

	result, err := e.ExecuteCommand(`  WHERE status = "active"  `)
	if err != nil {
		t.Fatalf("ExecuteCommand() error = %v", err)
	}
	if !strings.Contains(result.Output, "Found: 1 records (was: 2)") {
		t.Errorf("WHERE output = %q, want one matching record", result.Output)
	}
	if got := e.RecordCount(); got != 1 {
		t.Errorf("record count = %d, want 1", got)
	}

	if _, err := e.ExecuteCommand("WHERE "); err == nil {
		t.Error("WHERE with an empty condition did not return an error")
	}
}

func TestCompleteFieldPaths(t *testing.T) {
	address := orderedmap.New()
	address.Set("city", "Moscow")
	record := format.NewRecord()
	record.Set("status", "active")
	record.Set("address", address)
	dataSchema := schema.Infer([]format.Record{record}, 100)

	topLevel := completeFieldPaths(completer.CompletionContext{
		HasTrailingSpace: true,
		Schema:           dataSchema,
	})
	if !containsString(topLevel, "status") || containsString(topLevel, "address.city") {
		t.Errorf("top-level completions = %v, want top-level fields only", topLevel)
	}

	nested := completeFieldPaths(completer.CompletionContext{
		Arguments:   []string{"address."},
		CurrentWord: "address.",
		Schema:      dataSchema,
	})
	if !containsString(nested, "address.city") {
		t.Errorf("nested completions = %v, want address.city", nested)
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}

	return false
}
