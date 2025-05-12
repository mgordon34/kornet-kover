package sports

type NBAConfig struct {
    SportConfig
}

func NewNBA() *NBAConfig {
    return &NBAConfig{
        SportConfig: SportConfig{
            sport: NBA,
            sportsbookConfig: &SportsbookConfig{
                Markets: map[string]MarketConfig{
                    "mainline": {
                        Markets: []string{"player_points", "player_rebounds"},
                        Bookmaker: "williamhill_us",
                    },
                    "alternate": {
                        Markets: []string{"player_points_alternate", "player_rebounds_alternate"},
                        Bookmaker: "fanduel",
                    },
                },
            },
            scraperConfig: &ScraperConfig{
                Domain: "https://www.basketball-reference.com",
                BoxScoreURL: "/boxscores",
                StatMapping: map[string]string{
                    "pts": "points",
                    "ast": "assists",
                },
            },
            analysisConfig: &AnalysisConfig{
                DefaultStats: []string{"points", "rebounds", "assists"},
                StatWeights: map[string]float64{
                    "points": 1.0,
                    "assists": 1.5,
                },
            },
        },
    }
}