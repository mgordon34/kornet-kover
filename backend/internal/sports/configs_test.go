package sports

import "testing"

func TestSportSpecificConstructors(t *testing.T) {
	tests := []struct {
		name  string
		cfg   Config
		sport Sport
	}{
		{name: "NBA", cfg: NewNBA(), sport: NBA},
		{name: "WNBA", cfg: NewWNBA(), sport: WNBA},
		{name: "MLB", cfg: NewMLB(), sport: MLB},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cfg.GetSport() != tt.sport {
				t.Fatalf("GetSport() = %s, want %s", tt.cfg.GetSport(), tt.sport)
			}
			if tt.cfg.GetSportsbookConfig() == nil {
				t.Fatalf("GetSportsbookConfig() returned nil")
			}
			if tt.cfg.GetScraperConfig() == nil {
				t.Fatalf("GetScraperConfig() returned nil")
			}
			if tt.cfg.GetAnalysisConfig() == nil {
				t.Fatalf("GetAnalysisConfig() returned nil")
			}
		})
	}
}
