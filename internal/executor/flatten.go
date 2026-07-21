package executor

import (
	"fmt"

	"github.com/alex-ermolaxe/qdata/internal/format"
	"github.com/iancoleman/orderedmap"
)

// Flatten flattens a nested object field into the top level
// If there are field name collisions, all fields from the nested object are prefixed with the object field name
func Flatten(records []format.Record, fieldName string) ([]format.Record, error) {
	if len(records) == 0 {
		return records, nil
	}

	result := make([]format.Record, len(records))

	for i, record := range records {
		flattened, err := flattenRecord(record, fieldName)
		if err != nil {
			return nil, err
		}
		result[i] = flattened
	}

	return result, nil
}

func flattenRecord(record format.Record, fieldName string) (format.Record, error) {
	// Get the nested object field
	nestedObj, exists := record.Get(fieldName)
	if !exists {
		// Field doesn't exist, return a copy of the record unchanged
		return copyRecord(record), nil
	}

	// Convert nested object to orderedmap
	var nestedMap *orderedmap.OrderedMap
	switch v := nestedObj.(type) {
	case *orderedmap.OrderedMap:
		nestedMap = v
	case orderedmap.OrderedMap:
		nestedMap = &v
	case map[string]interface{}:
		// Convert regular map to orderedmap
		nestedMap = orderedmap.New()
		for k, val := range v {
			nestedMap.Set(k, val)
		}
	default:
		return nil, fmt.Errorf("field %q is not an object", fieldName)
	}

	// Check for collisions: determine if any nested field names already exist at top level (excluding the nested field itself)
	hasCollisions := false
	for _, key := range nestedMap.Keys() {
		if val, exists := record.Get(key); exists && val != nil {
			hasCollisions = true
			break
		}
	}

	// Create a new record
	newRecord := orderedmap.New()

	// Copy over all existing fields except the nested object field
	for _, key := range record.Keys() {
		if key != fieldName {
			val, _ := record.Get(key)
			if nested, ok := val.(*orderedmap.OrderedMap); ok {
				newRecord.Set(key, copyRecord(nested))
			} else {
				newRecord.Set(key, val)
			}
		}
	}

	// Add the flattened fields
	if hasCollisions {
		// Prefix all nested fields with the object field name
		for _, key := range nestedMap.Keys() {
			val, _ := nestedMap.Get(key)
			prefixedKey := fieldName + "_" + key
			newRecord.Set(prefixedKey, val)
		}
	} else {
		// No collisions, add fields directly at top level
		for _, key := range nestedMap.Keys() {
			val, _ := nestedMap.Get(key)
			newRecord.Set(key, val)
		}
	}

	return newRecord, nil
}
