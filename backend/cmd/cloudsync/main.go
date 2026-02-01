package main

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/mgordon34/kornet-kover/internal/storage"
	"github.com/mgordon34/kornet-kover/internal/syncer"
)

const cloudDBEnv = "CLOUD_DB_URL"

func main() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found; relying on environment variables")
	}

	cloudURL := os.Getenv(cloudDBEnv)
	if cloudURL == "" {
		log.Fatalf("%s is required", cloudDBEnv)
	}

	ctx := context.Background()
	localDB := storage.GetDB()
	cloudDB, err := pgxpool.New(ctx, cloudURL)
	if err != nil {
		log.Fatalf("Unable to connect to cloud database: %v", err)
	}
	defer cloudDB.Close()

	if err := syncer.SyncNBA(ctx, localDB, cloudDB); err != nil {
		log.Fatalf("Sync failed: %v", err)
	}
}
