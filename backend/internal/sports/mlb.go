package sports

type MLBConfig struct {
    SportConfig
}

func NewMLB() *MLBConfig {
    return &MLBConfig{
        SportConfig: SportConfig{
            sport: MLB,
            sportsbookConfig: &SportsbookConfig{
                Markets: map[string]MarketConfig{
                    "mainline": {
                        Markets: []string{"player_hits", "player_strikeouts"},
                        Bookmaker: "williamhill_us",
                    },
                    "alternate": {
                        Markets: []string{"player_hits_alternate", "player_strikeouts_alternate"},
                        Bookmaker: "fanduel",
                    },
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