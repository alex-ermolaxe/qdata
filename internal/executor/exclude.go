package executor

import (
	"github.com/alex-ermolaxe/qdata/internal/format"
	"github.com/alex-ermolaxe/qdata/internal/schema"
)

// Exclude removes specified fields from each record
func Exclude(records []format.Record, fields []string) []format.Record {
	result := make([]format.Record, len(records))

	for i, record := range records {
		// Copy record to avoid mutating the original
		newRecord := copyRecord(record)

		for _, field := range fields {
			schema.DeleteNested(newRecord, field)
		}

		result[i] = newRecord
	}

	return result
}

// copyRecord creates a deep copy of a record
func copyRecord(record format.Record) format.Record {
	newRecord := format.Record{}

	for k, v := range record {
		if nested, ok := v.(map[string]any); ok {
			newRecord[k] = copyRecord(nested)
		} else {
			newRecord[k] = v
		}
	}

	return newRecord
}
