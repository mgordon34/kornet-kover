package sportsbook

import (
	"github.com/mgordon34/kornet-kover/api/odds"
	"github.com/mgordon34/kornet-kover/api/players"
)

type APIGetter func(url string, addlArgs []string) (response string, err error)

type SportsbookSources interface {
	GetOddsAPI(endpoint string, addlArgs []string) (string, error)
	GetPropOdds(endpoint string, addlArgs []string) (string, error)
}

type SportsbookStore interface {
	GetLastLine(oddsType string) (odds.PlayerLine, error)
	AddPlayerLines(playerLines []odds.PlayerLine)
	PlayerNameToIndex(nameMap map[string]string, playerName string) (string, error)
}

type defaultSportsbookSources struct{}

func (d defaultSportsbookSources) GetOddsAPI(endpoint string, addlArgs []string) (string, error) {
	return requestOddsAPI(endpoint, addlArgs)
}

func (d defaultSportsbookSources) GetPropOdds(endpoint string, addlArgs []string) (string, error) {
	return requestPropOdds(endpoint, addlArgs)
}

type defaultSportsbookStore struct{}

func (d defaultSportsbookStore) GetLastLine(oddsType string) (odds.PlayerLine, error) {
	return odds.GetLastLine(oddsType)
}

func (d defaultSportsbookStore) AddPlayerLines(playerLines []odds.PlayerLine) {
	odds.AddPlayerLines(playerLines)
}

func (d defaultSportsbookStore) PlayerNameToIndex(nameMap map[string]string, playerName string) (string, error) {
	return players.PlayerNameToIndex(nameMap, playerName)
}
