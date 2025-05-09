package sports

import "github.com/mgordon34/kornet-kover/internal/utils"

// New returns the appropriate sport configuration
func New(sport utils.Sport) Config {
    switch sport {
    case utils.NBA:
        return NewNBA()
    case utils.MLB:
        return NewMLB()
    default:
        return nil
    }
} 