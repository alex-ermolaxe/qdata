package executor

import (
	"encoding/json"
	"fmt"

	"github.com/alex-ermolaxe/qdata/internal/format"
)

// Show выводит записи в терминал
func Show(records []format.Record, limit int, offset int) error {
	total := len(records)

	// Применяем offset
	if offset >= total {
		fmt.Println("[]")
		fmt.Printf("── 0 of %d records ──\n", total)
		return nil
	}
	records = records[offset:]

	// Применяем limit
	shown := len(records)
	if limit > 0 && limit < shown {
		records = records[:limit]
		shown = limit
	}

	// Сериализуем в JSON с отступами
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize records: %w", err)
	}

	fmt.Println(string(data))
	fmt.Printf("── %d of %d records ──\n", shown, total)

	return nil
}
