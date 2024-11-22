package analysis

import (
	"log"
	"time"

	"github.com/mgordon34/kornet-kover/api/players"
)

type Analysis struct {
    PlayerIndex     string
    BaseStats       players.PlayerAvg
    Prediction      players.PlayerAvg
}

func RunAnalysisOnGame(roster players.Roster, opponents players.Roster) []Analysis {
    startDate, _ := time.Parse("2006-01-02", "2018-10-01")
    endDate := time.Now()
    var predictedStats []Analysis
    log.Println("running analysis")

    for _, player := range roster.Starters {
        controlMap := players.GetPlayerPerByYear(player, startDate, endDate)
        var totalPip players.PlayerAvg
        for _, defender := range opponents.Starters {
            log.Printf("%v defended by %v", player, defender)
            affectedMap := players.GetPlayerPerWithPlayerByYear(player, defender, players.Opponent, startDate, endDate)
            pipFactor := players.CalculatePIPFactor(controlMap, affectedMap)

            if totalPip == nil {
                totalPip = pipFactor
            } else {
                totalPip.AddAvg(pipFactor)
            }
        }

        prediction := controlMap[2024].PredictStats(totalPip)
        baseStats := controlMap[2024].ConvertToStats()
        predictedStats = append(
            predictedStats,
            Analysis{PlayerIndex: player, BaseStats: baseStats, Prediction: prediction},
        )
    }

    return predictedStats
}
