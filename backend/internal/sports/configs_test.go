package sports

import "testing"

func TestConfigs_HasExpectedEntries(t *testing.T) {
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
			cfg, ok := Configs[tt.sport]
			if !ok {
				t.Fatalf("Configs missing %s", tt.sport)
			}
			if cfg.Sportsbook.LeagueName == "" {
				t.Fatalf("Sportsbook config is empty")
			}
			if cfg.Scraper.Domain == "" {
				t.Fatalf("Scraper config is empty")
			}
			if len(cfg.Analysis.DefaultStats) == 0 {
				t.Fatalf("Analysis config is empty")
			}
		})
	}
}
