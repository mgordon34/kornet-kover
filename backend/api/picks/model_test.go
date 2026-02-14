package picks

import (
	"testing"
	"time"
)

func TestPropPickModelFields(t *testing.T) {
	now := time.Now()
	p := PropPick{Id: 1, StratId: 2, LineId: 3, Valid: true, Date: now}
	if p.Id != 1 || p.StratId != 2 || p.LineId != 3 || !p.Valid || p.Date.IsZero() {
		t.Fatalf("unexpected prop pick model: %+v", p)
	}
}
