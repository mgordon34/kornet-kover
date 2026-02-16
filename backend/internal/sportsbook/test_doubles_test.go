package sportsbook

import (
	"errors"

	"github.com/mgordon34/kornet-kover/api/odds"
	"github.com/mgordon34/kornet-kover/internal/sports"
)

type fakeSportsbookSources struct {
	getOddsAPIFn  APIGetter
	getPropOddsFn APIGetter
}

func (f fakeSportsbookSources) GetOddsAPI(endpoint string, addlArgs []string) (string, error) {
	if f.getOddsAPIFn == nil {
		return "", errors.New("GetOddsAPI not configured")
	}
	return f.getOddsAPIFn(endpoint, addlArgs)
}

func (f fakeSportsbookSources) GetPropOdds(endpoint string, addlArgs []string) (string, error) {
	if f.getPropOddsFn == nil {
		return "", errors.New("GetPropOdds not configured")
	}
	return f.getPropOddsFn(endpoint, addlArgs)
}

type fakeSportsbookStore struct {
	getLastLineFn       func(oddsType string) (odds.PlayerLine, error)
	addPlayerLinesFn    func(playerLines []odds.PlayerLine)
	playerNameToIndexFn func(nameMap map[string]string, playerName string) (string, error)
}

func (f fakeSportsbookStore) GetLastLine(oddsType string) (odds.PlayerLine, error) {
	if f.getLastLineFn == nil {
		return odds.PlayerLine{}, errors.New("GetLastLine not configured")
	}
	return f.getLastLineFn(oddsType)
}

func (f fakeSportsbookStore) AddPlayerLines(playerLines []odds.PlayerLine) {
	if f.addPlayerLinesFn != nil {
		f.addPlayerLinesFn(playerLines)
	}
}

func (f fakeSportsbookStore) PlayerNameToIndex(nameMap map[string]string, playerName string) (string, error) {
	if f.playerNameToIndexFn == nil {
		return "", errors.New("PlayerNameToIndex not configured")
	}
	return f.playerNameToIndexFn(nameMap, playerName)
}

type fakeConfigProvider struct {
	getSportsbookFn func(sport sports.Sport) (*sports.SportsbookConfig, error)
	getScraperFn    func(sport sports.Sport) (*sports.ScraperConfig, error)
	getAnalysisFn   func(sport sports.Sport) (*sports.AnalysisConfig, error)
}

func (f fakeConfigProvider) SportsbookConfig(sport sports.Sport) (*sports.SportsbookConfig, error) {
	if f.getSportsbookFn == nil {
		return nil, nil
	}
	return f.getSportsbookFn(sport)
}

func (f fakeConfigProvider) ScraperConfig(sport sports.Sport) (*sports.ScraperConfig, error) {
	if f.getScraperFn == nil {
		return nil, nil
	}
	return f.getScraperFn(sport)
}

func (f fakeConfigProvider) AnalysisConfig(sport sports.Sport) (*sports.AnalysisConfig, error) {
	if f.getAnalysisFn == nil {
		return nil, nil
	}
	return f.getAnalysisFn(sport)
}
