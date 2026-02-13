package sports

type Sport string

const (
    NBA Sport = "nba"
    WNBA Sport = "wnba"
    MLB Sport = "mlb"
    NHL Sport = "nhl"
)

// Config defines the interface that all sport configurations must implement
type Config interface {
    GetSport() Sport
    GetSportsbookConfig() *SportsbookConfig
    GetScraperConfig() *ScraperConfig
    GetAnalysisConfig() *AnalysisConfig
}

type SportsbookConfig struct {
    LeagueName       string
    Markets          map[string]MarketConfig
    StatMapping      map[string]string
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

type SportConfig struct {
    sport            Sport
    sportsbookConfig *SportsbookConfig
    scraperConfig    *ScraperConfig
    analysisConfig   *AnalysisConfig
}

func (c *SportConfig) GetSport() Sport {
    return c.sport
}

func (c *SportConfig) GetSportsbookConfig() *SportsbookConfig {
    return c.sportsbookConfig
}

func (c *SportConfig) GetScraperConfig() *ScraperConfig {
    return c.scraperConfig
}

func (c *SportConfig) GetAnalysisConfig() *AnalysisConfig {
    return c.analysisConfig
}
