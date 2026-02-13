//go:build integration
// +build integration

package teams

import (
	"os"
	"testing"

	"github.com/mgordon34/kornet-kover/internal/storage"
)

func TestAddTeamsAndGetTeams(t *testing.T) {
	if os.Getenv("DB_URL") == "" {
		t.Skip("DB_URL not set; skipping integration test")
	}
	storage.InitTables()

	AddTeams([]Team{
		{Index: "ITM1", Name: "Integration Team 1"},
		{Index: "ITM2", Name: "Integration Team 2"},
	})

	out, err := GetTeams()
	if err != nil {
		t.Fatalf("GetTeams() error = %v", err)
	}

	found := map[string]bool{}
	for _, tm := range out {
		if tm.Index == "ITM1" || tm.Index == "ITM2" {
			found[tm.Index] = true
		}
	}

	if !found["ITM1"] || !found["ITM2"] {
		t.Fatalf("expected ITM1 and ITM2 in teams result")
	}
}
