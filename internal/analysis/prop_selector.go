package analysis

import (
	"log"

	"github.com/mgordon34/kornet-kover/api/odds"
)

type PropSelector struct {
    Thresholds      map[string]float32
    RequireOutlier  bool
    OrderBy         string
    MaxOver         int
    MaxUnder        int
    TotalMax        int
}

type PropPick struct {
    Stat            string
    PropOdd         odds.PlayerOdds
    Analysis
}

func (p PropSelector) PickProps(props map[string]map[string]odds.PlayerOdds, analyses []Analysis) ([]PropPick, error) {
    var picks []PropPick
    for _, analysis := range analyses {
        log.Printf("Running analysis on %v", analysis.PlayerIndex)

        log.Println("---------------------------------------")
        for stat, prediction := range analysis.Prediction.GetStats() {
            line, ok := props[analysis.PlayerIndex][stat]; if !ok {
                continue
            }
            diff := GetOddsDiff(props[analysis.PlayerIndex][stat], prediction)
            log.Printf("%v: %s prediction %.2f vs line %.2f. Diff: %.2f", analysis.PlayerIndex, stat, prediction, line.Over.Line, diff)


        }
    }

    return picks, nil
}

func GetOddsDiff(pOdds odds.PlayerOdds, prediction float32) float32 {
    line := pOdds.Over.Line

    if prediction > line {
        return prediction - pOdds.Over.Line
    } else {
        return prediction - pOdds.Under.Line
    }
}
