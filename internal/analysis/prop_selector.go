package analysis

import (
	"math"
    "sort"

	"github.com/mgordon34/kornet-kover/api/odds"
)

type PropSelector struct {
    Thresholds      map[string]float32
    TresholdType    ThresholdType
    RequireOutlier  bool
    MinOdds         int
    BetSize         float32
    MaxOver         int
    MaxUnder        int
    TotalMax        int
}

type PropPick struct {
    Stat            string
    Side            string
    Diff            float32
    PDiff           float32
    BetSize         float32
    Actual          float32
    Result          string
    odds.PlayerOdds
    Analysis
}

func (p PropPick) GetOdds() odds.PlayerLine {
    if p.Side == "Over" {
        return p.Over
    }

    return p.Under
}

type ThresholdType int

const (
    Raw ThresholdType = iota
    Percent
)

func (p PropSelector) PickProps(props map[string]map[string]odds.PlayerOdds, analyses []Analysis) ([]PropPick, error) {
    var picks, selectedPicks []PropPick
    for _, analysis := range analyses {

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
                BetSize: p.BetSize,
                PlayerOdds: line,
                Analysis: analysis,
            }
            picks = append(picks, pick)
        }
    }
    var overCount, underCount int
    sort.Slice(picks, func(i, j int) bool {
        return picks[i].PDiff > picks[j].PDiff
    })
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
    var odds int; var line float32
    switch pick.Side {
    case "Over":
        odds = pick.Over.Odds
        line = pick.Over.Line
    case "Under":
        odds = pick.Under.Odds
        line = pick.Under.Line
    }
    if odds < p.MinOdds {
        return false
    }
    if line < 2 {
        return false
    }

    threshold, ok := p.Thresholds[pick.Stat]; if !ok {
        return false
    }
    if p.RequireOutlier && !pick.HasOutlier(pick.Stat, pick.Side) {
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
