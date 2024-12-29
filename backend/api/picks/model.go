package picks

type PropPick struct {
    Id              int         `json:"id"`
    UserId          int         `json:"user_id"`
    PlayerIndex     string      `json:"player_index"`
    Side            string      `json:"side"`
    Line            float32     `json:"line"`
    Stat            string      `json:"stat"`
    Valid           bool        `json:"valid"`
}
