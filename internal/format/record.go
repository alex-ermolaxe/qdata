package format

import "github.com/iancoleman/orderedmap"

// Rerepresents a single record in the dataset, implemented as an ordered map
type Record = *orderedmap.OrderedMap

// NewRecord creates a new empty record
// Using ordered map to preserve field order, which is important for formats like CSV
func NewRecord() Record {
	return orderedmap.New()
}
