package sports

import "testing"

func TestSportModelFields(t *testing.T) {
	s := Sport{Id: 1, Name: "nba"}
	if s.Id != 1 || s.Name != "nba" {
		t.Fatalf("unexpected sport model: %+v", s)
	}
}
