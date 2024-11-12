package storage

import (
    "context"
    "database/sql"
    "fmt"
    "log"
    "os"
    "sync"

    "github.com/joho/godotenv"
    "github.com/jackc/pgx/v5/pgxpool"
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

var db *sql.DB

func PGXInitDB() {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    dbpool, err := pgxpool.New(context.Background(), os.Getenv("DB_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	var greeting string
	err = dbpool.QueryRow(context.Background(), "select name from players where index='tatumja01' limit 1").Scan(&greeting)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(greeting)
}

func GetDB() (*pgxpool.Pool) {
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

func InitDB() {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    dbHost := os.Getenv("DB_HOST")
    dbPort := os.Getenv("DB_PORT")
    dbUser := os.Getenv("DB_USER")
    dbPass := os.Getenv("DB_PASSWORD")
    dbName := os.Getenv("DB_NAME")

    connectstring := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", dbHost, dbUser, dbPass, dbName, dbPort)
    fmt.Println(connectstring)
    db, err = sql.Open("postgres", fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", dbHost, dbUser, dbPass, dbName, dbPort))

    if err != nil {
        panic(err.Error())
    }

    err = db.Ping()
    if err != nil {
        panic(err.Error())
    }

    fmt.Println("Successfully connected to database")
}

func InitTables() {
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
            usg REAL NOT NULL,
            ortg INT NOT NULL,
            drtg INT NOT NULL,
            CONSTRAINT uq_player_games UNIQUE(player_index, game)
        )`,
        `CREATE TABLE IF NOT EXISTS player_lines (
            id SERIAL PRIMARY KEY,
            sport VARCHAR(255) NOT NULL,
            player_index VARCHAR(20) REFERENCES players(index),
            timestamp timestamp NOT NULL,
            stat VARCHAR(50),
            side VARCHAR(50),
            line REAL NOT NULL,
            odds INT NOT NULL,
            CONSTRAINT uq_prop_index UNIQUE(sport, player_index, timestamp, stat, side)
        )`,
    }

    for _, command := range commands {
        _, err := pgInstance.Exec(context.Background(), command)
        if err != nil {
            log.Fatal("Error initializing table: ", err)
        }
    }
}

// func GetDB() *sql.DB {
// 	return db
// }
