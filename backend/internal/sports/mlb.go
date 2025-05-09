package sports

type MLB struct {
    sportbook *SportsbookConfig
    scraper   *ScraperConfig
    analysis  *AnalysisConfig
}

func NewMLB() *MLB {
    return &MLB{
        sportbook: &SportsbookConfig{
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
        scraper: &ScraperConfig{
            Domain: "https://www.baseball-reference.com",
            BoxScoreURL: "/boxes",
            StatMapping: map[string]string{
                "h": "hits",
                "so": "strikeouts",
            },
        },
        analysis: &AnalysisConfig{
            DefaultStats: []string{"hits", "strikeouts", "runs"},
            StatWeights: map[string]float64{
                "hits": 1.0,
                "strikeouts": 1.2,
            },
        },
    }
}

// Implement Config interface
func (c *MLB) GetSportsbookConfig() *SportsbookConfig {
    return c.sportbook
}

func (c *MLB) GetScraperConfig() *ScraperConfig {
    return c.scraper
}

func (c *MLB) GetAnalysisConfig() *AnalysisConfig {
    return c.analysis
} 