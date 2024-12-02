package analysis

import (
    "log"

	"github.com/mgordon34/kornet-kover/api/odds"
)

type PropSelector struct {
}

type PropPick struct {
    PropOdd         odds.PlayerOdds
    Analysis
}

func (p PropSelector) PickProps(props map[string]odds.PlayerOdds, analyses []Analysis) ([]PropPick, error) {
    var picks []PropPick
    for _, analysis := range analyses {
        log.Printf("Running analysis on %v", analysis.PlayerIndex)
    }

    return picks, nil
}
