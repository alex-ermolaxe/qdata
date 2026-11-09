package schema

import (
	"strings"

	"github.com/alex-ermolaxe/qdata/internal/format"
	"github.com/iancoleman/orderedmap"
)

func Split(path string) []string {
	return strings.Split(path, ".")
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

// GetNestedValues extracts a path from an object, traversing arrays and
// returning the matching value from each element in an array.
func GetNestedValues(record format.Record, path string) (any, bool) {
	parts := Split(path)

	return getNestedValues(record, parts)
}

func getNestedValues(value any, parts []string) (any, bool) {
	if len(parts) == 0 {
		return value, true
	}

	if items, ok := value.([]any); ok {
		values := make([]any, 0, len(items))
		for _, item := range items {
			itemValue, exists := getNestedValues(item, parts)
			if exists {
				values = append(values, itemValue)
			}
		}

		return values, true
	}

	object, ok := asOrderedMap(value)
	if !ok {
		return nil, false
	}
	_, child, exists := getObjectField(object, parts[0])
	if !exists {
		return nil, false
	}

	return getNestedValues(child, parts[1:])
}

// SelectNested copies the fields at path from source into destination.
// When the path crosses an array, it projects each object in that array.
func SelectNested(source, destination format.Record, path string) bool {
	parts := Split(path)

	return selectObject(source, destination, parts)
}

func selectObject(source, destination any, parts []string) bool {
	if len(parts) == 0 {
		return false
	}

	sourceMap, ok := asOrderedMap(source)
	if !ok {
		return false
	}
	targetMap, ok := asOrderedMap(destination)
	if !ok {
		return false
	}

	key, value, exists := getObjectField(sourceMap, parts[0])
	if !exists {
		return false
	}
	if len(parts) == 1 {
		targetMap.Set(key, value)

		return true
	}

	targetValue, _ := targetMap.Get(key)
	switch value.(type) {
	case *orderedmap.OrderedMap, orderedmap.OrderedMap, map[string]any:
		targetChild, ok := asOrderedMap(targetValue)
		if !ok {
			targetChild = orderedmap.New()
		}
		selectObject(value, targetChild, parts[1:])
		targetMap.Set(key, targetChild)
	case []any:
		if selected, ok := selectArray(value, targetValue, parts[1:]); ok {
			targetMap.Set(key, selected)
		}
	default:
		return false
	}

	return true
}

func selectArray(source, destination any, parts []string) ([]any, bool) {
	items, ok := source.([]any)
	if !ok {
		return nil, false
	}
	selected, ok := destination.([]any)
	if !ok || len(selected) != len(items) {
		selected = make([]any, len(items))
	}

	for i, item := range items {
		switch item.(type) {
		case *orderedmap.OrderedMap, orderedmap.OrderedMap, map[string]any:
			target := orderedmap.New()
			if existing, exists := asOrderedMap(selected[i]); exists {
				target = existing
			}
			selectObject(item, target, parts)
			selected[i] = target
		default:
			selected[i] = item
		}
	}

	return selected, true
}

func asOrderedMap(value any) (*orderedmap.OrderedMap, bool) {
	switch object := value.(type) {
	case *orderedmap.OrderedMap:
		if object == nil {
			return nil, false
		}

		return object, true
	case orderedmap.OrderedMap:
		return &object, true
	case map[string]any:
		result := orderedmap.New()
		for key, value := range object {
			result.Set(key, value)
		}

		return result, true
	default:
		return nil, false
	}
}

func getObjectField(object *orderedmap.OrderedMap, key string) (string, any, bool) {
	if value, ok := object.Get(key); ok {
		return key, value, true
	}
	upperKey := strings.ToUpper(key)
	for _, actualKey := range object.Keys() {
		if strings.ToUpper(actualKey) == upperKey {
			value, ok := object.Get(actualKey)

			return actualKey, value, ok
		}
	}

	return "", nil, false
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
