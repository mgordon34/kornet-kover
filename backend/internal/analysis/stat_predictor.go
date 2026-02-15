package analysis

import (
	"log"
	"time"

	"github.com/mgordon34/kornet-kover/api/players"
	"github.com/mgordon34/kornet-kover/internal/sports"
	"github.com/mgordon34/kornet-kover/internal/utils"
)

type Analysis struct {
	PlayerIndex string
	BaseStats   players.PlayerAvg
	Prediction  players.PlayerAvg
	Outliers    map[string]float32
}

type AnalysisStore interface {
	GetPlayerPIPPrediction(playerIndex string, date time.Time) (players.NBAPIPPrediction, error)
	AddPIPPrediction(predictions []players.NBAPIPPrediction)
	GetPlayerPerByYear(sport sports.Sport, player string, startDate time.Time, endDate time.Time) map[int]players.PlayerAvg
	GetPlayerPerWithPlayerByYear(player string, defender string, relationship players.Relationship, startDate time.Time, endDate time.Time) map[int]players.PlayerAvg
	GetMLBPlayerPerWithPlayerByYear(player string, defender string, startDate time.Time, endDate time.Time) map[int]players.PlayerAvg
	CalculatePIPFactor(controlMap map[int]players.PlayerAvg, relatedMap map[int]players.PlayerAvg) players.PlayerAvg
}

type AnalysisServiceDeps struct {
	Store AnalysisStore
}

type AnalysisService struct {
	deps AnalysisServiceDeps
}

type defaultAnalysisStore struct{}

func (d defaultAnalysisStore) GetPlayerPIPPrediction(playerIndex string, date time.Time) (players.NBAPIPPrediction, error) {
	return players.GetPlayerPIPPrediction(playerIndex, date)
}

func (d defaultAnalysisStore) AddPIPPrediction(predictions []players.NBAPIPPrediction) {
	players.AddPIPPrediction(predictions)
}

func (d defaultAnalysisStore) GetPlayerPerByYear(sport sports.Sport, player string, startDate time.Time, endDate time.Time) map[int]players.PlayerAvg {
	return players.GetPlayerPerByYear(sport, player, startDate, endDate)
}

func (d defaultAnalysisStore) GetPlayerPerWithPlayerByYear(player string, defender string, relationship players.Relationship, startDate time.Time, endDate time.Time) map[int]players.PlayerAvg {
	return players.GetPlayerPerWithPlayerByYear(player, defender, relationship, startDate, endDate)
}

func (d defaultAnalysisStore) GetMLBPlayerPerWithPlayerByYear(player string, defender string, startDate time.Time, endDate time.Time) map[int]players.PlayerAvg {
	return players.GetMLBPlayerPerWithPlayerByYear(player, defender, startDate, endDate)
}

func (d defaultAnalysisStore) CalculatePIPFactor(controlMap map[int]players.PlayerAvg, relatedMap map[int]players.PlayerAvg) players.PlayerAvg {
	return players.CalculatePIPFactor(controlMap, relatedMap)
}

func NewAnalysisService(deps AnalysisServiceDeps) *AnalysisService {
	if deps.Store == nil {
		deps.Store = defaultAnalysisStore{}
	}
	return &AnalysisService{deps: deps}
}

func (s *AnalysisService) RunAnalysisOnGame(roster []players.PlayerRoster, opponents []players.PlayerRoster, endDate time.Time, forceUpdate bool, storePIP bool) []Analysis {
	startDate, _ := time.Parse("2006-01-02", "2018-10-01")
	var predictedStats []Analysis

	prunedPlayers := prunePlayers(roster)
	prunedOpponents := prunePlayers(opponents)

	for _, player := range prunedPlayers[:min(len(prunedPlayers), 5)] {
		controlMap := s.deps.Store.GetPlayerPerByYear(sports.NBA, player, startDate, endDate)

		currYear := utils.DateToNBAYear(endDate)
		_, ok := controlMap[currYear]
		if !ok {
			log.Printf("Player %v has no stats for current year. Skipping...", player)
			continue
		}

		pipPred := s.GetOrCreatePrediction(player, prunedOpponents[:min(len(prunedOpponents), 8)], players.Opponent, controlMap, startDate, endDate, forceUpdate)
		prediction := players.NBAAvg{
			NumGames: pipPred.NumGames,
			Minutes:  pipPred.Minutes,
			Points:   pipPred.Points,
			Rebounds: pipPred.Rebounds,
			Assists:  pipPred.Assists,
			Threes:   pipPred.Threes,
			Usg:      pipPred.Usg,
			Ortg:     pipPred.Ortg,
			Drtg:     pipPred.Drtg,
		}

		baseStats := controlMap[currYear].ConvertToStats()
		outliers := GetOutliers(baseStats, prediction)
		predictedStats = append(
			predictedStats,
			Analysis{
				PlayerIndex: player,
				BaseStats:   baseStats,
				Prediction:  prediction,
				Outliers:    outliers,
			},
		)
	}

	if storePIP {
		s.CreateAndStorePIPPrediction(predictedStats, endDate)
	}

	return predictedStats
}

func (s *AnalysisService) RunMLBAnalysisOnGame(roster []players.PlayerRoster, opponents []players.PlayerRoster, endDate time.Time, forceUpdate bool, storePIP bool) []Analysis {
	startDate, _ := time.Parse("2006-01-02", "2019-03-01")
	var predictedStats []Analysis

	prunedPlayers := prunePlayers(roster)
	prunedOpponents := prunePlayers(opponents)

	for _, player := range prunedPlayers[:min(len(prunedPlayers), 9)] {
		controlMap := s.deps.Store.GetPlayerPerByYear(sports.MLB, player, startDate, endDate)

		_, ok := controlMap[endDate.Year()]
		if !ok {
			log.Printf("Player %v has no stats for current year. Skipping...", player)
			continue
		}

		pipPred := s.CreateMLBPrediction(player, prunedOpponents[:1], players.Opponent, controlMap, startDate, endDate)
		log.Printf("PIPPred: %v", pipPred)

		// yearlyStats := controlMap[endDate.Year()].(players.MLBBattingAvg)
	}

	if storePIP {
		s.CreateAndStorePIPPrediction(predictedStats, endDate)
	}

	return predictedStats
}

func prunePlayers(roster []players.PlayerRoster) []string {
	var activePlayers []string

	for _, player := range roster {
		if player.Status == "Available" && player.AvgMins > 10 {
			activePlayers = append(activePlayers, player.PlayerIndex)
		}
	}

	return activePlayers
}

func (s *AnalysisService) GetOrCreatePrediction(playerIndex string, opponents []string, relationship players.Relationship, controlMap map[int]players.PlayerAvg, startDate time.Time, endDate time.Time, forceUpdate bool) players.NBAPIPPrediction {
	if forceUpdate {
		log.Printf("Force creating new PIPPrediction on %v players...", len(opponents))
		return s.CreatePIPPrediction(playerIndex, opponents, relationship, controlMap, startDate, endDate)
	}

	pipPred, err := s.deps.Store.GetPlayerPIPPrediction(playerIndex, endDate)
	if err != nil {
		log.Println("Could not find PIPPrediction, creating new:", err)
		pipPred = s.CreatePIPPrediction(playerIndex, opponents, relationship, controlMap, startDate, endDate)
	}

	return pipPred
}

func (s *AnalysisService) CreatePIPPrediction(playerIndex string, opponents []string, relationship players.Relationship, controlMap map[int]players.PlayerAvg, startDate time.Time, endDate time.Time) players.NBAPIPPrediction {
	var totalPip players.PlayerAvg
	currYear := utils.DateToNBAYear(endDate)

	for _, defender := range opponents {
		affectedMap := s.deps.Store.GetPlayerPerWithPlayerByYear(playerIndex, defender, players.Opponent, startDate, endDate)
		pipFactor := s.deps.Store.CalculatePIPFactor(controlMap, affectedMap)

		if totalPip == nil {
			totalPip = pipFactor
		} else {
			totalPip = totalPip.AddAvg(pipFactor)
		}
	}

	pred := controlMap[currYear].PredictStats(totalPip).(players.NBAAvg)
	prediction := players.NBAPIPPrediction{
		PlayerIndex: playerIndex,
		Date:        endDate,
		Version:     players.CurrNBAPIPPredVersion(),
		NumGames:    pred.NumGames,
		Minutes:     pred.Minutes,
		Points:      pred.Points,
		Rebounds:    pred.Rebounds,
		Assists:     pred.Assists,
		Threes:      pred.Threes,
		Usg:         pred.Usg,
		Ortg:        pred.Ortg,
		Drtg:        pred.Drtg,
	}

	return prediction
}

func (s *AnalysisService) CreateAndStorePIPPrediction(analyses []Analysis, date time.Time) {
	log.Printf("Adding %v PIPPredictions to DB", len(analyses))
	var pPreds []players.NBAPIPPrediction
	for _, analysis := range analyses {
		pred := analysis.Prediction.(players.NBAAvg)
		pPred := players.NBAPIPPrediction{
			PlayerIndex: analysis.PlayerIndex,
			Date:        date,
			Version:     players.CurrNBAPIPPredVersion(),
			NumGames:    pred.NumGames,
			Minutes:     pred.Minutes,
			Points:      pred.Points,
			Rebounds:    pred.Rebounds,
			Assists:     pred.Assists,
			Threes:      pred.Threes,
			Usg:         pred.Usg,
			Ortg:        pred.Ortg,
			Drtg:        pred.Drtg,
		}
		pPreds = append(pPreds, pPred)
	}

	s.deps.Store.AddPIPPrediction(pPreds)
}

func (s *AnalysisService) CreateMLBPrediction(playerIndex string, opponents []string, relationship players.Relationship, controlMap map[int]players.PlayerAvg, startDate time.Time, endDate time.Time) players.MLBBattingAvg {
	var totalPip players.PlayerAvg

	for _, defender := range opponents {
		log.Printf("Batter: %v, Defender: %v", playerIndex, defender)
		affectedMap := s.deps.Store.GetMLBPlayerPerWithPlayerByYear(playerIndex, defender, startDate, endDate)
		log.Printf("AffectedMap: %v", affectedMap)
		pipFactor := s.deps.Store.CalculatePIPFactor(controlMap, affectedMap)
		log.Printf("PipFactor: %v", pipFactor)

		if totalPip == nil {
			totalPip = pipFactor
		} else {
			totalPip = totalPip.AddAvg(pipFactor)
		}
	}

	pred := controlMap[endDate.Year()].PredictStats(totalPip).(players.MLBBattingAvg)

	return pred
}

func GetOutliers(baseStats players.PlayerAvg, predictedStats players.PlayerAvg) map[string]float32 {
	outliers := make(map[string]float32)

	bStats := baseStats.GetStats()
	pStats := predictedStats.GetStats()
	for stat, value := range pStats {
		diff := value - bStats[stat]
		pDiff := (value - bStats[stat]) / bStats[stat]
		if (pDiff < -.9 || pDiff > .0) && (diff > 0 || diff < -100) {
			outliers[stat] = pDiff
		}
	}

	return outliers
}

func (a Analysis) HasOutlier(stat string, side string) bool {
	diff, ok := a.Outliers[stat]
	if !ok {
		return false
	}
	if (diff > 0 && side == "Over") || (diff < 0 && side == "Under") {
		return true
	}
	return false
}
