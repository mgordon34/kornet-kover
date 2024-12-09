package analysis

import (
	"log"
	"time"

	"github.com/mgordon34/kornet-kover/api/players"
	"github.com/mgordon34/kornet-kover/internal/utils"
)

type Analysis struct {
    PlayerIndex     string
    BaseStats       players.PlayerAvg
    Prediction      players.PlayerAvg
    Outliers        map[string]float32
}

func RunAnalysisOnGame(roster players.Roster, opponents players.Roster, endDate time.Time) []Analysis {
    startDate, _ := time.Parse("2006-01-02", "2018-10-01")
    log.Printf("Running analysis from %v to %v", startDate, endDate)
    var predictedStats []Analysis

    for _, player := range roster.Starters {
        controlMap := players.GetPlayerPerByYear(player, startDate, endDate)

        currYear := utils.DateToNBAYear(endDate)
        _, ok := controlMap[currYear]; if !ok {
            log.Printf("Player %v has no stats for current year. Skipping...", player)
            continue
        }

        var totalPip players.PlayerAvg
        for _, defender := range opponents.Starters {
            affectedMap := players.GetPlayerPerWithPlayerByYear(player, defender, players.Opponent, startDate, endDate)
            pipFactor := players.CalculatePIPFactor(controlMap, affectedMap)

            if totalPip == nil {
                totalPip = pipFactor
            } else {
                totalPip = totalPip.AddAvg(pipFactor)
            }
        }

        prediction := controlMap[currYear].PredictStats(totalPip)
        baseStats := controlMap[currYear].ConvertToStats()
        outliers :=  GetOutliers(baseStats, prediction)
        predictedStats = append(
            predictedStats,
            Analysis{
                PlayerIndex: player,
                BaseStats: baseStats,
                Prediction: prediction,
                Outliers: outliers,
            },
        )
    }

    log.Println(predictedStats)
    return predictedStats
}

func GetOutliers(baseStats players.PlayerAvg, predictedStats players.PlayerAvg) map[string]float32 {
    outliers := make(map[string]float32)

    bStats := baseStats.GetStats()
    pStats := predictedStats.GetStats()
    for stat, value := range pStats {
        pDiff := (value - bStats[stat]) / bStats[stat]
        if pDiff < -.2 ||  pDiff > .2 {
            outliers[stat] = pDiff
        }
    }

    return outliers
}

func (a Analysis) HasOutlier(stat string, side string) bool {
    diff, ok := a.Outliers[stat]; if !ok {
        return false
    }
    if (diff > 0 && side == "Over") || (diff < 0 && side == "Under") {
        return true
    }
    return false
}
