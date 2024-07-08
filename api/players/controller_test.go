package players

import (
	"testing"

	"github.com/mgordon34/kornet-kover/internal/storage"
)

func TestPlayerNameToIndex(t *testing.T) {
    storage.InitDB()

    playerName := "Aaron Gordon"
    want := "gordoaa01"
    index, err := PlayerNameToIndex(playerName)
    if err != nil {
        t.Fatalf(`PlayerNameToIndex resulted in err: %v`, err)
    }
    if index != want {
        t.Fatalf(`PlayerNameToIndex = %s, want match for %s`, index, want)
    }
}
