package sports

import "testing"

func TestDefaultSportConfigs_HasExpectedEntries(t *testing.T) {
	tests := []struct {
		name  string
		sport Sport
	}{
		{name: "NBA", sport: NBA},
		{name: "WNBA", sport: WNBA},
		{name: "MLB", sport: MLB},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, ok := defaultSportConfigs[tt.sport]
			if !ok {
				t.Fatalf("defaultSportConfigs missing %s", tt.sport)
			}
			if cfg.Sportsbook == nil {
				t.Fatalf("Sportsbook config is nil")
			}
			if cfg.Scraper == nil {
				t.Fatalf("Scraper config is nil")
			}
			if cfg.Analysis == nil {
				t.Fatalf("Analysis config is nil")
			}
		})
	}
}
