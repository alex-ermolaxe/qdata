package parser

// Operator - comparison operator
type Operator string

const (
	OpEqual          Operator = "="
	OpNotEqual       Operator = "!="
	OpGreater        Operator = ">"
	OpLess           Operator = "<"
	OpGreaterOrEqual Operator = ">="
	OpLessOrEqual    Operator = "<="
	OpContains       Operator = "~"
	OpNotContains    Operator = "!~"
	OpStartsWith     Operator = "^"
	OpEndsWith       Operator = "$"
	OpIn             Operator = "IN"
	OpExists         Operator = "EXISTS"
)

// LogicalOp - logical operator between conditions
type LogicalOp string

const (
	LogicalAnd LogicalOp = "AND"
	LogicalOr  LogicalOp = "OR"
)

// Condition - a single filter condition
// Example: age > 30
type Condition struct {
	Field    string
	Operator Operator
	Value    any // string, float64, []any
}

// ConditionGroup - a group of conditions linked by a logical operator
type ConditionGroup struct {
	Conditions []Condition
	Operators  []LogicalOp // operators between conditions, len = len(Conditions) - 1
}
