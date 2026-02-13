package games

import (
	"testing"
	"time"
)

func TestGameModelFields(t *testing.T) {
	now := time.Now()
	g := Game{Id: 1, Sport: "nba", HomeIndex: "AAA", AwayIndex: "BBB", HomeScore: 100, AwayScore: 90, Date: now}
	if g.Id != 1 || g.Sport != "nba" || g.HomeIndex != "AAA" || g.Date.IsZero() {
		t.Fatalf("unexpected game model: %+v", g)
	}
}
