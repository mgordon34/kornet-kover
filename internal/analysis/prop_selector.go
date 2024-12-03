package analysis

import (
	"log"
	"math"

	"github.com/mgordon34/kornet-kover/api/odds"
)

type PropSelector struct {
    Thresholds      map[string]float32
    TresholdType    ThresholdType
    RequireOutlier  bool
    MaxOver         int
    MaxUnder        int
    TotalMax        int
}

type PropPick struct {
    Stat            string
    Side            string
    Diff            float32
    PDiff           float32
    odds.PlayerOdds
    Analysis
}

type ThresholdType int

const (
    Raw ThresholdType = iota
    Percent
)

func (p PropSelector) PickProps(props map[string]map[string]odds.PlayerOdds, analyses []Analysis) ([]PropPick, error) {
    var picks, selectedPicks []PropPick
    for _, analysis := range analyses {
        log.Printf("Running analysis on %v", analysis.PlayerIndex)

        log.Println("---------------------------------------")
        for stat, prediction := range analysis.Prediction.GetStats() {
            line, ok := props[analysis.PlayerIndex][stat]; if !ok {
                continue
            }
            diff, pDiff := GetOddsDiff(props[analysis.PlayerIndex][stat], prediction)
            var side string
            if diff > 0 {
                side = "Over"
            } else {
                side = "Under"
            }
            pick := PropPick{
                Stat: stat,
                Side: side,
                Diff: diff,
                PDiff: pDiff,
                PlayerOdds: props[analysis.PlayerIndex][stat],
                Analysis: analysis,
            }
            picks = append(picks, pick)
            log.Printf("%v: %s prediction %.2f vs line %.2f. Diff: %.2f PDiff %.2f%%", analysis.PlayerIndex, stat, prediction, line.Over.Line, pick.Diff, pick.PDiff*100)
        }
    }
    var overCount, underCount int
    for _, pick := range picks {
        if p.isPickElligible(pick) {
            selectedPicks = append(selectedPicks, pick)
            if pick.Side == "Over" {
                overCount++
            } else {
                underCount++
            }
        }

        if overCount >= p.MaxOver || underCount >= p.MaxUnder || overCount + underCount >= p.TotalMax {
            break
        }
    }

    return selectedPicks, nil
}

func (p PropSelector) isPickElligible(pick PropPick) bool {
    var diff float64
    switch p.TresholdType {
    case Raw:
        diff = float64(pick.Diff)
    case Percent:
        diff = float64(pick.PDiff)
    }

    threshold, ok := p.Thresholds[pick.Stat]; if !ok {
        return false
    }
    return math.Abs(diff) > float64(threshold)
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
