package picks

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestGetPropPicksHandler_ParamErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/prop-picks", GetPropPicks)

	req := httptest.NewRequest(http.MethodGet, "/prop-picks?user_id=bad&date=2026-01-01", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/prop-picks?user_id=1&date=bad", nil)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec2.Code)
	}
}

func TestGetPropPicksHandler_UsesInjectedGetter(t *testing.T) {
	original := getPropPicksFn
	getPropPicksFn = func(userID int, date time.Time) ([]PropPickFormatted, error) {
		return []PropPickFormatted{{Id: 1, StratId: 1, StratName: "S", Name: "P", Stat: "points"}}, nil
	}
	t.Cleanup(func() { getPropPicksFn = original })

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/prop-picks", GetPropPicks)

	req := httptest.NewRequest(http.MethodGet, "/prop-picks?user_id=1&date=2026-01-01", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	getPropPicksFn = func(userID int, date time.Time) ([]PropPickFormatted, error) {
		return nil, errors.New("db down")
	}
	req2 := httptest.NewRequest(http.MethodGet, "/prop-picks?user_id=1&date=2026-01-01", nil)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec2.Code)
	}
}

func TestGetPropPickHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/prop-pick/:strat", GetPropPick)

	req := httptest.NewRequest(http.MethodGet, "/prop-pick/not-a-number", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}

	original := getPropPickFn
	getPropPickFn = func(id int) (PropPick, error) { return PropPick{Id: id}, nil }
	t.Cleanup(func() { getPropPickFn = original })

	req2 := httptest.NewRequest(http.MethodGet, "/prop-pick/2", nil)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec2.Code)
	}

	getPropPickFn = func(id int) (PropPick, error) { return PropPick{}, errors.New("boom") }
	req3 := httptest.NewRequest(http.MethodGet, "/prop-pick/3", nil)
	rec3 := httptest.NewRecorder()
	r.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec3.Code)
	}
}

func TestGetBettorPropPicksHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/prop-picks/bettor", GetBettorPropPicks)

	req := httptest.NewRequest(http.MethodGet, "/prop-picks/bettor?user_id=bad", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}

	original := getBettorPicksFn
	getBettorPicksFn = func(userID int, date time.Time) ([]BettorPickRow, error) {
		return nil, errors.New("db down")
	}
	t.Cleanup(func() { getBettorPicksFn = original })

	req2 := httptest.NewRequest(http.MethodGet, "/prop-picks/bettor?user_id=1", nil)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec2.Code)
	}

	getBettorPicksFn = func(userID int, date time.Time) ([]BettorPickRow, error) {
		return []BettorPickRow{{ID: 1, StratID: 1, StratName: "S", PlayerName: "P", Side: "Over", Line: 20.5, Stat: "points", Odds: -110, Points: 22}}, nil
	}

	req3 := httptest.NewRequest(http.MethodGet, "/prop-picks/bettor?user_id=1", nil)
	rec3 := httptest.NewRecorder()
	r.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec3.Code)
	}
}
