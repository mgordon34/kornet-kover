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
    commands := []string{}

    commands = append(
        commands, 
        `CREATE TABLE IF NOT EXISTS sports (
            id SERIAL PRIMARY KEY,
            name VARCHAR(255) NOT NULL
        )`)

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
