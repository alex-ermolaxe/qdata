package engine

import (
	"testing"

	"github.com/alex-ermolaxe/qdata/internal/format"
	"github.com/iancoleman/orderedmap"
)

func TestHandleTransformCommand(t *testing.T) {
	asset := orderedmap.New()
	asset.Set("name", "K06")
	record := format.NewRecord()
	record.Set("sap_order_id", "000001810600")
	record.Set("assets", []any{asset})

	engine := &Engine{session: NewSession("input.json", nil, []format.Record{record})}
	if err := engine.handleCommand("TRANSFORM TO-MAP KEY sap_order_id VALUE assets.name"); err != nil {
		t.Fatalf("handleCommand() error = %v", err)
	}

	if len(engine.session.Current) != 1 {
		t.Fatalf("current records = %d, want 1", len(engine.session.Current))
	}
	value, ok := engine.session.Current[0].Get("000001810600")
	if !ok {
		t.Fatal("transformed record is missing dynamic key")
	}
	names, ok := value.([]any)
	if !ok || len(names) != 1 || names[0] != "K06" {
		t.Errorf("dynamic key value = %#v, want []any{\"K06\"}", value)
	}
}

func TestParseTransformArgs(t *testing.T) {
	tests := []struct {
		name      string
		args      string
		wantKey   string
		wantValue string
		wantErr   bool
	}{
		{
			name:      "valid arguments",
			args:      "TO-MAP KEY sap_order_id VALUE assets.name",
			wantKey:   "sap_order_id",
			wantValue: "assets.name",
		},
		{
			name:      "case insensitive keywords",
			args:      "to-map key order.id value assets.name",
			wantKey:   "order.id",
			wantValue: "assets.name",
		},
		{
			name:    "missing value path",
			args:    "TO-MAP KEY sap_order_id VALUE",
			wantErr: true,
		},
		{
			name:    "unsupported transform",
			args:    "FROM-JSON KEY id VALUE name",
			wantErr: true,
		},
		{
			name:    "extra arguments",
			args:    "TO-MAP KEY id VALUE name EXTRA",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotKey, gotValue, err := parseTransformArgs(tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseTransformArgs() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if gotKey != tt.wantKey || gotValue != tt.wantValue {
				t.Errorf("parseTransformArgs() = (%q, %q), want (%q, %q)",
					gotKey, gotValue, tt.wantKey, tt.wantValue)
			}
		})
	}
}
