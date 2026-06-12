package executor

import (
	"github.com/alex-ermolaxe/qdata/internal/format"
	"github.com/alex-ermolaxe/qdata/internal/schema"
	"github.com/iancoleman/orderedmap"
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
	newRecord := format.NewRecord()

	for _, key := range record.Keys() {
		val, _ := record.Get(key)

		if nested, ok := val.(*orderedmap.OrderedMap); ok {
			newRecord.Set(key, copyRecord(nested))
		} else {
			newRecord.Set(key, val)
		}
	}

	return newRecord
}
