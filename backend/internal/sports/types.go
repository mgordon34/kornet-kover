package sports

import "fmt"

type Sport string

const (
	NBA  Sport = "nba"
	WNBA Sport = "wnba"
	MLB  Sport = "mlb"
	NHL  Sport = "nhl"
)

var ErrUnsupportedSport = fmt.Errorf("unsupported sport")

type SportsbookConfig struct {
	LeagueName  string
	Markets     map[string]MarketConfig
	StatMapping map[string]string
}

type MarketConfig struct {
	Markets   []string
	Bookmaker string
}

type ScraperConfig struct {
	Domain      string
	BoxScoreURL string
	StatMapping map[string]string
}

type AnalysisConfig struct {
	DefaultStats []string
	StatWeights  map[string]float64
}

type SportConfig struct {
	Sportsbook SportsbookConfig
	Scraper    ScraperConfig
	Analysis   AnalysisConfig
}
