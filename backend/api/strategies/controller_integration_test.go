//go:build integration
// +build integration

package strategies

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mgordon34/kornet-kover/internal/storage"
)

func createTestUser(t *testing.T) int {
	t.Helper()
	db := storage.GetDB()
	var userID int
	err := db.QueryRow(context.Background(), `INSERT INTO users (name, email, password) VALUES ($1, $2, $3) RETURNING id`, "test-user", "test-user@example.com", "pw").Scan(&userID)
	if err != nil {
		t.Fatalf("failed creating user: %v", err)
	}
	return userID
}

func TestAddAndGetStrategies(t *testing.T) {
	storage.UseLocalDBForIntegrationTests(t)
	storage.InitTables()

	userID := createTestUser(t)

	stratID, err := addStrategy(Strategy{UserId: userID, Name: "Test Strategy"})
	if err != nil {
		t.Fatalf("addStrategy() error = %v", err)
	}
	if stratID == 0 {
		t.Fatalf("addStrategy() returned id=0")
	}

	all, err := getStrategies(userID)
	if err != nil {
		t.Fatalf("getStrategies() error = %v", err)
	}
	if len(all) == 0 {
		t.Fatalf("getStrategies() expected at least one strategy")
	}

	one, err := getStrategy(stratID)
	if err != nil {
		t.Fatalf("getStrategy() error = %v", err)
	}
	if one.Id != stratID || one.UserId != userID {
		t.Fatalf("getStrategy() unexpected response: %+v", one)
	}
}

func TestGetStrategiesHandlerSuccess(t *testing.T) {
	storage.UseLocalDBForIntegrationTests(t)
	gin.SetMode(gin.TestMode)
	storage.InitTables()
	userID := createTestUser(t)
	_, err := addStrategy(Strategy{UserId: userID, Name: "Handler Strategy"})
	if err != nil {
		t.Fatalf("addStrategy() error = %v", err)
	}

	r := gin.New()
	r.GET("/strategies", GetStrategies)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/strategies?user_id=%d", userID), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	var got []Strategy
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed decoding response: %v", err)
	}
}
