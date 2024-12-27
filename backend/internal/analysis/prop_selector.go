package analysis

import (
	"log"
	"math"
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mgordon34/kornet-kover/api/odds"
	"github.com/mgordon34/kornet-kover/api/players"
	"github.com/mgordon34/kornet-kover/internal/scraper"
)

type PropSelector struct {
    Thresholds      map[string]float32
    TresholdType    ThresholdType
    SortType        SortType
    SortDir         string
    RequireOutlier  bool
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

    nbaPred := pick.Prediction.(players.NBAAvg)
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
    // Gather roster for today's games
    games := scraper.ScrapeTodaysGames()
    // games = games[:1]

    // Run analysis on each game
    var results []Analysis
    for _, game := range games {
        log.Printf("Running analysis on %v vs %v", game[0], game[1])
        results = append(results, RunAnalysisOnGame(game[0], game[1], today, true, true)...)
        results = append(results, RunAnalysisOnGame(game[1], game[0], today, true, true)...)
    }

    picker := PropSelector{
        Thresholds: map[string]float32{
            "points": .3,
            "rebounds": .3,
            "assists": .3,
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
    picks, err = picker.PickProps(oddsMap, results)
    if err  != nil {
        return picks, err
    }
    for _, pick := range picks {
        log.Printf("%v: Selected %v %v Predicted %.2f vs. Line %.2f. Diff: %.2f", pick.PlayerIndex, pick.Side, pick.Stat, pick.Prediction.GetStats()[pick.Stat], pick.Over.Line, pick.Diff)
    }

    return picks, nil
}
