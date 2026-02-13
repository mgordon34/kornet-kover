//go:build integration
// +build integration

package storage

import (
	"context"
	"os"
	"testing"
)

func TestGetDBAndPingAndInitTables(t *testing.T) {
	if os.Getenv("DB_URL") == "" {
		t.Skip("DB_URL not set; skipping integration test")
	}

	db := GetDB()
	if db == nil {
		t.Fatalf("GetDB() returned nil")
	}

	if err := Ping(context.Background()); err != nil {
		t.Fatalf("Ping() error = %v", err)
	}

	InitTables()
}
