package players

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/mgordon34/kornet-kover/internal/storage"
	"github.com/mgordon34/kornet-kover/internal/utils"
)

func AddPlayers(players []Player) {
	db := storage.GetDB()
	txn, _ := db.Begin(context.Background())
	_, err := txn.Exec(
		context.Background(),
		` CREATE TEMP TABLE players_temp
        ON COMMIT DROP
        AS SELECT * FROM players
        WITH NO DATA`,
	)
	if err != nil {
		panic(err)
	}
	var playersInterface [][]interface{}
	for _, player := range players {
		playersInterface = append(
			playersInterface,
			[]interface{}{player.Index, player.Sport, player.Name},
		)
	}

	_, err = txn.CopyFrom(
		context.Background(),
		pgx.Identifier{"players_temp"},
		[]string{"index", "sport", "name"},
		pgx.CopyFromRows(playersInterface),
	)
	if err != nil {
		panic(err)
	}

	_, err = txn.Exec(
		context.Background(),
		`INSERT INTO players (index, sport, name)
        SELECT index, sport, name FROM players_temp
        ON CONFLICT DO NOTHING`,
	)
	if err != nil {
		panic(err)
	}

	if err := txn.Commit(context.Background()); err != nil {
		panic(err)
	}
}

func GetPlayer(index string) (Player, error) {
	db := storage.GetDB()
	sql := `SELECT index, sport, name from players where index = ($1)`

	rows, err := db.Query(context.Background(), sql, index)
	if err != nil {
		log.Fatal("Error querying for player index: ", err)
	}
	defer rows.Close()

	player, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Player])
	if err != nil {
		log.Println("Error converting player: ", err)
		return Player{}, err
	}

	return player, nil
}

func PlayerNameToIndex(nameMap map[string]string, playerName string) (string, error) {
	playerName, err := utils.NormalizeString(playerName)
	if err != nil {
		return "", err
	}

	if playerName == "Herb Jones" {
		return "joneshe01", nil
	} else if playerName == "Moe Wagner" {
		return "wagnemo01", nil
	} else if playerName == "Nicolas Claxton" {
		return "claxtni01", nil
	} else if playerName == "Cam Johnson" {
		return "johnsca02", nil
	}
	index, ok := nameMap[playerName]
	if ok {
		return index, nil
	}

	db := storage.GetDB()
	sql := `SELECT index FROM players WHERE UPPER(name) ILIKE ($1);`
	if playerName == "Alexandre Sarr" {
		playerName = "Alex Sarr"
	}

	row := db.QueryRow(context.Background(), sql, playerName+"%")
	if err := row.Scan(&index); err != nil {
		log.Printf("Error finding player index for %s", playerName)
		return "", err
	}
	nameMap[playerName] = index
	return index, nil
}

func AddPlayerGames(pGames []PlayerGame) {
	db := storage.GetDB()
	txn, _ := db.Begin(context.Background())
	_, err := txn.Exec(
		context.Background(),
		`CREATE TEMP TABLE player_games_temp
        ON COMMIT DROP
        AS SELECT * FROM nba_player_games
        WITH NO DATA`,
	)
	if err != nil {
		panic(err)
	}
	var playersInterface [][]interface{}
	for _, pGame := range pGames {
		playersInterface = append(
			playersInterface,
			[]interface{}{
				pGame.PlayerIndex,
				pGame.Game,
				pGame.TeamIndex,
				pGame.Minutes,
				pGame.Points,
				pGame.Rebounds,
				pGame.Assists,
				pGame.Threes,
				pGame.Usg,
				pGame.Ortg,
				pGame.Drtg,
			},
		)
	}

	_, err = txn.CopyFrom(
		context.Background(),
		pgx.Identifier{"player_games_temp"},
		[]string{
			"player_index",
			"game",
			"team_index",
			"minutes",
			"points",
			"rebounds",
			"assists",
			"threes",
			"usg",
			"ortg",
			"drtg",
		},
		pgx.CopyFromRows(playersInterface),
	)
	if err != nil {
		panic(err)
	}

	_, err = txn.Exec(
		context.Background(),
		` INSERT INTO nba_player_games (player_index, game, team_index, minutes, points, rebounds, assists, threes, usg, ortg, drtg)
        SELECT player_index, game, team_index, minutes, points, rebounds, assists, threes, usg, ortg, drtg FROM player_games_temp
        ON CONFLICT DO NOTHING`,
	)
	if err != nil {
		panic(err)
	}

	if err := txn.Commit(context.Background()); err != nil {
		panic(err)
	}
}

func AddMLBPlayerGamesBatting(pGames []MLBPlayerGameBatting) {
	db := storage.GetDB()
	txn, _ := db.Begin(context.Background())
	_, err := txn.Exec(
		context.Background(),
		`CREATE TEMP TABLE player_games_temp
        ON COMMIT DROP
        AS SELECT * FROM mlb_player_games_batting
        WITH NO DATA`,
	)
	if err != nil {
		panic(err)
	}
	var playersInterface [][]interface{}
	for _, pGame := range pGames {
		playersInterface = append(
			playersInterface,
			[]interface{}{
				pGame.PlayerIndex,
				pGame.Game,
				pGame.TeamIndex,
				pGame.AtBats,
				pGame.Runs,
				pGame.Hits,
				pGame.RBIs,
				pGame.HomeRuns,
				pGame.Walks,
				pGame.Strikeouts,
				pGame.PAs,
				pGame.Pitches,
				pGame.Strikes,
				pGame.OBP,
				pGame.SLG,
				pGame.OPS,
				pGame.WPA,
				pGame.Details,
			},
		)
	}

	_, err = txn.CopyFrom(
		context.Background(),
		pgx.Identifier{"player_games_temp"},
		[]string{
			"player_index",
			"game",
			"team_index",
			"at_bats",
			"runs",
			"hits",
			"rbis",
			"home_runs",
			"walks",
			"strikeouts",
			"pas",
			"pitches",
			"strikes",
			"obp",
			"slg",
			"ops",
			"wpa",
			"details",
		},
		pgx.CopyFromRows(playersInterface),
	)
	if err != nil {
		panic(err)
	}

	_, err = txn.Exec(
		context.Background(),
		` INSERT INTO mlb_player_games_batting (player_index, game, team_index, at_bats, runs, hits, rbis, home_runs, walks, strikeouts, pas, pitches, strikes, obp, slg, ops, wpa, details)
        SELECT player_index, game, team_index, at_bats, runs, hits, rbis, home_runs, walks, strikeouts, pas, pitches, strikes, obp, slg, ops, wpa, details FROM player_games_temp
        ON CONFLICT DO NOTHING`,
	)
	if err != nil {
		panic(err)
	}

	if err := txn.Commit(context.Background()); err != nil {
		panic(err)
	}

}

func AddMLBPlayerGamesPitching(pGames []MLBPlayerGamePitching) {
	db := storage.GetDB()
	txn, _ := db.Begin(context.Background())
	_, err := txn.Exec(
		context.Background(),
		`CREATE TEMP TABLE player_games_temp
        ON COMMIT DROP
        AS SELECT * FROM mlb_player_games_pitching
        WITH NO DATA`,
	)
	if err != nil {
		panic(err)
	}
	var playersInterface [][]interface{}
	for _, pGame := range pGames {
		playersInterface = append(
			playersInterface,
			[]interface{}{
				pGame.PlayerIndex,
				pGame.Game,
				pGame.TeamIndex,
				pGame.Innings,
				pGame.Hits,
				pGame.Runs,
				pGame.EarnedRuns,
				pGame.Walks,
				pGame.Strikeouts,
				pGame.HomeRuns,
				pGame.ERA,
				pGame.BattersFaced,
				pGame.WPA,
			},
		)
	}

	_, err = txn.CopyFrom(
		context.Background(),
		pgx.Identifier{"player_games_temp"},
		[]string{
			"player_index",
			"game",
			"team_index",
			"innings",
			"hits",
			"runs",
			"earned_runs",
			"walks",
			"strikeouts",
			"home_runs",
			"era",
			"batters_faced",
			"wpa",
		},
		pgx.CopyFromRows(playersInterface),
	)
	if err != nil {
		panic(err)
	}

	_, err = txn.Exec(
		context.Background(),
		` INSERT INTO mlb_player_games_pitching (player_index, game, team_index, innings, hits, runs, earned_runs, walks, strikeouts, home_runs, era, batters_faced, wpa)
        SELECT player_index, game, team_index, innings, hits, runs, earned_runs, walks, strikeouts, home_runs, era, batters_faced, wpa FROM player_games_temp
        ON CONFLICT DO NOTHING`,
	)
	if err != nil {
		panic(err)
	}

	if err := txn.Commit(context.Background()); err != nil {
		panic(err)
	}

}

func AddMLBPlayByPlays(pbp []MLBPlayByPlay) {
	db := storage.GetDB()
	txn, _ := db.Begin(context.Background())

	_, err := txn.Exec(
		context.Background(),
		`CREATE TEMPORARY TABLE IF NOT EXISTS play_by_play_temp (
			batter_index VARCHAR(20),
			pitcher_index VARCHAR(20),
			game INT,
			inning INT,
			outs INT,
			appearance INT,
			pitches INT,
			result VARCHAR(50),
			raw_result VARCHAR(200)
		) ON COMMIT DROP`,
	)
	if err != nil {
		panic(err)
	}

	playersInterface := make([][]interface{}, len(pbp))
	for i, play := range pbp {
		playersInterface[i] = []interface{}{
			play.BatterIndex,
			play.PitcherIndex,
			play.Game,
			play.Inning,
			play.Outs,
			play.Appearance,
			play.Pitches,
			play.Result,
			play.RawResult,
		}
	}

	_, err = txn.CopyFrom(
		context.Background(),
		pgx.Identifier{"play_by_play_temp"},
		[]string{
			"batter_index",
			"pitcher_index",
			"game",
			"inning",
			"outs",
			"appearance",
			"pitches",
			"result",
			"raw_result",
		},
		pgx.CopyFromRows(playersInterface),
	)
	if err != nil {
		panic(err)
	}

	_, err = txn.Exec(
		context.Background(),
		`INSERT INTO mlb_play_by_plays (batter_index, pitcher_index, game, inning, outs, appearance, pitches, result, raw_result)
		SELECT batter_index, pitcher_index, game, inning, outs, appearance, pitches, result, raw_result FROM play_by_play_temp
		ON CONFLICT DO NOTHING`,
	)
	if err != nil {
		panic(err)
	}

	if err := txn.Commit(context.Background()); err != nil {
		panic(err)
	}
}

func GetPlayerStats(player string, startDate time.Time, endDate time.Time) (PlayerAvg, error) {
	db := storage.GetDB()
	sql := `SELECT count(*) as num_games, avg(minutes) as minutes, avg(points) as points, avg(rebounds) as rebounds, 
            avg(assists) as assists, avg(threes) as threes, avg(usg) as usg, avg(ortg) as ortg, avg(drtg) as drtg FROM nba_player_games
                left join games on games.id = nba_player_games.game
                where nba_player_games.player_index = ($1) and nba_player_games.minutes > 10 and games.date between ($2) and ($3)`

	rows, err := db.Query(context.Background(), sql, player, startDate.Format(time.DateOnly), endDate.AddDate(0, 0, -1).Format(time.DateOnly))
	if err != nil {
		log.Fatal("Error querying for player stats: ", err)
	}
	defer rows.Close()

	stats, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[NBAAvg])
	if err != nil {
		return NBAAvg{}, err
	}

	return stats, nil
}

type Relationship int

const (
	Teammate Relationship = iota
	Opponent
)

func GetPlayersForGame(gameId int, homeIndex string, playerGameTable string, sortString string) (map[string][]Player, error) {
	playerMap := make(map[string][]Player)
	db := storage.GetDB()
	sql := `SELECT pl.index, pl.name, pg.team_index FROM players pl
                 LEFT JOIN ` + playerGameTable + ` pg ON pg.player_index=pl.index
                 LEFT JOIN games gg ON gg.id=pg.game
                 WHERE gg.id=($1)
                 ORDER BY pg.` + sortString + ` DESC`
	rows, err := db.Query(context.Background(), sql, gameId)
	if err != nil {
		log.Fatal("Error querying for player games: ", err)
	}
	defer rows.Close()
	for rows.Next() {
		var player Player
		var teamIndex string
		err = rows.Scan(&player.Index, &player.Name, &teamIndex)
		if err != nil {
			log.Fatal("Error converting rows to playerLines: ", err)
		}

		if teamIndex == homeIndex {
			playerMap["home"] = append(playerMap["home"], player)
		} else {
			playerMap["away"] = append(playerMap["away"], player)
		}
	}

	return playerMap, nil
}

func GetMLBPlayersMissingHandedness() ([]Player, error) {
	var playerSlice []Player
	db := storage.GetDB()
	sql := `SELECT pl.index, pl.name, pl.details 
			FROM players pl 
			WHERE pl.sport='mlb' AND 
			(pl.details IS NULL OR 
			NOT (pl.details ? 'batting_handedness' AND pl.details ? 'pitching_handedness'))`
	rows, err := db.Query(context.Background(), sql)
	if err != nil {
		log.Fatal("Error querying for no handedness players: ", err)
	}
	defer rows.Close()
	for rows.Next() {
		var player Player
		err = rows.Scan(&player.Index, &player.Name, &player.Details)
		if err != nil {
			log.Fatal("Error converting rows to player: ", err)
		}
		playerSlice = append(playerSlice, player)
	}

	return playerSlice, nil
}

func AddMLBPlayerHandedness(playerIndex string, bats string, throws string) error {
	db := storage.GetDB()
	sql := `UPDATE players 
			SET details = COALESCE(details, '{}'::jsonb) || 
				jsonb_build_object('batting_handedness', $1::text, 'pitching_handedness', $2::text)
			WHERE index = $3`

	_, err := db.Exec(context.Background(), sql, bats, throws, playerIndex)
	if err != nil {
		return fmt.Errorf("error updating player handedness: %v", err)
	}

	return nil
}

type PlayerStatInfo struct {
	PlayerIndex string `json:"player_index"`
	NBAAvg
}

func GetPlayerStatsForGames(gameIds []string) (map[string]PlayerAvg, error) {
	playerMap := make(map[string]PlayerAvg)
	db := storage.GetDB()
	sql := `SELECT player_index, 1 as num_games, minutes, points, rebounds, assists, threes, usg, ortg, drtg FROM nba_player_games
                left join games gg on gg.id = nba_player_games.game
                where gg.id IN (%s)`

	param := strings.Join(gameIds, ",")
	sql = fmt.Sprintf(sql, param)

	rows, err := db.Query(context.Background(), sql)
	if err != nil {
		log.Fatal("Error querying for player stats: ", err)
	}
	defer rows.Close()

	stats, err := pgx.CollectRows(rows, pgx.RowToStructByName[PlayerStatInfo])
	if err != nil {
		return playerMap, err
	}

	for _, stat := range stats {
		playerMap[stat.PlayerIndex] = stat.NBAAvg
	}

	return playerMap, nil
}

type MLBPlayerStatInfo struct {
	PlayerIndex string `json:"player_index"`
	MLBBattingAvg
}

func GetMLBBattingStatsForGames(gameIds []string) (map[string]MLBBattingAvg, error) {
	playerMap := make(map[string]MLBBattingAvg)
	db := storage.GetDB()
	sql := `SELECT player_index, 1 as num_games, at_bats, runs, hits, rbis, home_runs, walks, strikeouts, pas, pitches, strikes, obp, slg, ops, wpa FROM mlb_player_games_batting
                left join games gg on gg.id = mlb_player_games_batting.game
                where gg.id IN (%s)`

	param := strings.Join(gameIds, ",")
	sql = fmt.Sprintf(sql, param)

	rows, err := db.Query(context.Background(), sql)
	if err != nil {
		log.Fatal("Error querying for player stats: ", err)
	}
	defer rows.Close()

	stats, err := pgx.CollectRows(rows, pgx.RowToStructByName[MLBPlayerStatInfo])
	if err != nil {
		return playerMap, err
	}

	for _, stat := range stats {
		playerMap[stat.PlayerIndex] = stat.MLBBattingAvg
	}

	return playerMap, nil
}

func GetPlayerPerByYear(player string, startDate time.Time, endDate time.Time) map[int]PlayerAvg {
	playerStats := make(map[int]PlayerAvg)

	for d := startDate; !d.After(endDate); d = d.AddDate(1, 0, 0) {
		useDate := d.AddDate(1, 0, 0)
		if useDate.After(endDate) {
			useDate = endDate
		}

		yearlyStats, _ := GetPlayerStats(player, d, useDate)
		if yearlyStats.IsValid() {
			playerStats[utils.DateToNBAYear(d)] = yearlyStats.ConvertToPer()
		}
	}

	return playerStats
}

func GetPlayerStatsWithPlayer(player string, defender string, relationship Relationship, startDate time.Time, endDate time.Time) (PlayerAvg, error) {
	db := storage.GetDB()
	sql := `SELECT count(*) as num_games, avg(minutes) as minutes, avg(points) as points, avg(rebounds) as rebounds, 
            avg(assists) as assists, avg(threes) as threes, avg(usg) as usg, avg(ortg) as ortg, avg(drtg) as drtg FROM nba_player_games
                left join games gg on gg.id = nba_player_games.game
                where nba_player_games.player_index = ($1) and nba_player_games.minutes > 10 and gg.date between ($3) and ($4)`
	opponent_filter := `
        AND (
            SELECT COUNT(*) FROM games ga
            LEFT JOIN nba_player_games pg ON pg.game=ga.id
            WHERE ga.id=gg.id AND pg.player_index IN (($1),($2))
        ) > 1`
	teammate_filter := `
         AND (
             SELECT COUNT(*) FROM games ga
             LEFT JOIN nba_player_games pg ON pg.game=ga.id
             WHERE ga.id=gg.id AND pg.player_index=($1)
         ) = 1
         AND (
             SELECT COUNT(*) FROM games ga
             LEFT JOIN nba_player_games pg ON pg.game=ga.id
             WHERE ga.id=gg.id AND pg.player_index=($2)
         ) = 0`

	switch relationship {
	case Teammate:
		sql = sql + teammate_filter
	case Opponent:
		sql = sql + opponent_filter
	}

	rows, err := db.Query(context.Background(), sql, player, defender, startDate.Format(time.DateOnly), endDate.AddDate(0, 0, -1).Format(time.DateOnly))
	if err != nil {
		log.Fatal("Error querying for player stats: ", err)
	}
	defer rows.Close()

	stats, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[NBAAvg])
	if err != nil {
		return NBAAvg{}, err
	}

	return stats, nil
}

func GetPlayerPerWithPlayerByYear(player string, defender string, relationship Relationship, startDate time.Time, endDate time.Time) map[int]PlayerAvg {
	playerStats := make(map[int]PlayerAvg)

	for d := startDate; !d.After(endDate); d = d.AddDate(1, 0, 0) {
		useDate := d.AddDate(1, 0, 0)
		if useDate.After(endDate) {
			useDate = endDate
		}

		yearlyStats, _ := GetPlayerStatsWithPlayer(player, defender, relationship, d, useDate)
		playerStats[utils.DateToNBAYear(d)] = yearlyStats.ConvertToPer()
	}

	return playerStats
}

func AddPIPPrediction(pPreds []NBAPIPPrediction) {
	db := storage.GetDB()
	txn, _ := db.Begin(context.Background())
	_, err := txn.Exec(
		context.Background(),
		`CREATE TEMP TABLE pip_prediction_temp
        ON COMMIT DROP
        AS SELECT * FROM nba_pip_predictions
        WITH NO DATA`,
	)
	if err != nil {
		panic(err)
	}
	var predsInterface [][]interface{}
	for _, pPred := range pPreds {
		predsInterface = append(
			predsInterface,
			[]interface{}{
				pPred.PlayerIndex,
				pPred.Date,
				pPred.Version,
				pPred.NumGames,
				pPred.Minutes,
				pPred.Points,
				pPred.Rebounds,
				pPred.Assists,
				pPred.Threes,
				pPred.Usg,
				pPred.Ortg,
				pPred.Drtg,
			},
		)
	}

	_, err = txn.CopyFrom(
		context.Background(),
		pgx.Identifier{"pip_prediction_temp"},
		[]string{
			"player_index",
			"date",
			"version",
			"num_games",
			"minutes",
			"points",
			"rebounds",
			"assists",
			"threes",
			"usg",
			"ortg",
			"drtg",
		},
		pgx.CopyFromRows(predsInterface),
	)
	if err != nil {
		panic(err)
	}

	_, err = txn.Exec(
		context.Background(),
		` INSERT INTO nba_pip_predictions (player_index, date, version, num_games, minutes, points, rebounds, assists, threes, usg, ortg, drtg)
        SELECT player_index, date, version, num_games, minutes, points, rebounds, assists, threes, usg, ortg, drtg FROM pip_prediction_temp
        ON CONFLICT (player_index, date, version) DO UPDATE
        SET num_games=excluded.num_games, minutes=excluded.minutes, points=excluded.points, rebounds=excluded.rebounds,
        assists=excluded.assists, threes=excluded.threes, usg=excluded.usg, ortg=excluded.ortg, drtg=excluded.drtg`,
	)
	if err != nil {
		panic(err)
	}

	if err := txn.Commit(context.Background()); err != nil {
		panic(err)
	}
}

func GetPIPPredictionsForDate(date time.Time) ([]NBAPIPPrediction, error) {
	var pipPreds []NBAPIPPrediction
	db := storage.GetDB()
	sql := `SELECT player_index, date, version, num_games, minutes, points, rebounds, assists, threes, usg, ortg, drtg FROM nba_pip_predictions
                where date=($1)`

	rows, err := db.Query(context.Background(), sql, date.Format(time.DateOnly))
	if err != nil {
		log.Fatal("Error querying for NBAPIPPredictions: ", err)
	}
	defer rows.Close()

	pipPreds, err = pgx.CollectRows(rows, pgx.RowToStructByName[NBAPIPPrediction])
	if err != nil {
		return pipPreds, err
	}

	return pipPreds, nil
}

func GetPlayerPIPPrediction(playerIndex string, date time.Time) (NBAPIPPrediction, error) {
	db := storage.GetDB()
	sql := `SELECT player_index, date, version, num_games, minutes, points, rebounds, assists, threes, usg, ortg, drtg FROM nba_pip_predictions
                where date=($1) and player_index=($2)`

	rows, err := db.Query(context.Background(), sql, date.Format(time.DateOnly), playerIndex)
	if err != nil {
		return NBAPIPPrediction{}, err
	}
	defer rows.Close()

	pipPred, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[NBAPIPPrediction])
	if err != nil {
		return NBAPIPPrediction{}, err
	}

	return pipPred, nil
}

func CalculatePIPFactor(controlMap map[int]PlayerAvg, relatedMap map[int]PlayerAvg) PlayerAvg {
	var totals PlayerAvg
	for year := range relatedMap {
		pChange := relatedMap[year].CompareAvg(controlMap[year])
		if totals == nil {
			totals = pChange
		} else {
			totals = totals.AddAvg(pChange)
		}
	}

	return totals
}

func GetOrCreatePrediction(playerIndex string, date time.Time) PlayerAvg {
	pipPred, err := GetPlayerPIPPrediction(playerIndex, date)
	if err != nil {
		log.Println("Failed to find PIPPrediction: ", err)
	}

	nbaAvg := NBAAvg{
		NumGames: pipPred.NumGames,
		Minutes:  pipPred.Minutes,
		Points:   float32(pipPred.Points),
		Rebounds: float32(pipPred.Rebounds),
		Assists:  float32(pipPred.Assists),
		Threes:   float32(pipPred.Threes),
		Usg:      pipPred.Usg,
		Ortg:     float32(pipPred.Ortg),
		Drtg:     float32(pipPred.Drtg),
	}
	return nbaAvg
}

func UpdatePlayerTables(playerIndex string) {
	db := storage.GetDB()
	sql := `SELECT count(1) FROM players
                where index=($1)`

	var count int
	row := db.QueryRow(context.Background(), sql, playerIndex)
	if err := row.Scan(&count); err != nil {
		log.Printf("Error finding player index for %s", playerIndex)
	}
	if count == 0 {
		log.Printf("Could not find player %v, adding...", playerIndex)
		sql = `
        INSERT INTO players (index, name, sport)
        VALUES ($1, $2, $3)
        ON CONFLICT DO NOTHING
        RETURNING id`
		var resId int
		err := db.QueryRow(context.Background(), sql, playerIndex, "Placeholder Name", "nba").Scan(&resId)
		if err != nil {
			log.Printf("Error adding player %s: %v", playerIndex, err)
		}
		log.Println(row)
		log.Printf("Added player: %v", playerIndex)
	}
}

func UpdateRosters(rosterSlots []PlayerRoster) error {
	log.Printf("Updating %v slots on active_rosters", len(rosterSlots))
	t := time.Now()
	today := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	db := storage.GetDB()

	txn, _ := db.Begin(context.Background())
	_, err := txn.Exec(
		context.Background(),
		`CREATE TEMP TABLE active_rosters_temp
        ON COMMIT DROP
        AS SELECT * FROM active_rosters
        WITH NO DATA`,
	)
	if err != nil {
		return err
	}
	var rosterInterface [][]interface{}
	for _, rSlot := range rosterSlots {
		rosterInterface = append(
			rosterInterface,
			[]interface{}{
				rSlot.Sport,
				rSlot.PlayerIndex,
				rSlot.TeamIndex,
				rSlot.Status,
				rSlot.AvgMins,
				today,
			},
		)
	}

	_, err = txn.CopyFrom(
		context.Background(),
		pgx.Identifier{"active_rosters_temp"},
		[]string{
			"sport",
			"player_index",
			"team_index",
			"status",
			"avg_minutes",
			"last_updated",
		},
		pgx.CopyFromRows(rosterInterface),
	)
	if err != nil {
		return err
	}

	_, err = txn.Exec(
		context.Background(),
		`INSERT INTO active_rosters (sport, player_index, team_index, status, avg_minutes, last_updated)
        SELECT sport, player_index, team_index, status, avg_minutes, last_updated FROM active_rosters_temp
        ON CONFLICT (sport, player_index) DO UPDATE
        SET team_index=excluded.team_index, status=excluded.status, avg_minutes=excluded.avg_minutes, 
        last_updated=excluded.last_updated`,
	)
	if err != nil {
		return err
	}

	if err := txn.Commit(context.Background()); err != nil {
		return err
	}

	log.Println("Done updating rosters")
	return nil
}

func GetActiveRosters() (map[string][]PlayerRoster, error) {
	rosterMap := make(map[string][]PlayerRoster)
	db := storage.GetDB()
	sql := `SELECT id, sport, player_index, team_index, status, avg_minutes FROM active_rosters
            ORDER BY avg_minutes DESC`

	rows, err := db.Query(context.Background(), sql)
	if err != nil {
		msg := fmt.Sprint("Error querying for active roster: ", err)
		log.Println(msg)
		return rosterMap, errors.New(msg)
	}
	defer rows.Close()

	playerRosters, err := pgx.CollectRows(rows, pgx.RowToStructByName[PlayerRoster])
	if err != nil {
		msg := fmt.Sprint("Error scanning active_roster query: ", err)
		log.Println(msg)
		return rosterMap, errors.New(msg)
	}

	for _, player := range playerRosters {
		rosterMap[player.TeamIndex] = append(rosterMap[player.TeamIndex], player)
	}

	return rosterMap, nil
}
