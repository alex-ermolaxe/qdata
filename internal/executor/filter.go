package executor

import (
	"fmt"
	"strings"

	"github.com/alex-ermolaxe/qdata/internal/format"
	"github.com/alex-ermolaxe/qdata/internal/parser"
	"github.com/alex-ermolaxe/qdata/internal/schema"
)

// Filter применяет группу условий к записям и возвращает отфильтрованные
func Filter(records []format.Record, group *parser.ConditionGroup) ([]format.Record, error) {
	var result []format.Record

	for _, record := range records {
		match, err := matchGroup(record, group)
		if err != nil {
			return nil, err
		}
		if match {
			result = append(result, record)
		}
	}

	return result, nil
}

// matchGroup проверяет соответствие записи группе условий
func matchGroup(record format.Record, group *parser.ConditionGroup) (bool, error) {
	if len(group.Conditions) == 0 {
		return true, nil
	}

	result, err := matchCondition(record, group.Conditions[0])
	if err != nil {
		return false, err
	}

	for i, op := range group.Operators {
		next, err := matchCondition(record, group.Conditions[i+1])
		if err != nil {
			return false, err
		}

		switch op {
		case parser.LogicalAnd:
			result = result && next
		case parser.LogicalOr:
			result = result || next
		}
	}

	return result, nil
}

// matchCondition проверяет соответствие записи одному условию
func matchCondition(record format.Record, condition parser.Condition) (bool, error) {
	// EXISTS — просто проверяем наличие поля
	if condition.Operator == parser.OpExists {
		_, exists := schema.GetNested(record, condition.Field)
		return exists, nil
	}

	// Получаем значение поля
	fieldVal, exists := schema.GetNested(record, condition.Field)
	if !exists {
		return false, nil
	}

	return compare(fieldVal, condition.Operator, condition.Value)
}

// compare сравнивает значение поля с условием
func compare(fieldVal any, op parser.Operator, condVal any) (bool, error) {
	switch op {
	case parser.OpEqual:
		return equalValues(fieldVal, condVal), nil

	case parser.OpNotEqual:
		return !equalValues(fieldVal, condVal), nil

	case parser.OpContains:
		fs, cs, ok := toStrings(fieldVal, condVal)
		if !ok {
			return false, nil
		}
		return strings.Contains(strings.ToLower(fs), strings.ToLower(cs)), nil

	case parser.OpNotContains:
		fs, cs, ok := toStrings(fieldVal, condVal)
		if !ok {
			return false, nil
		}
		return !strings.Contains(strings.ToLower(fs), strings.ToLower(cs)), nil

	case parser.OpStartsWith:
		fs, cs, ok := toStrings(fieldVal, condVal)
		if !ok {
			return false, nil
		}
		return strings.HasPrefix(strings.ToLower(fs), strings.ToLower(cs)), nil

	case parser.OpEndsWith:
		fs, cs, ok := toStrings(fieldVal, condVal)
		if !ok {
			return false, nil
		}
		return strings.HasSuffix(strings.ToLower(fs), strings.ToLower(cs)), nil

	case parser.OpGreater:
		fn, cn, ok := toNumbers(fieldVal, condVal)
		if !ok {
			return false, nil
		}
		return fn > cn, nil

	case parser.OpLess:
		fn, cn, ok := toNumbers(fieldVal, condVal)
		if !ok {
			return false, nil
		}
		return fn < cn, nil

	case parser.OpGreaterOrEqual:
		fn, cn, ok := toNumbers(fieldVal, condVal)
		if !ok {
			return false, nil
		}
		return fn >= cn, nil

	case parser.OpLessOrEqual:
		fn, cn, ok := toNumbers(fieldVal, condVal)
		if !ok {
			return false, nil
		}
		return fn <= cn, nil

	case parser.OpIn:
		list, ok := condVal.([]any)
		if !ok {
			return false, fmt.Errorf("IN operator requires a list")
		}
		for _, item := range list {
			if equalValues(fieldVal, item) {
				return true, nil
			}
		}
		return false, nil
	}

	return false, fmt.Errorf("unknown operator: %s", op)
}

// equalValues сравнивает два значения
func equalValues(a, b any) bool {
	// Числа
	an, aok := toFloat(a)
	bn, bok := toFloat(b)
	if aok && bok {
		return an == bn
	}

	// Строки
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

// toStrings приводит два значения к строкам
func toStrings(a, b any) (string, string, bool) {
	as, aok := a.(string)
	bs, bok := b.(string)
	return as, bs, aok && bok
}

// toNumbers приводит два значения к числам
func toNumbers(a, b any) (float64, float64, bool) {
	an, aok := toFloat(a)
	bn, bok := toFloat(b)
	return an, bn, aok && bok
}

// toFloat приводит значение к float64
func toFloat(v any) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	}
	return 0, false
}
