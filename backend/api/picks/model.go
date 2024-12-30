package picks

import "time"

type PropPick struct {
    Id              int         `json:"id"`
    UserId          int         `json:"user_id"`
    LineId          int         `json:"line_id"`
    Valid           bool        `json:"valid"`
    Date            time.Time   `json:"date"`
}
