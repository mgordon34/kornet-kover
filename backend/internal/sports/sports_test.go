package sports

import (
	"errors"
	"testing"
)

func TestDefaultConfigProvider_SupportedSports(t *testing.T) {
	provider := DefaultConfigProvider()

	sportsToCheck := []Sport{NBA, WNBA, MLB}
	for _, sport := range sportsToCheck {
		t.Run(string(sport), func(t *testing.T) {
			sportsbookConfig, err := provider.SportsbookConfig(sport)
			if err != nil {
				t.Fatalf("SportsbookConfig(%s) err = %v", sport, err)
			}
			if sportsbookConfig == nil {
				t.Fatalf("SportsbookConfig(%s) returned nil", sport)
			}

			scraperConfig, err := provider.ScraperConfig(sport)
			if err != nil {
				t.Fatalf("ScraperConfig(%s) err = %v", sport, err)
			}
			if scraperConfig == nil {
				t.Fatalf("ScraperConfig(%s) returned nil", sport)
			}

			analysisConfig, err := provider.AnalysisConfig(sport)
			if err != nil {
				t.Fatalf("AnalysisConfig(%s) err = %v", sport, err)
			}
			if analysisConfig == nil {
				t.Fatalf("AnalysisConfig(%s) returned nil", sport)
			}
		})
	}
}

func TestDefaultConfigProvider_UnsupportedSport(t *testing.T) {
	provider := DefaultConfigProvider()

	_, err := provider.SportsbookConfig(NHL)
	if !errors.Is(err, ErrUnsupportedSport) {
		t.Fatalf("expected ErrUnsupportedSport, got %v", err)
	}
}

func TestNewConfigProvider_UsesInjectedMap(t *testing.T) {
	provider := NewConfigProvider(map[Sport]SportConfig{
		NBA: {
			Sportsbook: &SportsbookConfig{LeagueName: "basketball_test"},
			Scraper:    &ScraperConfig{Domain: "https://example.com"},
			Analysis:   &AnalysisConfig{DefaultStats: []string{"points"}},
		},
	})

	gotSportsbook, err := provider.SportsbookConfig(NBA)
	if err != nil {
		t.Fatalf("SportsbookConfig(NBA) err = %v", err)
	}
	if gotSportsbook.LeagueName != "basketball_test" {
		t.Fatalf("LeagueName = %s, want basketball_test", gotSportsbook.LeagueName)
	}

	_, err = provider.ScraperConfig(MLB)
	if !errors.Is(err, ErrUnsupportedSport) {
		t.Fatalf("expected ErrUnsupportedSport for missing sport, got %v", err)
	}
}
