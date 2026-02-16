package scraper

import (
	"time"

	"github.com/mgordon34/kornet-kover/api/games"
	"github.com/mgordon34/kornet-kover/api/players"
	"github.com/mgordon34/kornet-kover/api/teams"
	"github.com/mgordon34/kornet-kover/internal/sports"
)

type ScraperSources interface {
	ScrapeGames(sport sports.Sport, startDate time.Time, endDate time.Time) error
	GetInjuredPlayers() map[string]string
	ScrapePlayersForTeam(teamIndex string, injuredPlayers map[string]string) []players.PlayerRoster
}

type ScraperStore interface {
	GetLastGame() (games.Game, error)
	GetTeams() ([]teams.Team, error)
	UpdatePlayerTables(playerIndex string)
	UpdateRosters(rosterSlots []players.PlayerRoster) error
}

type defaultScraperSources struct{}

func (d defaultScraperSources) ScrapeGames(sport sports.Sport, startDate time.Time, endDate time.Time) error {
	return ScrapeGames(sport, startDate, endDate)
}

func (d defaultScraperSources) GetInjuredPlayers() map[string]string {
	return GetInjuredPlayers()
}

func (d defaultScraperSources) ScrapePlayersForTeam(teamIndex string, injuredPlayers map[string]string) []players.PlayerRoster {
	return scrapePlayersForTeam(teamIndex, injuredPlayers)
}

type defaultScraperStore struct{}

func (d defaultScraperStore) GetLastGame() (games.Game, error) {
	return games.GetLastGame()
}

func (d defaultScraperStore) GetTeams() ([]teams.Team, error) {
	return teams.GetTeams()
}

func (d defaultScraperStore) UpdatePlayerTables(playerIndex string) {
	players.UpdatePlayerTables(playerIndex)
}

func (d defaultScraperStore) UpdateRosters(rosterSlots []players.PlayerRoster) error {
	return players.UpdateRosters(rosterSlots)
}
