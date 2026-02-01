package syncer

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mgordon34/kornet-kover/api/players"
	"github.com/mgordon34/kornet-kover/internal/sports"
)

type cloudGame struct {
	ID        int       `db:"id"`
	Sport     string    `db:"sport"`
	HomeIndex string    `db:"home_index"`
	AwayIndex string    `db:"away_index"`
	HomeScore int       `db:"home_score"`
	AwayScore int       `db:"away_score"`
	Date      time.Time `db:"date"`
}

type cloudPlayerGame struct {
	PlayerIndex string  `db:"player_index"`
	GameID      int     `db:"game"`
	TeamIndex   string  `db:"team_index"`
	Minutes     float32 `db:"minutes"`
	Points      int     `db:"points"`
	Rebounds    int     `db:"rebounds"`
	Assists     int     `db:"assists"`
	Threes      int     `db:"threes"`
	Usg         float32 `db:"usg"`
	Ortg        int     `db:"ortg"`
	Drtg        int     `db:"drtg"`
}

type cloudPlayer struct {
	Index string `db:"index"`
	Sport string `db:"sport"`
	Name  string `db:"name"`
}

func SyncNBA(ctx context.Context, localDB *pgxpool.Pool, cloudDB *pgxpool.Pool) error {
	localDate, localHasGame, err := latestGameDate(ctx, localDB)
	if err != nil {
		return err
	}

	cloudDate, cloudHasGame, err := latestGameDate(ctx, cloudDB)
	if err != nil {
		return err
	}

	if !cloudHasGame {
		log.Printf("No NBA games found in cloud database")
		return nil
	}

	if localHasGame && sameDate(localDate, cloudDate) {
		log.Printf("Local NBA games are up to date (latest date: %s)", cloudDate.Format("2006-01-02"))
		return nil
	}

	games, err := fetchCloudGames(ctx, cloudDB, localHasGame, localDate)
	if err != nil {
		return err
	}

	if len(games) == 0 {
		log.Printf("No NBA games to sync")
		return nil
	}

	gameIDMap, err := ensureLocalGames(ctx, localDB, games)
	if err != nil {
		return err
	}

	cloudGameIDs := make([]int, 0, len(games))
	for _, game := range games {
		cloudGameIDs = append(cloudGameIDs, game.ID)
	}

	pGames, err := fetchCloudPlayerGames(ctx, cloudDB, cloudGameIDs)
	if err != nil {
		return err
	}

	if len(pGames) == 0 {
		log.Printf("No nba_player_games to sync")
		return nil
	}

	playerIDs := uniquePlayerIndexes(pGames)
	if err := syncPlayers(ctx, cloudDB, playerIDs); err != nil {
		return err
	}

	localPlayerGames := mapPlayerGames(pGames, gameIDMap)
	players.AddPlayerGames(localPlayerGames)

	log.Printf("Synced %d games, %d player games, %d players", len(games), len(localPlayerGames), len(playerIDs))
	return nil
}

func latestGameDate(ctx context.Context, db *pgxpool.Pool) (time.Time, bool, error) {
	var date time.Time
	row := db.QueryRow(ctx, `SELECT date FROM games WHERE sport = $1 ORDER BY date DESC LIMIT 1`, sports.NBA)
	if err := row.Scan(&date); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return time.Time{}, false, nil
		}
		return time.Time{}, false, fmt.Errorf("failed to query latest game date: %w", err)
	}
	return date, true, nil
}

func fetchCloudGames(ctx context.Context, cloudDB *pgxpool.Pool, hasLocal bool, localDate time.Time) ([]cloudGame, error) {
	var rows pgx.Rows
	var err error
	if hasLocal {
		rows, err = cloudDB.Query(ctx, `SELECT id, sport, home_index, away_index, home_score, away_score, date FROM games WHERE sport = $1 AND date >= $2 ORDER BY date ASC`, sports.NBA, localDate)
	} else {
		rows, err = cloudDB.Query(ctx, `SELECT id, sport, home_index, away_index, home_score, away_score, date FROM games WHERE sport = $1 ORDER BY date ASC`, sports.NBA)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query cloud games: %w", err)
	}
	defer rows.Close()

	games, err := pgx.CollectRows(rows, pgx.RowToStructByName[cloudGame])
	if err != nil {
		return nil, fmt.Errorf("failed to read cloud games: %w", err)
	}

	return games, nil
}

func ensureLocalGames(ctx context.Context, localDB *pgxpool.Pool, games []cloudGame) (map[int]int, error) {
	mapping := make(map[int]int, len(games))

	for _, game := range games {
		localID, err := insertOrGetLocalGameID(ctx, localDB, game)
		if err != nil {
			return nil, err
		}
		mapping[game.ID] = localID
	}

	return mapping, nil
}

func insertOrGetLocalGameID(ctx context.Context, localDB *pgxpool.Pool, game cloudGame) (int, error) {
	var localID int
	err := localDB.QueryRow(
		ctx,
		`INSERT INTO games (sport, home_index, away_index, home_score, away_score, date)
        VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT DO NOTHING
        RETURNING id`,
		game.Sport,
		game.HomeIndex,
		game.AwayIndex,
		game.HomeScore,
		game.AwayScore,
		game.Date,
	).Scan(&localID)

	if err == nil {
		return localID, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return 0, fmt.Errorf("failed to insert game: %w", err)
	}

	err = localDB.QueryRow(
		ctx,
		`SELECT id FROM games WHERE date = $1 AND sport = $2 AND home_index = $3`,
		game.Date,
		game.Sport,
		game.HomeIndex,
	).Scan(&localID)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch existing game id: %w", err)
	}

	return localID, nil
}

func fetchCloudPlayerGames(ctx context.Context, cloudDB *pgxpool.Pool, cloudGameIDs []int) ([]cloudPlayerGame, error) {
	rows, err := cloudDB.Query(
		ctx,
		`SELECT player_index, game, team_index, minutes, points, rebounds, assists, threes, usg, ortg, drtg
        FROM nba_player_games
        WHERE game = ANY($1)`,
		pgx.Array(cloudGameIDs),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query cloud player games: %w", err)
	}
	defer rows.Close()

	pGames, err := pgx.CollectRows(rows, pgx.RowToStructByName[cloudPlayerGame])
	if err != nil {
		return nil, fmt.Errorf("failed to read cloud player games: %w", err)
	}

	return pGames, nil
}

func uniquePlayerIndexes(pGames []cloudPlayerGame) []string {
	seen := make(map[string]struct{})
	for _, game := range pGames {
		seen[game.PlayerIndex] = struct{}{}
	}

	players := make([]string, 0, len(seen))
	for playerID := range seen {
		players = append(players, playerID)
	}

	return players
}

func syncPlayers(ctx context.Context, cloudDB *pgxpool.Pool, playerIDs []string) error {
	if len(playerIDs) == 0 {
		return nil
	}

	rows, err := cloudDB.Query(
		ctx,
		`SELECT index, sport, name FROM players WHERE index = ANY($1)`,
		pgx.Array(playerIDs),
	)
	if err != nil {
		return fmt.Errorf("failed to query cloud players: %w", err)
	}
	defer rows.Close()

	cloudPlayers, err := pgx.CollectRows(rows, pgx.RowToStructByName[cloudPlayer])
	if err != nil {
		return fmt.Errorf("failed to read cloud players: %w", err)
	}

	if len(cloudPlayers) == 0 {
		return nil
	}

	localPlayers := make([]players.Player, 0, len(cloudPlayers))
	for _, player := range cloudPlayers {
		localPlayers = append(localPlayers, players.Player{
			Index: player.Index,
			Sport: player.Sport,
			Name:  player.Name,
		})
	}

	players.AddPlayers(localPlayers)
	return nil
}

func mapPlayerGames(pGames []cloudPlayerGame, gameIDMap map[int]int) []players.PlayerGame {
	mapped := make([]players.PlayerGame, 0, len(pGames))
	for _, game := range pGames {
		localGameID, ok := gameIDMap[game.GameID]
		if !ok {
			continue
		}
		mapped = append(mapped, players.PlayerGame{
			PlayerIndex: game.PlayerIndex,
			Game:        localGameID,
			TeamIndex:   game.TeamIndex,
			Minutes:     game.Minutes,
			Points:      game.Points,
			Rebounds:    game.Rebounds,
			Assists:     game.Assists,
			Threes:      game.Threes,
			Usg:         game.Usg,
			Ortg:        game.Ortg,
			Drtg:        game.Drtg,
		})
	}

	return mapped
}

func sameDate(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}
