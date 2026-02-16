package sports

var Configs = map[Sport]SportConfig{
	NBA: {
		Sportsbook: SportsbookConfig{
			StatMapping: map[string]string{
				"player_points":   "points",
				"player_rebounds": "rebounds",
				"player_assists":  "assists",
				"player_threes":   "threes",
			},
			LeagueName: "basketball_nba",
			Markets: map[string]MarketConfig{
				"mainline": {
					Markets:   []string{"player_points", "player_rebounds"},
					Bookmaker: "williamhill_us",
				},
				"alternate": {
					Markets:   []string{"player_points_alternate", "player_rebounds_alternate"},
					Bookmaker: "fanduel",
				},
			},
		},
		Scraper: ScraperConfig{
			Domain:      "https://www.basketball-reference.com",
			BoxScoreURL: "/boxscores",
			StatMapping: map[string]string{
				"pts": "points",
				"ast": "assists",
			},
		},
		Analysis: AnalysisConfig{
			DefaultStats: []string{"points", "rebounds", "assists"},
			StatWeights: map[string]float64{
				"points":  1.0,
				"assists": 1.5,
			},
		},
	},
	WNBA: {
		Sportsbook: SportsbookConfig{
			StatMapping: map[string]string{
				"player_points":   "points",
				"player_rebounds": "rebounds",
				"player_assists":  "assists",
				"player_threes":   "threes",
			},
			LeagueName: "basketball_wnba",
			Markets: map[string]MarketConfig{
				"mainline": {
					Markets:   []string{"player_points", "player_rebounds"},
					Bookmaker: "williamhill_us",
				},
				"alternate": {
					Markets:   []string{"player_points_alternate", "player_rebounds_alternate"},
					Bookmaker: "fanduel",
				},
			},
		},
		Scraper: ScraperConfig{
			Domain:      "https://www.basketball-reference.com",
			BoxScoreURL: "/wnba/boxscores",
			StatMapping: map[string]string{
				"pts": "points",
				"ast": "assists",
			},
		},
		Analysis: AnalysisConfig{
			DefaultStats: []string{"points", "rebounds", "assists"},
			StatWeights: map[string]float64{
				"points":  1.0,
				"assists": 1.5,
			},
		},
	},
	MLB: {
		Sportsbook: SportsbookConfig{
			StatMapping: map[string]string{
				"batter_home_runs": "home_runs",
				"batter_hits":      "hits",
				"batter_rbis":      "rbis",
			},
			LeagueName: "baseball_mlb",
			Markets: map[string]MarketConfig{
				"mainline": {
					Markets:   []string{"batter_home_runs", "batter_hits", "batter_rbis"},
					Bookmaker: "draftkings",
				},
				"alternate": {
					Markets:   []string{"batter_home_runs_alternate", "batter_hits_alternate", "batter_rbis_alternate"},
					Bookmaker: "draftkings",
				},
			},
		},
		Scraper: ScraperConfig{
			Domain:      "https://www.baseball-reference.com",
			BoxScoreURL: "/boxes",
			StatMapping: map[string]string{
				"h":  "hits",
				"so": "strikeouts",
			},
		},
		Analysis: AnalysisConfig{
			DefaultStats: []string{"hits", "strikeouts", "runs"},
			StatWeights: map[string]float64{
				"hits":       1.0,
				"strikeouts": 1.2,
			},
		},
	},
}
