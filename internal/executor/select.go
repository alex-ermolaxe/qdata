package executor

import (
	"github.com/alex-ermolaxe/qdata/internal/format"
	"github.com/alex-ermolaxe/qdata/internal/schema"
)

// Select оставляет в каждой записи только указанные поля
func Select(records []format.Record, fields []string) []format.Record {
	// SELECT * — возвращаем всё как есть
	if len(fields) == 1 && fields[0] == "*" {
		return records
	}

	result := make([]format.Record, len(records))

	for i, record := range records {
		newRecord := format.Record{}

		for _, field := range fields {
			val, exists := schema.GetNested(record, field)
			if exists {
				schema.SetNested(newRecord, field, val)
			}
		}

		result[i] = newRecord
	}

	return result
}
