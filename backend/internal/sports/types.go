package sports

// Config defines the interface that all sport configurations must implement
type Config interface {
    GetSportsbookConfig() *SportsbookConfig
    GetScraperConfig() *ScraperConfig
    GetAnalysisConfig() *AnalysisConfig
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