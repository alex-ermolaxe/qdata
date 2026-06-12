package executor

import (
	"fmt"
	"sort"
	"strings"

	"github.com/alex-ermolaxe/qdata/internal/format"
	"github.com/alex-ermolaxe/qdata/internal/schema"
)

// SortDirection - sort direction
type SortDirection string

const (
	SortAsc  SortDirection = "ASC"
	SortDesc SortDirection = "DESC"
)

// Sort sorts records by the specified field
func Sort(records []format.Record, field string, direction SortDirection) ([]format.Record, error) {
	result := make([]format.Record, len(records))
	copy(result, records)

	var sortErr error

	sort.SliceStable(result, func(i, j int) bool {
		if sortErr != nil {
			return false
		}

		valI, existsI := schema.GetNested(result[i], field)
		valJ, existsJ := schema.GetNested(result[j], field)

		// Records without the field are placed at the end
		if !existsI && !existsJ {
			return false
		}
		if !existsI {
			return false
		}
		if !existsJ {
			return true
		}

		less, err := lessValues(valI, valJ)
		if err != nil {
			sortErr = err
			return false
		}

		if direction == SortDesc {
			return !less
		}
		return less
	})

	if sortErr != nil {
		return nil, fmt.Errorf("failed to sort records: %w", sortErr)
	}

	return result, nil
}

// lessValues compares two values for sorting
func lessValues(a, b any) (bool, error) {
	an, aok := toFloat(a)
	bn, bok := toFloat(b)
	if aok && bok {
		return an < bn, nil
	}

	as, aok := a.(string)
	bs, bok := b.(string)
	if aok && bok {
		return strings.ToLower(as) < strings.ToLower(bs), nil
	}

	return false, fmt.Errorf(
		"cannot compare values of different types: %T and %T", a, b,
	)
}
