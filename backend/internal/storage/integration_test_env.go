//go:build integration
// +build integration

package storage

import (
	"os"
	"testing"
)

// UseLocalDBForIntegrationTests maps LOCAL_DB_URL to DB_URL for integration tests.
func UseLocalDBForIntegrationTests(t testing.TB) {
	t.Helper()

	localDBURL := os.Getenv("LOCAL_DB_URL")
	if localDBURL == "" {
		t.Skip("LOCAL_DB_URL not set; skipping integration test")
	}

	if err := os.Setenv("DB_URL", localDBURL); err != nil {
		t.Fatalf("failed setting DB_URL from LOCAL_DB_URL: %v", err)
	}
}
