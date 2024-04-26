package storage

import (

    "database/sql"
    "fmt"
    "log"
    "os"

    "github.com/joho/godotenv"
    _ "github.com/lib/pq"
)


var db *sql.DB

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
            CONSTRAINT uq_games UNIQUE(date, home_index)
        )`,
        `CREATE TABLE IF NOT EXISTS players (
            id SERIAL PRIMARY KEY,
            index VARCHAR(20) UNIQUE,
            name VARCHAR(255)
        )`,
        `CREATE TABLE IF NOT EXISTS player_games (
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
    }

    for _, command := range commands {
        _, err := db.Exec(command)
        if err != nil {
            log.Fatal("Error initializing table: ", err)
        }
    }
}

func GetDB() *sql.DB {
	return db
}
