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

func RunAnalysisOnGame(roster players.Roster, opponents players.Roster, endDate time.Time, updateDB bool) []Analysis {
    startDate, _ := time.Parse("2006-01-02", "2018-10-01")
    var predictedStats []Analysis

    roster.Starters = roster.Starters[:5]
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

    if updateDB {
        CreateAndStorePIPPrediction(predictedStats, endDate)
    }

    return predictedStats
}

func CreateAndStorePIPPrediction(analyses []Analysis, date time.Time) {
    log.Printf("Adding %v PIPPredictions to DB", len(analyses))
    var pPreds []players.NBAPIPPrediction
    for _, analysis := range analyses {
        pred := analysis.Prediction.(players.NBAAvg)
        pPred := players.NBAPIPPrediction{
            PlayerIndex: analysis.PlayerIndex,
            Date: date,
            Version: players.CurrNBAPIPPredVersion(),
            NumGames: pred.NumGames,
            Minutes: pred.Minutes,
            Points: int(pred.Points),
            Rebounds: int(pred.Rebounds),
            Assists: int(pred.Assists),
            Usg: pred.Usg,
            Ortg: int(pred.Ortg),
            Drtg: int(pred.Drtg),
        }
        pPreds = append(pPreds, pPred)
    }

    players.AddPIPPrediction(pPreds)
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
