package engine

import (
	"strings"
	"testing"

	"github.com/alex-ermolaxe/qdata/internal/format"
	"github.com/alex-ermolaxe/qdata/internal/schema"
	"github.com/iancoleman/orderedmap"
)

func TestExcludeSortFlattenAndSchemaCommandSpecs(t *testing.T) {
	t.Run("EXCLUDE", func(t *testing.T) {
		record := format.NewRecord()
		record.Set("id", 1)
		record.Set("name", "Alice")
		e := &Engine{session: NewSession("input.json", nil, []format.Record{record})}

		if _, err := e.ExecuteCommand("EXCLUDE name"); err != nil {
			t.Fatalf("EXCLUDE error = %v", err)
		}
		if _, ok := e.session.Current[0].Get("name"); ok {
			t.Error("EXCLUDE retained the removed field")
		}
	})

	t.Run("SORT", func(t *testing.T) {
		first := format.NewRecord()
		first.Set("rank", 1)
		second := format.NewRecord()
		second.Set("rank", 2)
		e := &Engine{session: NewSession("input.json", nil, []format.Record{first, second})}

		if _, err := e.ExecuteCommand("SORT rank DESC"); err != nil {
			t.Fatalf("SORT error = %v", err)
		}
		if got, _ := e.session.Current[0].Get("rank"); got != 2 {
			t.Errorf("first sorted rank = %v, want 2", got)
		}
	})

	t.Run("FLATTEN", func(t *testing.T) {
		address := orderedmap.New()
		address.Set("city", "Moscow")
		record := format.NewRecord()
		record.Set("address", address)
		e := &Engine{session: NewSession("input.json", nil, []format.Record{record})}

		if _, err := e.ExecuteCommand("FLATTEN address"); err != nil {
			t.Fatalf("FLATTEN error = %v", err)
		}
		if city, ok := e.session.Current[0].Get("city"); !ok || city != "Moscow" {
			t.Errorf("flattened city = %v, exists=%v; want Moscow", city, ok)
		}
	})

	t.Run("SCHEMA", func(t *testing.T) {
		record := format.NewRecord()
		record.Set("rank", 1)
		e := &Engine{session: NewSession("input.json", nil, []format.Record{record})}

		result, err := e.ExecuteCommand("SCHEMA")
		if err != nil {
			t.Fatalf("SCHEMA error = %v", err)
		}
		if !strings.Contains(result.Output, "rank") {
			t.Errorf("SCHEMA output = %q, want field rank", result.Output)
		}
	})
}

func TestSortExcludeAndFlattenCompletion(t *testing.T) {
	address := orderedmap.New()
	address.Set("city", "Moscow")
	record := format.NewRecord()
	record.Set("id", 1)
	record.Set("address", address)
	dataSchema := schema.Infer([]format.Record{record}, 100)

	if got := completeSort("SORT ", dataSchema); !containsString(got, "id") {
		t.Errorf("SORT field completions = %v, want id", got)
	}
	if got := completeSort("SORT id ", dataSchema); !containsString(got, "DESC") {
		t.Errorf("SORT direction completions = %v, want DESC", got)
	}
	if got := completeExclude("EXCLUDE ", dataSchema); !containsString(got, "id") {
		t.Errorf("EXCLUDE completions = %v, want id", got)
	}
	if got := completeFlatten("FLATTEN ", dataSchema); !containsString(got, "address") {
		t.Errorf("FLATTEN completions = %v, want address", got)
	}
}
