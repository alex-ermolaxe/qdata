package csv

import (
	gcsv "encoding/csv"
	"fmt"
	"io"
	"strconv"

	"github.com/alex-ermolaxe/qdata/internal/format"
)

// CSVFormat is an implementation of Format interface for CSV
type CSVFormat struct{}

func (c *CSVFormat) Decode(r io.Reader) ([]format.Record, error) {
	reader := gcsv.NewReader(r)

	// Read header
	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV headers: %w", err)
	}

	var records []format.Record

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV row: %w", err)
		}

		if len(row) != len(headers) {
			return nil, fmt.Errorf(
				"row has %d fields but header has %d fields",
				len(row), len(headers),
			)
		}

		record := format.Record{}
		for i, header := range headers {
			record[header] = parseValue(row[i])
		}

		records = append(records, record)
	}

	return records, nil
}

func (c *CSVFormat) Encode(w io.Writer, records []format.Record) error {
	if len(records) == 0 {
		return nil
	}

	writer := gcsv.NewWriter(w)
	defer writer.Flush()

	// Collect headers from the first record
	headers := make([]string, 0)
	for key := range records[0] {
		headers = append(headers, key)
	}

	// Write header
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write CSV headers: %w", err)
	}

	// Write rows
	for _, record := range records {
		row := make([]string, len(headers))
		for i, header := range headers {
			val, ok := record[header]
			if !ok {
				row[i] = ""
				continue
			}
			row[i] = valueToString(val)
		}

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	if err := writer.Error(); err != nil {
		return fmt.Errorf("CSV writer error: %w", err)
	}

	return nil
}

func (c *CSVFormat) Extensions() []string {
	return []string{"csv"}
}

// Register registers CSV format in the registry
func Register() {
	format.Register("csv", &CSVFormat{})
}

// parseValue attempts to determine the value type from a string
func parseValue(s string) any {
	// Empty string
	if s == "" {
		return nil
	}

	// Boolean value
	if s == "true" {
		return true
	}
	if s == "false" {
		return false
	}

	// Number
	if num, err := strconv.ParseFloat(s, 64); err == nil {
		return num
	}

	// String
	return s
}

// valueToString converts a value to a string for writing to CSV
func valueToString(val any) string {
	if val == nil {
		return ""
	}

	switch v := val.(type) {
	case float64:
		// Remove unnecessary zeros: 87.500000 -> 87.5
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}
