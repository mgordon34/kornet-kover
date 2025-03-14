package strategies

type Strategy struct {
    Id              int         `json:"id"`
    UserId          int         `json:"user_id"`
    Name            string      `json:"name"`
}

type StrategyFilter struct {
    Id              int         `json:"id"`
    StrategyId      int         `json:"strategy_id"`
    Function        string      `json:"function"`
    Operator        string      `json:"operator"`
    Threshold       int         `json:"value"`
}
