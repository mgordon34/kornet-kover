package sports

type MLBConfig struct {
    SportConfig
}

func NewMLB() *MLBConfig {
    return &MLBConfig{
        SportConfig: SportConfig{
            sport: MLB,
            sportsbookConfig: &SportsbookConfig{
                Markets: map[string]string{
                    "player_hits": "hits",
                    "player_strikeouts": "strikeouts",
                },
                MainlineConfig: MarketConfig{
                    Markets: []string{"player_hits", "player_strikeouts"},
                    Bookmaker: "williamhill_us",
                },
                AlternateConfig: MarketConfig{
                    Markets: []string{"player_hits_alternate", "player_strikeouts_alternate"},
                    Bookmaker: "fanduel",
                },
            },
            scraperConfig: &ScraperConfig{
                Domain: "https://www.baseball-reference.com",
                BoxScoreURL: "/boxes",
                StatMapping: map[string]string{
                "h": "hits",
                "so": "strikeouts",
                },
            },
            analysisConfig: &AnalysisConfig{
                DefaultStats: []string{"hits", "strikeouts", "runs"},
                StatWeights: map[string]float64{
                    "hits": 1.0,
                    "strikeouts": 1.2,
                },
            },
        },
    }
}