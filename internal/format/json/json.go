package json

import (
	gojson "encoding/json"
	"fmt"
	"io"

	"github.com/alex-ermolaxe/qdata/internal/format"
)

// JSONFormat — реализация Format интерфейса для JSON
type JSONFormat struct{}

func (j *JSONFormat) Decode(r io.Reader) ([]format.Record, error) {
	var records []format.Record

	decoder := gojson.NewDecoder(r)
	if err := decoder.Decode(&records); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
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

// Register регистрирует JSON формат в реестре
func Register() {
	format.Register("json", &JSONFormat{})
}
