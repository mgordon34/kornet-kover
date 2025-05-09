package sports

type NBA struct {
    sportbook *SportsbookConfig
    scraper   *ScraperConfig
    analysis  *AnalysisConfig
}

func NewNBA() *NBA {
    return &NBA{
        sportbook: &SportsbookConfig{
            Markets: map[string]string{
                "player_points": "points",
                "player_rebounds": "rebounds",
                "player_assists": "assists",
                "player_threes": "threes",
            },
            MainlineConfig: MarketConfig{
                Markets: []string{"player_points", "player_rebounds"},
                Bookmaker: "williamhill_us",
            },
            AlternateConfig: MarketConfig{
                Markets: []string{"player_points_alternate", "player_rebounds_alternate"},
                Bookmaker: "fanduel",
            },
        },
        scraper: &ScraperConfig{
            Domain: "https://www.basketball-reference.com",
            BoxScoreURL: "/boxscores",
            StatMapping: map[string]string{
                "pts": "points",
                "ast": "assists",
            },
        },
        analysis: &AnalysisConfig{
            DefaultStats: []string{"points", "rebounds", "assists"},
            StatWeights: map[string]float64{
                "points": 1.0,
                "assists": 1.5,
            },
        },
    }
}

func (c *NBA) GetSportsbookConfig() *SportsbookConfig {
    return c.sportbook
}

func (c *NBA) GetScraperConfig() *ScraperConfig {
    return c.scraper
}

func (c *NBA) GetAnalysisConfig() *AnalysisConfig {
    return c.analysis
} 