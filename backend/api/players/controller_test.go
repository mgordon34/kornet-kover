package players

import (
	"testing"

	"github.com/mgordon34/kornet-kover/internal/storage"
)

func TestPlayerNameToIndex(t *testing.T) {
	storage.InitTables()

	nameMap := map[string]string{}
	playerName := "Aaron Gordon"
	want := "gordoaa01"
	index, err := PlayerNameToIndex(nameMap, playerName)
	if err != nil {
		t.Fatalf(`PlayerNameToIndex resulted in err: %v`, err)
	}
	if index != want {
		t.Fatalf(`PlayerNameToIndex = %s, want match for %s`, index, want)
	}
}

func TestPlayerNameToIndexWithBadName(t *testing.T) {
	storage.InitTables()

	nameMap := map[string]string{}
	badName := "Aaron Gordo"
	index, err := PlayerNameToIndex(nameMap, badName)
	if err == nil {
		t.Fatalf(`PlayerNameToIndex incorrectly found result: %v`, index)
	}
}
