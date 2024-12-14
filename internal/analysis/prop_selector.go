package analysis

import (
	"math"
	"sort"

	"github.com/mgordon34/kornet-kover/api/odds"
)

type PropSelector struct {
    Thresholds      map[string]float32
    TresholdType    ThresholdType
    SortType        SortType
    SortDir         string
    RequireOutlier  bool
    MinOdds         int
    MinLine         float32
    MinDiff         float32
    BetSize         float32
    MaxOver         int
    MaxUnder        int
    TotalMax        int
}

type ThresholdType int

const (
    Raw ThresholdType = iota
    Percent
)

type SortType int

const (
    Diff ThresholdType = iota
    PDiff
)

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

func (p PropPick) GetLine() odds.PlayerLine {
    if p.Side == "Over" {
        return p.Over
    }

    return p.Under
}

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

    p.sortPicks(picks)

    var overCount, underCount int
    for _, pick := range picks {
        if (pick.Side == "Over" && overCount >= p.MaxOver) || (pick.Side == "Under" && underCount >= p.MaxUnder) {
            continue
        }
        if p.isPickElligible(pick) {
            selectedPicks = append(selectedPicks, pick)
            if pick.Side == "Over" {
                overCount++
            } else {
                underCount++
            }
        }
    }

    return selectedPicks, nil
}

func (p PropSelector) sortPicks(picks []PropPick) {
    sort.Slice(picks, func(i, j int) bool {
        rankings := map[string]int{
            "points": 3,
            "rebounds": 2,
            "assists": 1,
        }
        if picks[i].Stat == picks[j].Stat {
            return math.Abs(float64(picks[i].PDiff)) > math.Abs(float64(picks[j].PDiff))
        } else {
            return rankings[picks[i].Stat] > rankings[picks[j].Stat]
        }
    })
}

func (p PropSelector) isPickElligible(pick PropPick) bool {
    var diff float64
    switch p.TresholdType {
    case Raw:
        diff = float64(pick.Diff)
    case Percent:
        diff = float64(pick.PDiff)
    }
    if pick.GetLine().Odds < p.MinOdds {
        return false
    }
    if p.MinLine != 0 && pick.GetLine().Line < p.MinLine {
        return false
    }
    if p.MinDiff != 0 && math.Abs(float64(pick.Diff)) < float64(p.MinDiff) {
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
