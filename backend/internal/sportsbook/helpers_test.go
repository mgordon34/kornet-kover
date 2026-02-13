package sportsbook

import "testing"

func TestGetMarketType(t *testing.T) {
	if got := getMarketType("player_points_alternate"); got != "alternate" {
		t.Fatalf("getMarketType alternate = %q", got)
	}
	if got := getMarketType("player_points"); got != "mainline" {
		t.Fatalf("getMarketType mainline = %q", got)
	}
}

func TestParseNameFromDescription(t *testing.T) {
	if got := parseNameFromDescription("Aaron Gordon (Rebounds)"); got != "Aaron Gordon" {
		t.Fatalf("parseNameFromDescription() = %q", got)
	}
}
