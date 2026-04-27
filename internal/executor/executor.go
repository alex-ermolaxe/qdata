package executor

import (
	"encoding/json"
	"fmt"

	"github.com/alex-ermolaxe/qdata/internal/format"
)

// Show outputs records to terminal
func Show(records []format.Record, limit int, offset int) error {
	total := len(records)

	// Apply offset
	if offset >= total {
		fmt.Println("[]")
		fmt.Printf("── 0 of %d records ──\n", total)
		return nil
	}
	records = records[offset:]

	// Apply limit
	shown := len(records)
	if limit > 0 && limit < shown {
		records = records[:limit]
		shown = limit
	}

	// Serialize to indented JSON
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize records: %w", err)
	}

	fmt.Println(string(data))
	fmt.Printf("── %d of %d records ──\n", shown, total)

	return nil
}
