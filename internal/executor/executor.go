package executor

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alex-ermolaxe/qdata/internal/format"
)

// Show outputs records to terminal
func Show(records []format.Record, limit int, offset int) error {
	output, err := RenderShow(records, limit, offset)
	if err != nil {
		return err
	}
	fmt.Print(output)
	return nil
}

// RenderShow formats records for display without writing to stdout.
func RenderShow(records []format.Record, limit int, offset int) (string, error) {
	total := len(records)

	// Apply offset
	if offset >= total {
		return fmt.Sprintf("[]\n── 0 of %d records ──\n", total), nil
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
		return "", fmt.Errorf("failed to serialize records: %w", err)
	}

	var output strings.Builder
	fmt.Fprintln(&output, string(data))
	fmt.Fprintf(&output, "── %d of %d records ──\n", shown, total)
	return output.String(), nil
}
