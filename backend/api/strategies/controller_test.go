package strategies

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

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

func TestGetStrategiesHandlerSuccessAndError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	svc := NewStrategyService(StrategyServiceDeps{GetStrategies: func(userID int) ([]Strategy, error) {
		if userID == 1 {
			return []Strategy{{Id: 1, UserId: 1, Name: "S"}}, nil
		}
		return nil, errors.New("boom")
	}})
	r.GET("/strategies", svc.GetStrategiesHandler())

	req := httptest.NewRequest(http.MethodGet, "/strategies?user_id=1", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/strategies?user_id=2", nil)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec2.Code)
	}
}

func TestGetStrategyHandlerSuccessAndError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	svc := NewStrategyService(StrategyServiceDeps{GetStrategy: func(stratID int) (Strategy, error) {
		if stratID == 1 {
			return Strategy{Id: 1, UserId: 1, Name: "S"}, nil
		}
		return Strategy{}, fmt.Errorf("not found")
	}})
	r.GET("/strategies/:strat", svc.GetStrategyHandler())

	req := httptest.NewRequest(http.MethodGet, "/strategies/1", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/strategies/2", nil)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec2.Code)
	}
}
