package executor

import (
	"fmt"
	"reflect"

	"github.com/alex-ermolaxe/qdata/internal/format"
	"github.com/alex-ermolaxe/qdata/internal/schema"
	"github.com/iancoleman/orderedmap"
)

// TransformToMap creates one record per input record, using the value at
// keyPath as its key and the value at valuePath as its value.
func TransformToMap(records []format.Record, keyPath, valuePath string) ([]format.Record, error) {
	result := make([]format.Record, len(records))
	for i, record := range records {
		keyValue, exists := schema.GetNested(record, keyPath)
		if !exists {
			return nil, fmt.Errorf("record %d: key field %q does not exist", i+1, keyPath)
		}
		if keyValue == nil || !isScalar(keyValue) {
			return nil, fmt.Errorf("record %d: key field %q must contain a non-null scalar", i+1, keyPath)
		}

		value, exists := schema.GetNestedValues(record, valuePath)
		if !exists {
			return nil, fmt.Errorf("record %d: value field %q does not exist", i+1, valuePath)
		}

		transformed := orderedmap.New()
		transformed.Set(fmt.Sprint(keyValue), value)
		result[i] = transformed
	}

	return result, nil
}

func isScalar(value any) bool {
	if value == nil {
		return false
	}
	kind := reflect.TypeOf(value).Kind()

	return kind == reflect.String ||
		kind == reflect.Bool ||
		(kind >= reflect.Int && kind <= reflect.Int64) ||
		(kind >= reflect.Uint && kind <= reflect.Uint64) ||
		kind == reflect.Float32 ||
		kind == reflect.Float64
}
