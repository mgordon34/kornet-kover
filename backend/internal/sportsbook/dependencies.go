package sportsbook

import (
	"github.com/mgordon34/kornet-kover/api/odds"
	"github.com/mgordon34/kornet-kover/api/players"
	"github.com/mgordon34/kornet-kover/internal/sports"
)

type APIGetter func(url string, addlArgs []string) (response string, err error)

type OddsProvider interface {
	Get(endpoint string, addlArgs []string) (string, error)
}

type LineReader interface {
	GetLastLine(oddsType string) (odds.PlayerLine, error)
}

type LineWriter interface {
	AddPlayerLines(playerLines []odds.PlayerLine)
}

type PlayerIndexResolver interface {
	PlayerNameToIndex(nameMap map[string]string, playerName string) (string, error)
}

type SportsbookConfigRepository interface {
	GetSportsbook(sport sports.Sport) *sports.SportsbookConfig
}

type OddsProviderFunc func(endpoint string, addlArgs []string) (string, error)

func (f OddsProviderFunc) Get(endpoint string, addlArgs []string) (string, error) {
	return f(endpoint, addlArgs)
}

type LineReaderFunc func(oddsType string) (odds.PlayerLine, error)

func (f LineReaderFunc) GetLastLine(oddsType string) (odds.PlayerLine, error) {
	return f(oddsType)
}

type LineWriterFunc func(playerLines []odds.PlayerLine)

func (f LineWriterFunc) AddPlayerLines(playerLines []odds.PlayerLine) {
	f(playerLines)
}

type PlayerIndexResolverFunc func(nameMap map[string]string, playerName string) (string, error)

func (f PlayerIndexResolverFunc) PlayerNameToIndex(nameMap map[string]string, playerName string) (string, error) {
	return f(nameMap, playerName)
}

type SportsbookConfigRepositoryFunc func(sport sports.Sport) *sports.SportsbookConfig

func (f SportsbookConfigRepositoryFunc) GetSportsbook(sport sports.Sport) *sports.SportsbookConfig {
	return f(sport)
}

type oddsAPIProvider struct{}

func (p oddsAPIProvider) Get(endpoint string, addlArgs []string) (string, error) {
	return requestOddsAPI(endpoint, addlArgs)
}

type propOddsProvider struct{}

func (p propOddsProvider) Get(endpoint string, addlArgs []string) (string, error) {
	return requestPropOdds(endpoint, addlArgs)
}

type lineRepository struct{}

func (r lineRepository) GetLastLine(oddsType string) (odds.PlayerLine, error) {
	return odds.GetLastLine(oddsType)
}

func (r lineRepository) AddPlayerLines(playerLines []odds.PlayerLine) {
	odds.AddPlayerLines(playerLines)
}

type playerRepository struct{}

func (r playerRepository) PlayerNameToIndex(nameMap map[string]string, playerName string) (string, error) {
	return players.PlayerNameToIndex(nameMap, playerName)
}

type sportsbookConfigRepository struct{}

func (r sportsbookConfigRepository) GetSportsbook(sport sports.Sport) *sports.SportsbookConfig {
	return sports.GetSportsbook(sport)
}
