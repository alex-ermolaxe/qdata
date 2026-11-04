package executor_test

import (
	"strings"
	"testing"

	"github.com/alex-ermolaxe/qdata/internal/executor"
	formatjson "github.com/alex-ermolaxe/qdata/internal/format/json"
	"github.com/iancoleman/orderedmap"
)

func TestTransformToMap(t *testing.T) {
	input := `[{
		"sap_order_id": "000001810600",
		"assets": [{"name": "K06"}, {"name": "K610"}],
		"other": "ignored"
	}]`
	records, err := (&formatjson.JSONFormat{}).Decode(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	got, err := executor.TransformToMap(records, "sap_order_id", "assets.name")
	if err != nil {
		t.Fatalf("TransformToMap() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("TransformToMap() returned %d records, want 1", len(got))
	}
	value, ok := got[0].Get("000001810600")
	if !ok {
		t.Fatal("result missing key \"000001810600\"")
	}
	assets, ok := value.([]any)
	if !ok {
		t.Fatalf("mapped value has type %T, want []any", value)
	}
	if len(assets) != 2 || assets[0] != "K06" || assets[1] != "K610" {
		t.Errorf("mapped assets = %#v, want []any{\"K06\", \"K610\"}", assets)
	}
	if _, ok := got[0].Get("assets"); ok {
		t.Error("result retained original field \"assets\"")
	}
	if _, ok := records[0].Get("sap_order_id"); !ok {
		t.Error("TransformToMap modified the source record")
	}
}

func TestTransformToMapValuePathCanSelectWholeArray(t *testing.T) {
	first := orderedmap.New()
	first.Set("name", "K06")
	record := createTestRecord(map[string]any{
		"id": 1.0,
	})
	record.Set("assets", []any{first})

	got, err := executor.TransformToMap([]*orderedmap.OrderedMap{record}, "id", "assets")
	if err != nil {
		t.Fatalf("TransformToMap() error = %v", err)
	}
	value, ok := got[0].Get("1")
	if !ok {
		t.Fatal("result missing key \"1\"")
	}
	assets, ok := value.([]any)
	if !ok || len(assets) != 1 || assets[0] != first {
		t.Errorf("mapped value = %#v, want original assets array", value)
	}
}

func TestTransformToMapErrors(t *testing.T) {
	validRecord := createTestRecord(map[string]any{
		"id":    "order-1",
		"value": "item",
	})
	recordWithoutKey := createTestRecord(map[string]any{
		"value": "item",
	})
	recordWithObjectKey := createTestRecord(map[string]any{
		"id":    map[string]any{"nested": "not scalar"},
		"value": "item",
	})
	tests := []struct {
		name       string
		records    []*orderedmap.OrderedMap
		keyPath    string
		valuePath  string
		wantErrSub string
	}{
		{
			name:       "missing key",
			records:    []*orderedmap.OrderedMap{recordWithoutKey},
			keyPath:    "id",
			valuePath:  "value",
			wantErrSub: `key field "id" does not exist`,
		},
		{
			name:       "missing value",
			records:    []*orderedmap.OrderedMap{validRecord},
			keyPath:    "id",
			valuePath:  "missing",
			wantErrSub: `value field "missing" does not exist`,
		},
		{
			name:       "non-scalar key",
			records:    []*orderedmap.OrderedMap{recordWithObjectKey},
			keyPath:    "id",
			valuePath:  "value",
			wantErrSub: `key field "id" must contain a non-null scalar`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := executor.TransformToMap(tt.records, tt.keyPath, tt.valuePath)
			if err == nil || !strings.Contains(err.Error(), tt.wantErrSub) {
				t.Fatalf("TransformToMap() error = %v, want substring %q", err, tt.wantErrSub)
			}
		})
	}
}
