package schema

import (
	"fmt"
	"strings"

	"github.com/alex-ermolaxe/qdata/internal/format"
	"github.com/iancoleman/orderedmap"
)

// FieldType - field type
type FieldType string

const (
	TypeString  FieldType = "string"
	TypeNumber  FieldType = "number"
	TypeBool    FieldType = "bool"
	TypeArray   FieldType = "array"
	TypeObject  FieldType = "object"
	TypeUnknown FieldType = "unknown"
)

// Field - description of a schema field
type Field struct {
	Path     string    // full path: "address.city"
	Type     FieldType // field type
	Children []*Field  // nested fields (for objects)
}

// Schema - data schema, contains all fields
type Schema struct {
	Fields []*Field          // top-level fields
	index  map[string]*Field // index of all fields by path for fast access
}

// AllPaths returns all field paths including nested ones
func (s *Schema) AllPaths() []string {
	paths := make([]string, 0, len(s.index))
	for path := range s.index {
		paths = append(paths, path)
	}
	return paths
}

// Get returns a field by path
func (s *Schema) Get(path string) (*Field, bool) {
	f, ok := s.index[path]
	return f, ok
}

// ChildPaths returns paths of child fields for the specified prefix
// For example for "address" returns ["address.city", "address.zip"]
func (s *Schema) ChildPaths(prefix string) []string {
	var paths []string
	for path := range s.index {
		if len(path) > len(prefix)+1 &&
			path[:len(prefix)] == prefix &&
			path[len(prefix)] == '.' {
			// Take only direct children
			rest := path[len(prefix)+1:]
			hasNested := false
			for p := range s.index {
				if len(p) > len(path)+1 &&
					p[:len(path)] == path &&
					p[len(path)] == '.' {
					hasNested = true
					break
				}
			}
			_ = hasNested
			_ = rest
			paths = append(paths, path)
		}
	}
	return paths
}

// Infer analyzes records and builds a schema
// Analyzes the first maxRecords records to determine types
func Infer(records []format.Record, maxRecords int) *Schema {
	schema := &Schema{
		index: map[string]*Field{},
	}

	limit := maxRecords
	if limit > len(records) || limit <= 0 {
		limit = len(records)
	}

	for i := 0; i < limit; i++ {
		collectFields(records[i], "", schema)
	}

	return schema
}

// collectFields recursively traverses a record and collects fields
func collectFields(record format.Record, prefix string, schema *Schema) {
	for _, key := range record.Keys() {
		val, _ := record.Get(key)

		path := key
		if prefix != "" {
			path = prefix + "." + key
		}

		fieldType := inferType(val)

		if _, exists := schema.index[path]; !exists {
			field := &Field{
				Path: path,
				Type: fieldType,
			}
			schema.index[path] = field

			if prefix == "" {
				schema.Fields = append(schema.Fields, field)
			}
		}

		switch nested := val.(type) {
		case *orderedmap.OrderedMap:
			collectFields(nested, path, schema)
		case orderedmap.OrderedMap:
			collectFields(&nested, path, schema)
		}
	}
}

// inferType determines the type of a value
func inferType(val any) FieldType {
	if val == nil {
		return TypeUnknown
	}

	switch val.(type) {
	case string:
		return TypeString
	case float64, float32, int, int64, int32:
		return TypeNumber
	case bool:
		return TypeBool
	case []any:
		return TypeArray
	case orderedmap.OrderedMap:
		return TypeObject
	case *orderedmap.OrderedMap:
		return TypeObject
	default:
		return TypeUnknown
	}
}

// Print outputs the schema in a readable format
func (s *Schema) Print() {
	printFields(s.Fields, s.index, 0)
}

func printFields(fields []*Field, index map[string]*Field, depth int) {
	var indent strings.Builder
	prefix := ""
	for range depth {
		indent.WriteString("  ")
	}
	if depth > 0 {
		prefix = "└─ "
	}

	for _, f := range fields {
		parts := Split(f.Path)
		name := parts[len(parts)-1]

		if f.Type == TypeObject {
			fmt.Printf("%s%s%s\n", indent.String(), prefix, name)
			// Collect child fields
			var children []*Field
			for path, field := range index {
				if len(path) > len(f.Path)+1 &&
					path[:len(f.Path)] == f.Path &&
					path[len(f.Path)] == '.' {
					rest := path[len(f.Path)+1:]
					if !containsDot(rest) {
						children = append(children, field)
					}
				}
			}
			printFields(children, index, depth+1)
		} else {
			fmt.Printf("%s%s%-20s %s\n", indent.String(), prefix, name, f.Type)
		}
	}
}

func containsDot(s string) bool {
	for _, c := range s {
		if c == '.' {
			return true
		}
	}
	return false
}
