package schema

import "strings"

// Split splits a field path into parts
// "address.city" -> ["address", "city"]
// "name" -> ["name"]
func Split(path string) []string {
	return strings.Split(path, ".")
}

// Join combines path parts into a path
// ["address", "city"] -> "address.city"
func Join(parts []string) string {
	return strings.Join(parts, ".")
}

// GetNested extracts a value from a nested map by path
// path: "address.city", record: {"address": {"city": "Moscow"}} -> "Moscow", true
func GetNested(record map[string]any, path string) (any, bool) {
	parts := Split(path)
	current := record

	for i, part := range parts {
		val, ok := current[part]
		if !ok {
			return nil, false
		}

		// If this is the last part of the path - return the value
		if i == len(parts)-1 {
			return val, true
		}

		// Otherwise try to descend deeper
		nested, ok := val.(map[string]any)
		if !ok {
			return nil, false
		}
		current = nested
	}

	return nil, false
}

// SetNested sets a value in a nested map by path
// Creates intermediate maps if they don't exist
func SetNested(record map[string]any, path string, value any) {
	parts := Split(path)
	current := record

	for i, part := range parts {
		if i == len(parts)-1 {
			current[part] = value
			return
		}

		nested, ok := current[part].(map[string]any)
		if !ok {
			nested = map[string]any{}
			current[part] = nested
		}
		current = nested
	}
}

// DeleteNested deletes a field from a nested map by path
func DeleteNested(record map[string]any, path string) {
	parts := Split(path)
	current := record

	for i, part := range parts {
		if i == len(parts)-1 {
			delete(current, part)
			return
		}

		nested, ok := current[part].(map[string]any)
		if !ok {
			return
		}
		current = nested
	}
}
