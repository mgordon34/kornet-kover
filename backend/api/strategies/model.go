package strategies

type Strategy struct {
    Id              int         `json:"id"`
    UserId          int         `json:"user_id"`
    Name            string      `json:"name"`
}

type ComparisonType string

const (
    ValueComparison      ComparisonType = "value"
    FunctionComparison   ComparisonType = "function"
    ModifiedComparison   ComparisonType = "modified"
)

type ModifierOperator string

const (
    Multiply ModifierOperator = "*"
    Divide   ModifierOperator = "/"
    Add      ModifierOperator = "+"
    Subtract ModifierOperator = "-"
)

type ComparisonOperator string

const (
    GreaterThan     ComparisonOperator = ">"
    LessThan        ComparisonOperator = "<"
    GreaterOrEqual  ComparisonOperator = ">="
    LessOrEqual     ComparisonOperator = "<="
    Equal           ComparisonOperator = "=="
)

type StrategyFilter struct {
    Id                int                `json:"id"`
    StrategyId        int                `json:"strategy_id"`
    Function          string             `json:"function"`
    Stat              string             `json:"stat"`
    Operator          ComparisonOperator `json:"operator"`
    ComparisonType    ComparisonType     `json:"comparison_type"`
    CompareValue      *float64           `json:"compare_value,omitempty"`
    CompareFunction   *string            `json:"compare_function,omitempty"`
    CompareStat       *string            `json:"compare_stat,omitempty"`
    ModifierOperator  *ModifierOperator  `json:"modifier_operator,omitempty"`
}
