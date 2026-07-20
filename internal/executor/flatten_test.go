package executor

import (
	"testing"

	"github.com/alex-ermolaxe/qdata/internal/format"
	"github.com/iancoleman/orderedmap"
)

func TestFlatten(t *testing.T) {
	tests := []struct {
		name        string
		records     []format.Record
		fieldName   string
		wantErr     bool
		errContains string
		verify      func(t *testing.T, result []format.Record)
	}{
		{
			name: "flatten without collisions",
			records: []format.Record{
				createTestRecord(map[string]interface{}{
					"id":   1.0,
					"name": "Alice",
					"address": map[string]interface{}{
						"city":   "Moscow",
						"zip":    "101000",
						"street": "Arbat 12",
					},
				}),
			},
			fieldName: "address",
			wantErr:   false,
			verify: func(t *testing.T, result []format.Record) {
				if len(result) != 1 {
					t.Fatalf("expected 1 record, got %d", len(result))
				}

				record := result[0]

				// Check original fields are preserved
				assertFieldValue(t, record, "id", 1.0)
				assertFieldValue(t, record, "name", "Alice")

				// Check nested fields are flattened
				assertFieldValue(t, record, "city", "Moscow")
				assertFieldValue(t, record, "zip", "101000")
				assertFieldValue(t, record, "street", "Arbat 12")

				// Check nested field is removed
				if _, exists := record.Get("address"); exists {
					t.Error("expected 'address' field to be removed")
				}
			},
		},
		{
			name: "flatten with collisions - should use prefix",
			records: []format.Record{
				createTestRecord(map[string]interface{}{
					"id": 1.0,
					"city": "New York", // collision with work.city
					"work": map[string]interface{}{
						"city":   "Tbilisi",
						"zip":    "0178",
						"street": "Guramishvili ave 78",
					},
				}),
			},
			fieldName: "work",
			wantErr:   false,
			verify: func(t *testing.T, result []format.Record) {
				if len(result) != 1 {
					t.Fatalf("expected 1 record, got %d", len(result))
				}

				record := result[0]

				// Check original collision field is preserved
				assertFieldValue(t, record, "city", "New York")

				// Check nested fields have prefix
				assertFieldValue(t, record, "work_city", "Tbilisi")
				assertFieldValue(t, record, "work_zip", "0178")
				assertFieldValue(t, record, "work_street", "Guramishvili ave 78")

				// Check work field is removed
				if _, exists := record.Get("work"); exists {
					t.Error("expected 'work' field to be removed")
				}
			},
		},
		{
			name: "flatten nonexistent field - should return unchanged",
			records: []format.Record{
				createTestRecord(map[string]interface{}{
					"id":   1.0,
					"name": "Alice",
				}),
			},
			fieldName: "address",
			wantErr:   false,
			verify: func(t *testing.T, result []format.Record) {
				if len(result) != 1 {
					t.Fatalf("expected 1 record, got %d", len(result))
				}

				record := result[0]
				assertFieldValue(t, record, "id", 1.0)
				assertFieldValue(t, record, "name", "Alice")
			},
		},
		{
			name: "flatten non-object field - should error",
			records: []format.Record{
				createTestRecord(map[string]interface{}{
					"id":   1.0,
					"name": "Alice",
					"age":  30.0,
				}),
			},
			fieldName:   "age",
			wantErr:     true,
			errContains: "is not an object",
		},
		{
			name:      "flatten empty records",
			records:   []format.Record{},
			fieldName: "address",
			wantErr:   false,
			verify: func(t *testing.T, result []format.Record) {
				if len(result) != 0 {
					t.Fatalf("expected 0 records, got %d", len(result))
				}
			},
		},
		{
			name: "flatten multiple records",
			records: []format.Record{
				createTestRecord(map[string]interface{}{
					"id": 1.0,
					"address": map[string]interface{}{
						"city": "Moscow",
						"zip":  "101000",
					},
				}),
				createTestRecord(map[string]interface{}{
					"id": 2.0,
					"address": map[string]interface{}{
						"city": "Saint Petersburg",
						"zip":  "190000",
					},
				}),
				createTestRecord(map[string]interface{}{
					"id": 3.0,
					"address": map[string]interface{}{
						"city": "Kazan",
						"zip":  "420000",
					},
				}),
			},
			fieldName: "address",
			wantErr:   false,
			verify: func(t *testing.T, result []format.Record) {
				if len(result) != 3 {
					t.Fatalf("expected 3 records, got %d", len(result))
				}

				// Check first record
				assertFieldValue(t, result[0], "id", 1.0)
				assertFieldValue(t, result[0], "city", "Moscow")
				assertFieldValue(t, result[0], "zip", "101000")

				// Check second record
				assertFieldValue(t, result[1], "id", 2.0)
				assertFieldValue(t, result[1], "city", "Saint Petersburg")
				assertFieldValue(t, result[1], "zip", "190000")

				// Check third record
				assertFieldValue(t, result[2], "id", 3.0)
				assertFieldValue(t, result[2], "city", "Kazan")
				assertFieldValue(t, result[2], "zip", "420000")
			},
		},
		{
			name: "flatten partial collisions - should use prefix for all nested fields",
			records: []format.Record{
				createTestRecord(map[string]interface{}{
					"id": 1.0,
					"city": "New York", // collision
					"work": map[string]interface{}{
						"city":   "Tbilisi",
						"zip":    "0178",
						"street": "Guramishvili ave 78",
					},
				}),
			},
			fieldName: "work",
			wantErr:   false,
			verify: func(t *testing.T, result []format.Record) {
				record := result[0]

				// All nested fields should be prefixed
				assertFieldValue(t, record, "work_city", "Tbilisi")
				assertFieldValue(t, record, "work_zip", "0178")
				assertFieldValue(t, record, "work_street", "Guramishvili ave 78")

				// Original collision field preserved
				assertFieldValue(t, record, "city", "New York")
			},
		},
		{
			name: "flatten with nested object field",
			records: []format.Record{
				createTestRecord(map[string]interface{}{
					"id": 1.0,
					"user": map[string]interface{}{
						"name": "Alice",
						"contact": map[string]interface{}{
							"email": "alice@example.com",
						},
					},
				}),
			},
			fieldName: "user",
			wantErr:   false,
			verify: func(t *testing.T, result []format.Record) {
				record := result[0]

				// Nested object should be preserved as-is
				assertFieldValue(t, record, "name", "Alice")
				if _, exists := record.Get("contact"); !exists {
					t.Error("expected 'contact' field to exist")
				}
			},
		},
		{
			name: "flatten with array field",
			records: []format.Record{
				createTestRecord(map[string]interface{}{
					"id": 1.0,
					"tags": []interface{}{"admin", "user"},
					"metadata": map[string]interface{}{
						"created": "2024-01-01",
					},
				}),
			},
			fieldName: "metadata",
			wantErr:   false,
			verify: func(t *testing.T, result []format.Record) {
				record := result[0]

				// Original tags preserved
				tags, exists := record.Get("tags")
				if !exists {
					t.Error("expected 'tags' field to exist")
				}
				if tagList, ok := tags.([]interface{}); !ok || len(tagList) != 2 {
					t.Errorf("expected tags to be preserved correctly")
				}

				// Flattened metadata fields
				assertFieldValue(t, record, "created", "2024-01-01")
				if _, exists := record.Get("metadata"); exists {
					t.Error("expected 'metadata' field to be removed")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Flatten(tt.records, tt.fieldName)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.verify != nil {
				tt.verify(t, result)
			}
		})
	}
}

func TestFlattenRecord(t *testing.T) {
	t.Run("preserves original data", func(t *testing.T) {
		record := createTestRecord(map[string]interface{}{
			"id":   1.0,
			"name": "Alice",
			"address": map[string]interface{}{
				"city": "Moscow",
			},
		})

		flattened, err := flattenRecord(record, "address")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Original record should not be modified
		if _, exists := record.Get("address"); !exists {
			t.Error("original record should not be modified")
		}

		// Flattened record should not contain address
		if _, exists := flattened.Get("address"); exists {
			t.Error("flattened record should not contain address")
		}
	})

	t.Run("handles empty object", func(t *testing.T) {
		record := createTestRecord(map[string]interface{}{
			"id":      1.0,
			"address": map[string]interface{}{},
		})

		flattened, err := flattenRecord(record, "address")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should still work with empty object
		assertFieldValue(t, flattened, "id", 1.0)
	})

	t.Run("handles nil value in collision check", func(t *testing.T) {
		record := createTestRecord(map[string]interface{}{
			"id":      1.0,
			"city":    nil, // nil value should not be considered as collision
			"address": map[string]interface{}{
				"city": "Moscow",
				"zip":  "101000",
			},
		})

		flattened, err := flattenRecord(record, "address")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Since city is nil, no collision should be detected
		assertFieldValue(t, flattened, "city", "Moscow")
		assertFieldValue(t, flattened, "zip", "101000")
	})
}

func TestFlattenOrder(t *testing.T) {
	t.Run("preserves field order", func(t *testing.T) {
		record := createTestRecord(map[string]interface{}{
			"id":   1.0,
			"name": "Alice",
			"address": map[string]interface{}{
				"city":   "Moscow",
				"zip":    "101000",
				"street": "Arbat 12",
			},
			"email": "alice@example.com",
		})

		flattened, err := flattenRecord(record, "address")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Get all keys
		keys := flattened.Keys()

		// Check order: original fields first, then flattened fields
		if len(keys) < 5 {
			t.Errorf("expected at least 5 keys, got %d", len(keys))
		}

		// id and name should come before flattened address fields
		idIdx := findKeyIndex(keys, "id")
		nameIdx := findKeyIndex(keys, "name")
		cityIdx := findKeyIndex(keys, "city")

		if idIdx == -1 || nameIdx == -1 || cityIdx == -1 {
			t.Error("expected keys not found")
			return
		}

		if !(idIdx < cityIdx && nameIdx < cityIdx) {
			t.Error("expected original fields to appear before flattened fields")
		}
	})
}

// Benchmarks

func BenchmarkFlattenSingleRecord(b *testing.B) {
	record := createTestRecord(map[string]interface{}{
		"id":   1.0,
		"name": "Alice",
		"address": map[string]interface{}{
			"city":    "Moscow",
			"zip":     "101000",
			"street":  "Arbat 12",
			"country": "Russia",
		},
	})

	records := []format.Record{record}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Flatten(records, "address")
	}
}

func BenchmarkFlattenMultipleRecords(b *testing.B) {
	records := make([]format.Record, 100)
	for i := 0; i < 100; i++ {
		records[i] = createTestRecord(map[string]interface{}{
			"id":   float64(i),
			"name": "Person",
			"address": map[string]interface{}{
				"city":    "City",
				"zip":     "123456",
				"street":  "Street",
				"country": "Country",
			},
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Flatten(records, "address")
	}
}

func BenchmarkFlattenWithCollisions(b *testing.B) {
	records := make([]format.Record, 100)
	for i := 0; i < 100; i++ {
		records[i] = createTestRecord(map[string]interface{}{
			"id":   float64(i),
			"city": "New York", // collision
			"address": map[string]interface{}{
				"city":    "City",
				"zip":     "123456",
				"street":  "Street",
				"country": "Country",
			},
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Flatten(records, "address")
	}
}

// Helper functions

func createTestRecord(data map[string]interface{}) format.Record {
	record := orderedmap.New()
	for k, v := range data {
		if nestedMap, ok := v.(map[string]interface{}); ok {
			record.Set(k, convertToOrderedMap(nestedMap))
		} else {
			record.Set(k, v)
		}
	}
	return record
}

func convertToOrderedMap(data map[string]interface{}) *orderedmap.OrderedMap {
	om := orderedmap.New()
	for k, v := range data {
		if nestedMap, ok := v.(map[string]interface{}); ok {
			om.Set(k, convertToOrderedMap(nestedMap))
		} else {
			om.Set(k, v)
		}
	}
	return om
}

func assertFieldValue(t *testing.T, record format.Record, fieldName string, expectedValue interface{}) {
	t.Helper()
	value, exists := record.Get(fieldName)
	if !exists {
		t.Errorf("field %q not found", fieldName)
		return
	}
	if value != expectedValue {
		t.Errorf("field %q: expected %v, got %v", fieldName, expectedValue, value)
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func findKeyIndex(keys []string, key string) int {
	for i, k := range keys {
		if k == key {
			return i
		}
	}
	return -1
}
