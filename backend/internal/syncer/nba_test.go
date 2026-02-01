package syncer

import (
	"reflect"
	"testing"
	"time"
)

func TestSameDate(t *testing.T) {
	a := time.Date(2025, 10, 12, 0, 0, 0, 0, time.UTC)
	b := time.Date(2025, 10, 12, 23, 59, 59, 0, time.UTC)
	c := time.Date(2025, 10, 13, 0, 0, 0, 0, time.UTC)

	if !sameDate(a, b) {
		t.Fatalf("expected same date")
	}
	if sameDate(a, c) {
		t.Fatalf("expected different dates")
	}
}

func TestUniquePlayerIndexes(t *testing.T) {
	pGames := []cloudPlayerGame{
		{PlayerIndex: "a"},
		{PlayerIndex: "b"},
		{PlayerIndex: "a"},
	}

	got := uniquePlayerIndexes(pGames)
	if len(got) != 2 {
		t.Fatalf("expected 2 unique players, got %d", len(got))
	}
}

func TestMapPlayerGames(t *testing.T) {
	pGames := []cloudPlayerGame{
		{PlayerIndex: "a", GameID: 10, TeamIndex: "t1", Points: 5},
		{PlayerIndex: "b", GameID: 20, TeamIndex: "t2", Points: 7},
	}

	mapping := map[int]int{10: 100}
	got := mapPlayerGames(pGames, mapping)

	if len(got) != 1 {
		t.Fatalf("expected 1 mapped player game, got %d", len(got))
	}
	if !reflect.DeepEqual(got[0].PlayerIndex, "a") || got[0].Game != 100 {
		t.Fatalf("unexpected mapped player game: %#v", got[0])
	}
}
