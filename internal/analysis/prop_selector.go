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
    Diff            float32
    PDiff           float32
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
            diff, pDiff := GetOddsDiff(props[analysis.PlayerIndex][stat], prediction)
            pick := PropPick{
                Stat: stat,
                Diff: diff,
                PDiff: pDiff,
                PropOdd: props[analysis.PlayerIndex][stat],
                Analysis: analysis,
            }
            picks = append(picks, pick)
            log.Printf("%v: %s prediction %.2f vs line %.2f. Diff: %.2f PDiff %.2f%%", analysis.PlayerIndex, stat, prediction, line.Over.Line, pick.Diff, pick.PDiff*100)
        }
    }

    return picks, nil
}

func GetOddsDiff(pOdds odds.PlayerOdds, prediction float32) (float32, float32) {
    line := pOdds.Over.Line
    if prediction < line {
        line = pOdds.Under.Line
    }    

    diff := prediction - line
    pDiff := diff / line
    return diff, pDiff
}
