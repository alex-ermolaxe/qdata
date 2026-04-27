package parser

// Operator — оператор сравнения
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

// LogicalOp — логический оператор между условиями
type LogicalOp string

const (
	LogicalAnd LogicalOp = "AND"
	LogicalOr  LogicalOp = "OR"
)

// Condition — одно условие фильтрации
// Например: age > 30
type Condition struct {
	Field    string
	Operator Operator
	Value    any // string, float64, []any
}

// ConditionGroup — группа условий связанных логическим оператором
type ConditionGroup struct {
	Conditions []Condition
	Operators  []LogicalOp // операторы между условиями, len = len(Conditions) - 1
}
