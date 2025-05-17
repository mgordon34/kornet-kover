package sports

type MLBConfig struct {
    SportConfig
}

func NewMLB() *MLBConfig {
    return &MLBConfig{
        SportConfig: SportConfig{
            sport: MLB,
            sportsbookConfig: &SportsbookConfig{
                StatMapping: map[string]string{
                    "batter_home_runs": "home_runs",
                    "batter_hits": "hits",
                    "batter_rbis": "rbis",
                },
                LeagueName: "baseball_mlb",
                Markets: map[string]MarketConfig{
                    "mainline": {
                        Markets: []string{"batter_home_runs", "batter_hits", "batter_rbis"},
                        Bookmaker: "draftkings",
                    },
                    "alternate": {
                        Markets: []string{"batter_home_runs_alternate", "batter_hits_alternate", "batter_rbis_alternate"},
                        Bookmaker: "draftkings",
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
