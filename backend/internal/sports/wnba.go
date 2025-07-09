package sports

type WNBAConfig struct {
    SportConfig
}

func NewWNBA() *WNBAConfig {
    return &WNBAConfig{
        SportConfig: SportConfig{
            sport: WNBA,
            sportsbookConfig: &SportsbookConfig{
                StatMapping: map[string]string{
                    "player_points": "points",
                    "player_rebounds": "rebounds",
                    "player_assists": "assists",
                    "player_threes": "threes",
                },
                LeagueName: "basketball_wnba",
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
                BoxScoreURL: "/wnba/boxscores",
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
