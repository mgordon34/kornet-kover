package sports

import "fmt"

type staticConfigProvider struct {
	configs map[Sport]SportConfig
}

func NewConfigProvider(configs map[Sport]SportConfig) ConfigProvider {
	return staticConfigProvider{configs: configs}
}

func DefaultConfigProvider() ConfigProvider {
	return defaultConfigProvider
}

func (p staticConfigProvider) SportsbookConfig(sport Sport) (*SportsbookConfig, error) {
	config, err := p.getSportConfig(sport)
	if err != nil {
		return nil, err
	}
	return config.Sportsbook, nil
}

func (p staticConfigProvider) ScraperConfig(sport Sport) (*ScraperConfig, error) {
	config, err := p.getSportConfig(sport)
	if err != nil {
		return nil, err
	}
	return config.Scraper, nil
}

func (p staticConfigProvider) AnalysisConfig(sport Sport) (*AnalysisConfig, error) {
	config, err := p.getSportConfig(sport)
	if err != nil {
		return nil, err
	}
	return config.Analysis, nil
}

func (p staticConfigProvider) getSportConfig(sport Sport) (SportConfig, error) {
	config, ok := p.configs[sport]
	if !ok {
		return SportConfig{}, fmt.Errorf("%w: %s", ErrUnsupportedSport, sport)
	}
	return config, nil
}

var defaultConfigProvider ConfigProvider = NewConfigProvider(defaultSportConfigs)
