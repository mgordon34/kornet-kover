package picks

import "time"

type PropPick struct {
    Id              int         `json:"id"`
    StratId         int         `json:"strat_id"`
    LineId          int         `json:"line_id"`
    Valid           bool        `json:"valid"`
    Date            time.Time   `json:"date"`
}
