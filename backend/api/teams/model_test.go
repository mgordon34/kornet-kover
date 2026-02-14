package teams

import "testing"

func TestTeamQueryArgs(t *testing.T) {
	team := &Team{Index: "DEN", Name: "Denver Nuggets"}
	args := team.QueryArgs()
	if len(args) != 2 {
		t.Fatalf("QueryArgs len = %d, want 2", len(args))
	}
}
