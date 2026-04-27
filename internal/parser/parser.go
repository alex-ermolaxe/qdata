package parser

import (
	"fmt"
	"strconv"
	"strings"
)

// Parse разбирает строку условия в ConditionGroup
// Например: "age > 30 AND status = "active""
func Parse(input string) (*ConditionGroup, error) {
	group := &ConditionGroup{}

	// Разбиваем по AND/OR сохраняя операторы
	parts, operators, err := splitByLogical(input)
	if err != nil {
		return nil, err
	}

	for _, part := range parts {
		condition, err := parseCondition(strings.TrimSpace(part))
		if err != nil {
			return nil, err
		}
		group.Conditions = append(group.Conditions, *condition)
	}

	group.Operators = operators

	return group, nil
}

// splitByLogical разбивает строку по AND/OR
func splitByLogical(input string) ([]string, []LogicalOp, error) {
	var parts []string
	var operators []LogicalOp

	// Токенизируем входную строку
	tokens := tokenize(input)
	current := []string{}

	for _, token := range tokens {
		upper := strings.ToUpper(token)
		if upper == "AND" || upper == "OR" {
			if len(current) == 0 {
				return nil, nil, fmt.Errorf("unexpected logical operator: %s", token)
			}
			parts = append(parts, strings.Join(current, " "))
			operators = append(operators, LogicalOp(upper))
			current = []string{}
		} else {
			current = append(current, token)
		}
	}

	if len(current) > 0 {
		parts = append(parts, strings.Join(current, " "))
	}

	return parts, operators, nil
}

// parseCondition разбирает одно условие
// Например: age > 30, name ~ "john", status IN ["active", "pending"]
func parseCondition(input string) (*Condition, error) {
	tokens := tokenize(input)
	if len(tokens) < 2 {
		return nil, fmt.Errorf("invalid condition: %s", input)
	}

	field := tokens[0]

	// EXISTS — унарный оператор
	if strings.ToUpper(tokens[1]) == "EXISTS" {
		return &Condition{
			Field:    field,
			Operator: OpExists,
		}, nil
	}

	if len(tokens) < 3 {
		return nil, fmt.Errorf("invalid condition: %s", input)
	}

	op, err := parseOperator(tokens[1])
	if err != nil {
		return nil, err
	}

	// IN — значение это список
	if op == OpIn {
		values, err := parseList(tokens[2:])
		if err != nil {
			return nil, err
		}
		return &Condition{
			Field:    field,
			Operator: op,
			Value:    values,
		}, nil
	}

	value, err := parseValue(tokens[2])
	if err != nil {
		return nil, err
	}

	return &Condition{
		Field:    field,
		Operator: op,
		Value:    value,
	}, nil
}

// parseOperator разбирает строку оператора
func parseOperator(s string) (Operator, error) {
	switch s {
	case "=":
		return OpEqual, nil
	case "!=":
		return OpNotEqual, nil
	case ">":
		return OpGreater, nil
	case "<":
		return OpLess, nil
	case ">=":
		return OpGreaterOrEqual, nil
	case "<=":
		return OpLessOrEqual, nil
	case "~":
		return OpContains, nil
	case "!~":
		return OpNotContains, nil
	case "^":
		return OpStartsWith, nil
	case "$":
		return OpEndsWith, nil
	case "IN", "in":
		return OpIn, nil
	default:
		return "", fmt.Errorf("unknown operator: %s", s)
	}
}

// parseValue разбирает значение условия
func parseValue(s string) (any, error) {
	// Строка в кавычках
	if strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`) {
		return strings.Trim(s, `"`), nil
	}

	// Булево значение
	if strings.ToLower(s) == "true" {
		return true, nil
	}
	if strings.ToLower(s) == "false" {
		return false, nil
	}

	// Число
	if num, err := strconv.ParseFloat(s, 64); err == nil {
		return num, nil
	}

	// Строка без кавычек
	return s, nil
}

// parseList разбирает список значений ["a", "b", "c"]
func parseList(tokens []string) ([]any, error) {
	joined := strings.Join(tokens, " ")
	joined = strings.TrimSpace(joined)

	if !strings.HasPrefix(joined, "[") || !strings.HasSuffix(joined, "]") {
		return nil, fmt.Errorf("invalid list format, expected [val1, val2, ...]: %s", joined)
	}

	inner := joined[1 : len(joined)-1]
	parts := strings.Split(inner, ",")

	var values []any
	for _, part := range parts {
		val, err := parseValue(strings.TrimSpace(part))
		if err != nil {
			return nil, err
		}
		values = append(values, val)
	}

	return values, nil
}

// tokenize разбивает строку на токены с учётом строк в кавычках и операторов
func tokenize(input string) []string {
	var tokens []string
	var current strings.Builder
	inQuotes := false

	runes := []rune(input)
	i := 0

	for i < len(runes) {
		r := runes[i]

		// Обрабатываем строки в кавычках
		if r == '"' {
			inQuotes = !inQuotes
			current.WriteRune(r)
			i++
			continue
		}

		if inQuotes {
			current.WriteRune(r)
			i++
			continue
		}

		// Пробел — разделитель токенов
		if r == ' ' {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
			i++
			continue
		}

		// Двухсимвольные операторы: !=, >=, <=, !~
		if i+1 < len(runes) {
			two := string(runes[i : i+2])
			if two == "!=" || two == ">=" || two == "<=" || two == "!~" {
				if current.Len() > 0 {
					tokens = append(tokens, current.String())
					current.Reset()
				}
				tokens = append(tokens, two)
				i += 2
				continue
			}
		}

		// Односимвольные операторы: =, >, <, ~, ^, $
		if r == '=' || r == '>' || r == '<' || r == '~' || r == '^' || r == '$' {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
			tokens = append(tokens, string(r))
			i++
			continue
		}

		current.WriteRune(r)
		i++
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens
}
