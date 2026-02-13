package odds

import (
	"testing"
	"time"
)

func TestPlayerLineModelFields(t *testing.T) {
	now := time.Now()
	line := PlayerLine{Id: 1, Sport: "nba", PlayerIndex: "idx", Timestamp: now, Type: "mainline", Stat: "points", Side: "Over", Line: 20.5, Odds: -110, Link: "x"}
	if line.Id != 1 || line.Sport != "nba" || line.PlayerIndex != "idx" || line.Timestamp.IsZero() {
		t.Fatalf("unexpected line model: %+v", line)
	}
}
