package sports

import (
	"errors"
	"testing"
)

func TestGetConfig_SupportedSport(t *testing.T) {
	config, err := GetConfig(NBA)
	if err != nil {
		t.Fatalf("GetConfig(NBA) err = %v", err)
	}
	if config.Scraper.Domain == "" {
		t.Fatalf("expected scraper domain")
	}
	if config.Sportsbook.LeagueName == "" {
		t.Fatalf("expected sportsbook league")
	}
}

func TestGetConfig_UnsupportedSport(t *testing.T) {
	_, err := GetConfig(NHL)
	if !errors.Is(err, ErrUnsupportedSport) {
		t.Fatalf("expected ErrUnsupportedSport, got %v", err)
	}
}
