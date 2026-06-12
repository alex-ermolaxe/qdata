package schema

import (
	"strings"

	"github.com/alex-ermolaxe/qdata/internal/format"
	"github.com/iancoleman/orderedmap"
)

func Split(path string) []string {
	return strings.Split(path, ".")
}

func Join(parts []string) string {
	return strings.Join(parts, ".")
}

// GetNested extracts a value from a record by path
func GetNested(record format.Record, path string) (any, bool) {
	parts := Split(path)
	var current any = record

	for i, part := range parts {
		switch c := current.(type) {
		case *orderedmap.OrderedMap:
			val, ok := c.Get(part)
			if !ok {
				val, ok = findKeyCaseInsensitive(c, part)
				if !ok {
					return nil, false
				}
			}
			if i == len(parts)-1 {
				return val, true
			}
			current = val
		case orderedmap.OrderedMap: // без pointer
			val, ok := c.Get(part)
			if !ok {
				val, ok = findKeyCaseInsensitive(&c, part)
				if !ok {
					return nil, false
				}
			}
			if i == len(parts)-1 {
				return val, true
			}
			current = val
		default:
			return nil, false
		}
	}

	return nil, false
}

// findKeyCaseInsensitive searches for a key in OrderedMap case-insensitively
func findKeyCaseInsensitive(m *orderedmap.OrderedMap, key string) (any, bool) {
	upperKey := strings.ToUpper(key)
	for _, k := range m.Keys() {
		if strings.ToUpper(k) == upperKey {
			val, ok := m.Get(k)
			return val, ok
		}
	}
	return nil, false
}

// SetNested sets a value in a record by path
func SetNested(record format.Record, path string, value any) {
	parts := Split(path)
	var current any = record

	for i, part := range parts {
		if i == len(parts)-1 {
			if c, ok := current.(*orderedmap.OrderedMap); ok {
				c.Set(part, value)
			}
			return
		}

		if c, ok := current.(*orderedmap.OrderedMap); ok {
			val, exists := c.Get(part)
			if !exists {
				nested := orderedmap.New()
				c.Set(part, nested)
				current = nested
			} else {
				current = val
			}
		}
	}
}

// DeleteNested deletes a field from a record by path
func DeleteNested(record format.Record, path string) {
	parts := Split(path)
	var current any = record

	for i, part := range parts {
		if i == len(parts)-1 {
			if c, ok := current.(*orderedmap.OrderedMap); ok {
				c.Delete(part)
			}
			return
		}

		if c, ok := current.(*orderedmap.OrderedMap); ok {
			val, exists := c.Get(part)
			if !exists {
				return
			}
			current = val
		}
	}
}
