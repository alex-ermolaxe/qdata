package json

import (
	gojson "encoding/json"
	"fmt"
	"io"

	"github.com/alex-ermolaxe/qdata/internal/format"
	"github.com/iancoleman/orderedmap"
)

type JSONFormat struct{}

// JSONFormat - implementation of Format interface for JSON
func (j *JSONFormat) Decode(r io.Reader) ([]format.Record, error) {
	// Decode into a slice of OrderedMap
	var raw []orderedmap.OrderedMap

	decoder := gojson.NewDecoder(r)
	if err := decoder.Decode(&raw); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	records := make([]format.Record, len(raw))
	for i := range raw {
		records[i] = &raw[i]
	}

	return records, nil
}

func (j *JSONFormat) Encode(w io.Writer, records []format.Record) error {
	encoder := gojson.NewEncoder(w)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(records); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

func (j *JSONFormat) Extensions() []string {
	return []string{"json"}
}

func Register() {
	format.Register("json", &JSONFormat{})
}
