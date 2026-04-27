package executor

import (
	"github.com/alex-ermolaxe/qdata/internal/format"
	"github.com/alex-ermolaxe/qdata/internal/schema"
)

// Exclude удаляет указанные поля из каждой записи
func Exclude(records []format.Record, fields []string) []format.Record {
	result := make([]format.Record, len(records))

	for i, record := range records {
		// Копируем запись чтобы не мутировать оригинал
		newRecord := copyRecord(record)

		for _, field := range fields {
			schema.DeleteNested(newRecord, field)
		}

		result[i] = newRecord
	}

	return result
}

// copyRecord создаёт глубокую копию записи
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
