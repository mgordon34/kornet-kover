package storage

import (
	"context"
	"log"
	"os"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var (
	pgInstance *pgxpool.Pool
	pgOnce     sync.Once
)

func Ping(ctx context.Context) error {
	return pgInstance.Ping(ctx)
}

func Close() {
	pgInstance.Close()
}

func GetDB() *pgxpool.Pool {
	pgOnce.Do(func() {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}

		dbpool, err := pgxpool.New(context.Background(), os.Getenv("DB_URL"))
		if err != nil {
			log.Fatalf("Unable to connect to database: %v\n", err)
		}

		pgInstance = dbpool
	})

	return pgInstance
}

func InitTables() {
	GetDB()

	commands := []string{
		`CREATE TABLE IF NOT EXISTS teams (
            index VARCHAR(255) PRIMARY KEY,
            name VARCHAR(255) NOT NULL
        )`,
		`CREATE TABLE IF NOT EXISTS games (
            id SERIAL PRIMARY KEY,
            sport VARCHAR(255) NOT NULL,
            home_index VARCHAR(255) REFERENCES teams(index),
            away_index VARCHAR(255) REFERENCES teams(index),
            home_score INT NOT NULL,
            away_score INT NOT NULL,
            date DATE NOT NULL,
            CONSTRAINT uq_games UNIQUE(date, sport, home_index)
        )`,
		`CREATE TABLE IF NOT EXISTS players (
            id SERIAL PRIMARY KEY,
            index VARCHAR(20) UNIQUE,
            sport VARCHAR(255) NOT NULL,
            name VARCHAR(255),
            CONSTRAINT uq_players UNIQUE(index, sport)
        )`,
		`CREATE TABLE IF NOT EXISTS nba_player_games (
            id SERIAL PRIMARY KEY,
            player_index VARCHAR(20) REFERENCES players(index),
            game INT REFERENCES games(id),
            team_index VARCHAR(255) REFERENCES teams(index),
            minutes REAL NOT NULL,
            points INT NOT NULL,
            rebounds INT NOT NULL,
            assists INT NOT NULL,
            threes INT NOT NULL,
            usg REAL NOT NULL,
            ortg INT NOT NULL,
            drtg INT NOT NULL,
            CONSTRAINT uq_player_games UNIQUE(player_index, game)
        )`,
		`CREATE TABLE IF NOT EXISTS mlb_player_games_batting (
            id SERIAL PRIMARY KEY,
            player_index VARCHAR(20) REFERENCES players(index),
            game INT REFERENCES games(id),
            team_index VARCHAR(255) REFERENCES teams(index),
            at_bats INT NOT NULL,
            runs INT NOT NULL,
            hits INT NOT NULL,
            rbis INT NOT NULL,
            home_runs INT NOT NULL,
            walks INT NOT NULL,
            strikeouts INT NOT NULL,
            pas INT NOT NULL,
            pitches INT NOT NULL,
            strikes INT NOT NULL,
            obp REAL NOT NULL,
            slg REAL NOT NULL,
            ops REAL NOT NULL,
            wpa REAL NOT NULL,
            details TEXT NOT NULL,
            CONSTRAINT uq_mlb_player_games_batting UNIQUE(player_index, game)
        )`,
		`CREATE TABLE IF NOT EXISTS mlb_player_games_pitching (
            id SERIAL PRIMARY KEY,
            player_index VARCHAR(20) REFERENCES players(index),
            game INT REFERENCES games(id),
            team_index VARCHAR(255) REFERENCES teams(index),
            innings REAL NOT NULL,
            hits INT NOT NULL,
            runs INT NOT NULL,
            earned_runs INT NOT NULL,
            walks INT NOT NULL,
            strikeouts INT NOT NULL,
            home_runs INT NOT NULL,
            era REAL NOT NULL,
            batters_faced INT NOT NULL,
            wpa REAL NOT NULL,
            CONSTRAINT uq_mlb_player_games_pitching UNIQUE(player_index, game)
        )`,
		`CREATE TABLE IF NOT EXISTS mlb_play_by_plays (
            id SERIAL PRIMARY KEY,
            batter_index VARCHAR(20) REFERENCES players(index),
            pitcher_index VARCHAR(20) REFERENCES players(index),
            game INT REFERENCES games(id),
            inning INT NOT NULL,
            outs INT NOT NULL,
            appearance INT NOT NULL,
            pitches INT NOT NULL,
            result VARCHAR(50) NOT NULL,
            raw_result VARCHAR(200) NOT NULL,
            CONSTRAINT uq_mlb_play_by_plays UNIQUE(game, batter_index, pitcher_index, inning, appearance)
        )`,
		`CREATE TABLE IF NOT EXISTS player_lines (
            id SERIAL PRIMARY KEY,
            sport VARCHAR(255) NOT NULL,
            player_index VARCHAR(20) REFERENCES players(index) UNIQUE,
            timestamp timestamp NOT NULL,
            stat VARCHAR(50),
            side VARCHAR(50),
            type VARCHAR(50) NOT NULL,
            line REAL NOT NULL,
            odds INT NOT NULL,
            link VARCHAR(255),
            CONSTRAINT uq_prop_index UNIQUE(sport, player_index, timestamp, stat, side, line)
        )`,
		`CREATE TABLE IF NOT EXISTS nba_pip_factors (
            id SERIAL PRIMARY KEY,
            player_index VARCHAR(20) REFERENCES players(index),
            other_index VARCHAR(20) REFERENCES players(index),
            relationship VARCHAR(50),
            num_games INT,
            avg_minutes REAL NOT NULL,
            avg_points REAL NOT NULL,
            avg_rebounds REAL NOT NULL,
            avg_assists REAL NOT NULL,
            avg_threes REAL NOT NULL,
            avg_usg REAL NOT NULL,
            avg_ortg REAL NOT NULL,
            avg_drtg REAL NOT NULL,
            CONSTRAINT uq_pip_factors UNIQUE(player_index, other_index, relationship)
        )`,
		`CREATE TABLE IF NOT EXISTS nba_pip_predictions (
            id SERIAL PRIMARY KEY,
            player_index VARCHAR(20) REFERENCES players(index),
            date DATE NOT NULL,
            version INT NOT NULL,
            num_games INT NOT NULL,
            minutes REAL NOT NULL,
            points REAL NOT NULL,
            rebounds REAL NOT NULL,
            assists REAL NOT NULL,
            threes REAL NOT NULL,
            usg REAL NOT NULL,
            ortg REAL NOT NULL,
            drtg REAL NOT NULL,
            CONSTRAINT uq_pip_predictions UNIQUE(player_index, date, version)
        )`,
		`CREATE TABLE IF NOT EXISTS users (
            id SERIAL PRIMARY KEY,
            name VARCHAR(50) NOT NULL,
            email VARCHAR(255) NOT NULL,
            password VARCHAR(255) NOT NULL
        )`,
		`CREATE TABLE IF NOT EXISTS strategies (
            id SERIAL PRIMARY KEY,
            user_id INT REFERENCES users(id),
            name VARCHAR(255) NOT NULL
        )`,
		`CREATE TABLE IF NOT EXISTS strategy_filters (
            id SERIAL PRIMARY KEY,
            strategy_id INT REFERENCES strategies(id),
            function VARCHAR(255) NOT NULL,
            stat VARCHAR(255) NOT NULL,
            operator VARCHAR(50) NOT NULL,
            comparison_type VARCHAR(50) NOT NULL,
            compare_value REAL,
            compare_function VARCHAR(255),
            compare_stat VARCHAR(255),
            modifier_operator VARCHAR(50)
        )`,
		`CREATE TABLE IF NOT EXISTS prop_picks (
            id SERIAL PRIMARY KEY,
            strat_id INT REFERENCES strategies(id),
            line_id INT REFERENCES player_lines(id),
            valid BOOLEAN NOT NULL,
            date DATE NOT NULL,
            CONSTRAINT uq_prop_picks UNIQUE(strat_id, line_id, date)
        )`,
		`CREATE TABLE IF NOT EXISTS active_rosters (
            id SERIAL PRIMARY KEY,
            sport VARCHAR(20) NOT NULL,
            player_index VARCHAR(20) REFERENCES players(index),
            team_index VARCHAR(20) REFERENCES teams(index),
            status VARCHAR(255) NOT NULL,
            avg_minutes REAL NOT NULL,
            last_updated DATE NOT NULL,
            CONSTRAINT uq_active_rosters UNIQUE(sport, player_index)
        )`,
	}

	for _, command := range commands {
		_, err := pgInstance.Exec(context.Background(), command)
		if err != nil {
			log.Fatal("Error initializing table: ", err)
		}
	}
}
