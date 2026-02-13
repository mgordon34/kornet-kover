package main

import (
	"testing"

	"github.com/mgordon34/kornet-kover/api/players"
)

func TestConvertPlayerMaptoPlayerRosters(t *testing.T) {
	in := []players.Player{{Index: "a"}, {Index: "b"}}
	out := convertPlayerMaptoPlayerRosters(in)

	if len(out) != 2 {
		t.Fatalf("len = %d, want 2", len(out))
	}
	if out[0].PlayerIndex != "a" || out[0].Status != "Available" {
		t.Fatalf("unexpected first roster: %+v", out[0])
	}
}
