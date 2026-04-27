package executor

import (
	"fmt"
	"sort"
	"strings"

	"github.com/alex-ermolaxe/qdata/internal/format"
	"github.com/alex-ermolaxe/qdata/internal/schema"
)

// SortDirection — направление сортировки
type SortDirection string

const (
	SortAsc  SortDirection = "ASC"
	SortDesc SortDirection = "DESC"
)

// Sort сортирует записи по указанному полю
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

		// Записи без поля помещаем в конец
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

// lessValues сравнивает два значения для сортировки
func lessValues(a, b any) (bool, error) {
	// Числа
	an, aok := toFloat(a)
	bn, bok := toFloat(b)
	if aok && bok {
		return an < bn, nil
	}

	// Строки
	as, aok := a.(string)
	bs, bok := b.(string)
	if aok && bok {
		return strings.ToLower(as) < strings.ToLower(bs), nil
	}

	// Разные типы — не можем сравнить
	return false, fmt.Errorf(
		"cannot compare values of different types: %T and %T", a, b,
	)
}
