package strategies

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
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
	if os.Getenv("DB_URL") == "" {
		t.Skip("DB_URL not set; skipping integration test")
	}
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

func TestGetStrategiesHandlerBadUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/strategies", GetStrategies)

	req := httptest.NewRequest(http.MethodGet, "/strategies?user_id=bad", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
}

func TestGetStrategyHandlerBadParam(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/strategies/:strat", GetStrategy)

	req := httptest.NewRequest(http.MethodGet, "/strategies/not-a-number", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
}

func TestGetStrategiesHandlerSuccess(t *testing.T) {
	if os.Getenv("DB_URL") == "" {
		t.Skip("DB_URL not set; skipping integration test")
	}
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
