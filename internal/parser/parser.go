package parser

import (
	"fmt"
	"strconv"
	"strings"
)

// Parse parses a condition string into a ConditionGroup
// Example: "age > 30 AND status = \"active\""
func Parse(input string) (*ConditionGroup, error) {
	group := &ConditionGroup{}

	// Split by AND/OR while preserving operators
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

// splitByLogical splits a string by AND/OR
func splitByLogical(input string) ([]string, []LogicalOp, error) {
	var parts []string
	var operators []LogicalOp

	// Tokenize input string
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

// parseCondition parses a single condition
// Example: age > 30, name ~ "john", status IN ["active", "pending"]
func parseCondition(input string) (*Condition, error) {
	tokens := tokenize(input)
	if len(tokens) < 2 {
		return nil, fmt.Errorf("invalid condition: %s", input)
	}

	field := tokens[0]

	// EXISTS - unary operator
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

	// IN - value is a list
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

// parseOperator parses an operator string
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

// parseValue parses a condition value
func parseValue(s string) (any, error) {
	// String in quotes
	if strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`) {
		return strings.Trim(s, `"`), nil
	}

	// Boolean value
	if strings.ToLower(s) == "true" {
		return true, nil
	}
	if strings.ToLower(s) == "false" {
		return false, nil
	}

	// Number
	if num, err := strconv.ParseFloat(s, 64); err == nil {
		return num, nil
	}

	// String without quotes
	return s, nil
}

// parseList parses a list of values ["a", "b", "c"]
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

// tokenize splits a string into tokens considering quoted strings and operators
func tokenize(input string) []string {
	var tokens []string
	var current strings.Builder
	inQuotes := false

	runes := []rune(input)
	i := 0

	for i < len(runes) {
		r := runes[i]

		// Handle quoted strings
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

		// Space - token separator
		if r == ' ' {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
			i++
			continue
		}

		// Two-character operators: !=, >=, <=, !~
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

		// Single-character operators: =, >, <, ~, ^, $
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
