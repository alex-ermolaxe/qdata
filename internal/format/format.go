package format

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/iancoleman/orderedmap"
)

// Record - basic type for a single record
type Record = *orderedmap.OrderedMap

func NewRecord() Record {
	return orderedmap.New()
}

// Format - interface for working with a specific file format.
// To add a new format - just implement this interface.
type Format interface {
	// Decode reads data from reader and returns an array of records
	Decode(r io.Reader) ([]Record, error)

	// Encode serializes an array of records to writer
	Encode(w io.Writer, records []Record) error

	// Extensions returns a list of file extensions for auto-detection
	Extensions() []string
}

// registry - registry of registered formats
var registry = map[string]Format{}

// Register registers a new format by name
func Register(name string, f Format) {
	registry[name] = f
}

// Detect determines the format by file extension
func Detect(filename string) (Format, error) {
	ext := strings.TrimPrefix(filepath.Ext(filename), ".")
	for _, f := range registry {
		for _, e := range f.Extensions() {
			if e == ext {
				return f, nil
			}
		}
	}
	return nil, fmt.Errorf("unsupported format: .%s", ext)
}

// Get returns a format by name
func Get(name string) (Format, error) {
	f, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown format: %s", name)
	}
	return f, nil
}
