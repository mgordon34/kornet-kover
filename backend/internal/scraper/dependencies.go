package scraper

import (
	"time"

	"github.com/mgordon34/kornet-kover/api/games"
	"github.com/mgordon34/kornet-kover/api/players"
	"github.com/mgordon34/kornet-kover/api/teams"
	"github.com/mgordon34/kornet-kover/internal/sports"
)

type GameRepository interface {
	GetLastGame() (games.Game, error)
}

type TeamRepository interface {
	GetTeams() ([]teams.Team, error)
}

type PlayerRepository interface {
	UpdatePlayerTables(playerIndex string)
	UpdateRosters(rosterSlots []players.PlayerRoster) error
}

type GamesProvider interface {
	ScrapeGames(sport sports.Sport, startDate time.Time, endDate time.Time) error
}

type InjuryProvider interface {
	GetInjuredPlayers() map[string]string
}

type RosterProvider interface {
	ScrapePlayersForTeam(teamIndex string, injuredPlayers map[string]string) []players.PlayerRoster
}

type GameRepositoryFunc func() (games.Game, error)

func (f GameRepositoryFunc) GetLastGame() (games.Game, error) {
	return f()
}

type TeamRepositoryFunc func() ([]teams.Team, error)

func (f TeamRepositoryFunc) GetTeams() ([]teams.Team, error) {
	return f()
}

type GamesProviderFunc func(sport sports.Sport, startDate time.Time, endDate time.Time) error

func (f GamesProviderFunc) ScrapeGames(sport sports.Sport, startDate time.Time, endDate time.Time) error {
	return f(sport, startDate, endDate)
}

type InjuryProviderFunc func() map[string]string

func (f InjuryProviderFunc) GetInjuredPlayers() map[string]string {
	return f()
}

type RosterProviderFunc func(teamIndex string, injuredPlayers map[string]string) []players.PlayerRoster

func (f RosterProviderFunc) ScrapePlayersForTeam(teamIndex string, injuredPlayers map[string]string) []players.PlayerRoster {
	return f(teamIndex, injuredPlayers)
}

type playerRepositoryFuncs struct {
	updatePlayerTables func(playerIndex string)
	updateRosters      func(rosterSlots []players.PlayerRoster) error
}

func (r playerRepositoryFuncs) UpdatePlayerTables(playerIndex string) {
	r.updatePlayerTables(playerIndex)
}

func (r playerRepositoryFuncs) UpdateRosters(rosterSlots []players.PlayerRoster) error {
	return r.updateRosters(rosterSlots)
}

type defaultGameRepository struct{}

func (r defaultGameRepository) GetLastGame() (games.Game, error) {
	return games.GetLastGame()
}

type defaultTeamRepository struct{}

func (r defaultTeamRepository) GetTeams() ([]teams.Team, error) {
	return teams.GetTeams()
}

type defaultPlayerRepository struct{}

func (r defaultPlayerRepository) UpdatePlayerTables(playerIndex string) {
	players.UpdatePlayerTables(playerIndex)
}

func (r defaultPlayerRepository) UpdateRosters(rosterSlots []players.PlayerRoster) error {
	return players.UpdateRosters(rosterSlots)
}

type defaultGamesProvider struct{}

func (p defaultGamesProvider) ScrapeGames(sport sports.Sport, startDate time.Time, endDate time.Time) error {
	return ScrapeGames(sport, startDate, endDate)
}

type defaultInjuryProvider struct{}

func (p defaultInjuryProvider) GetInjuredPlayers() map[string]string {
	return GetInjuredPlayers()
}

type defaultRosterProvider struct{}

func (p defaultRosterProvider) ScrapePlayersForTeam(teamIndex string, injuredPlayers map[string]string) []players.PlayerRoster {
	return scrapePlayersForTeam(teamIndex, injuredPlayers)
}
