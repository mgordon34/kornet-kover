//go:build integration
// +build integration

package storage

import (
	"context"
	"testing"
)

func TestGetDBAndPingAndInitTables(t *testing.T) {
	UseLocalDBForIntegrationTests(t)

	db := GetDB()
	if db == nil {
		t.Fatalf("GetDB() returned nil")
	}

	if err := Ping(context.Background()); err != nil {
		t.Fatalf("Ping() error = %v", err)
	}

	InitTables()
	Close()
}
