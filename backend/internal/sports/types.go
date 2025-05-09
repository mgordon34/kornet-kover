package sports

type Sport string

const (
    NBA Sport = "nba"
    MLB Sport = "mlb"
)

// Config defines the interface that all sport configurations must implement
type Config interface {
    GetSportsbookConfig() *SportsbookConfig
    GetScraperConfig() *ScraperConfig
    GetAnalysisConfig() *AnalysisConfig
}

// SportConfig holds basic sport configuration
type SportConfig struct {
    Domain      string
    BoxScoreURL string
    OddsName    string
}

// Map of basic sport configurations
var SportConfigs = map[Sport]SportConfig{
    NBA: {
        Domain:      "https://www.basketball-reference.com",
        BoxScoreURL: "/boxscores",
        OddsName:    "basketball_nba",
    },
    MLB: {
        Domain:      "https://www.baseball-reference.com",
        BoxScoreURL: "/boxes",
        OddsName:    "baseball_mlb",
    },
}

type SportsbookConfig struct {
    Markets          map[string]string
    MainlineConfig   MarketConfig
    AlternateConfig  MarketConfig
}

type MarketConfig struct {
    Markets    []string
    Bookmaker  string
}

type ScraperConfig struct {
    Domain      string
    BoxScoreURL string
    StatMapping map[string]string
}

type AnalysisConfig struct {
    DefaultStats []string
    StatWeights map[string]float64
} 