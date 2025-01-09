package analysis

import (
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mgordon34/kornet-kover/api/odds"
	"github.com/mgordon34/kornet-kover/api/picks"
	"github.com/mgordon34/kornet-kover/api/players"
	"github.com/mgordon34/kornet-kover/internal/scraper"
)

type PropSelector struct {
    StratId         int
    StratName       string
    Thresholds      map[string]float32
    TresholdType    ThresholdType
    SortType        SortType
    SortDir         string
    RequireOutlier  bool
    MinMinutes      float32
    MinGames        int
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
    UserId          int
    LineId          int
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

func (p PropSelector) PickProps(props map[string]map[string]odds.PlayerOdds, analyses []Analysis, date time.Time, savePicks bool) ([]PropPick, error) {
    var pPicks, selectedPicks []PropPick
    for _, analysis := range analyses {

        for stat, prediction := range analysis.Prediction.GetStats() {
            line, ok := props[analysis.PlayerIndex][stat]; if !ok {
                continue
            }
            diff, pDiff := GetOddsDiff(props[analysis.PlayerIndex][stat], prediction)
            var side string
            var lineId int
            if diff > 0 {
                side = "Over"
                lineId = line.Over.Id
            } else {
                side = "Under"
                lineId = line.Under.Id
            }
            pick := PropPick{
                LineId: lineId,
                Stat: stat,
                Side: side,
                Diff: diff,
                PDiff: pDiff,
                BetSize: p.BetSize,
                PlayerOdds: line,
                Analysis: analysis,
            }
            pPicks = append(pPicks, pick)
        }
    }

    p.sortPicks(pPicks)

    var overCount, underCount int
    for _, pick := range pPicks {
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

    if savePicks {
        models := p.convertToPicksModel(selectedPicks, date)
        picks.MarkOldPicksInvalid(p.StratId, date)
        err := picks.AddPropPicks(models)
        if err != nil {
            return selectedPicks, errors.New(fmt.Sprintf("Error getting saving picks: %v", err))
        }
    }

    return selectedPicks, nil
}

func (p PropSelector) convertToPicksModel(pPicks []PropPick, date time.Time) []picks.PropPick {
    var models []picks.PropPick
    for _, pick := range pPicks {
        models = append(models, picks.PropPick{
            StratId: p.StratId,
            LineId: pick.LineId,
            Valid: true,
            Date: date,
        })
    }

    return models
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

    nbaPred := pick.Prediction.(players.NBAAvg)
    if p.MinMinutes != 0 && nbaPred.Minutes < p.MinMinutes {
        return false
    }
    if p.MinGames != 0 && nbaPred.NumGames < p.MinGames {
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

func GetPickProps(c *gin.Context) {
    picks, err := runPickProps()
    if err != nil {
        log.Printf("500 for picks-props: %s", err)
        c.JSON(http.StatusInternalServerError, err)
    }
    c.JSON(http.StatusOK, picks)
}

func runPickProps() ([]PropPick, error) {
    var picks []PropPick

    loc, _ := time.LoadLocation("America/New_York")
    t := time.Now()
    today := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)

    // Gather player Odds map for upcoming games
    oddsMap, err := odds.GetPlayerOddsForDate(today, []string{"points, rebounds, assists"})
    if err  != nil {
        return picks, err
    }

    var results []Analysis
    rosterMap, err := players.GetActiveRosters()
    if err != nil {
        return picks, err
    }
    matchups := scraper.ScrapeTodaysGames()

    for _, matchup := range matchups {
        results = append(results, RunAnalysisOnGame(rosterMap[matchup[0]], rosterMap[matchup[1]], today, true, true)...)
        results = append(results, RunAnalysisOnGame(rosterMap[matchup[1]], rosterMap[matchup[0]], today, true, true)...)
    }

    picker := PropSelector{
        StratId: 1,
        StratName: "Percent",
        Thresholds: map[string]float32{
            "points": .3,
            "rebounds": 10,
            "assists": 10,
        },
        TresholdType: Percent,
        RequireOutlier: false,
        MinGames: 10,
        MinOdds: -135,
        BetSize: 100,
        MaxOver: 100,
        MaxUnder: 0,
        TotalMax: 100,
    }
    picks, err = picker.PickProps(oddsMap, results, today, true)
    if err  != nil {
        return picks, err
    }
    log.Printf("================================%v=========================================", picker.StratName)
    for _, pick := range picks {
        log.Printf("%v: Selected %v %v Predicted %.2f vs. Line %.2f. Diff: %.2f, PDiff: %.2f, ID: %v", pick.PlayerIndex, pick.Side, pick.Stat, pick.Prediction.GetStats()[pick.Stat], pick.Over.Line, pick.Diff, pick.PDiff, pick.LineId)
    }

    rpicker := PropSelector{
        StratId: 2,
        StratName: "Raw",
        Thresholds: map[string]float32{
            "points": 2.5,
            "rebounds": 1,
            "assists": 1000,
        },
        TresholdType: Raw,
        RequireOutlier: false,
        MinGames: 10,
        MinOdds: -135,
        BetSize: 100,
        MaxOver: 100,
        MaxUnder: 0,
        TotalMax: 100,
    }
    apicks, err := rpicker.PickProps(oddsMap, results, today, true)
    if err  != nil {
        return picks, err
    }
    log.Printf("================================%v=========================================", rpicker.StratName)
    for _, pick := range apicks {
        log.Printf("%v: Selected %v %v Predicted %.2f vs. Line %.2f. Diff: %.2f, ID: %v", pick.PlayerIndex, pick.Side, pick.Stat, pick.Prediction.GetStats()[pick.Stat], pick.Over.Line, pick.Diff, pick.LineId)
    }

    return picks, nil
}
