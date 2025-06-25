package sports

// You could add caching if needed
var configCache = make(map[Sport]Config)

// Main config getter (kept for cases where full config is needed)
func New(sport Sport) Config {
    if cached, ok := configCache[sport]; ok {
        return cached
    }

    var config Config
    switch sport {
    case NBA:
        config = NewNBA()
    case WNBA:
        config = NewWNBA()
    case MLB:
        config = NewMLB()
    default:
        return nil
    }

    // Store in cache before returning
    configCache[sport] = config
    return config
}

// Convenience getters
func GetSportsbook(sport Sport) *SportsbookConfig {
    return New(sport).GetSportsbookConfig()
}

func GetScraper(sport Sport) *ScraperConfig {
    return New(sport).GetScraperConfig()
}

func GetAnalysis(sport Sport) *AnalysisConfig {
    return New(sport).GetAnalysisConfig()
} 
