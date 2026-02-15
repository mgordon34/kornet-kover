package analysis

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetPickPropsHandler_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/pick-props", PickPropsHandler(func() ([]PropPick, error) { return nil, errors.New("boom") }))

	req := httptest.NewRequest(http.MethodGet, "/pick-props", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestGetPickPropsHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/pick-props", PickPropsHandler(func() ([]PropPick, error) { return []PropPick{{LineId: 1}}, nil }))

	req := httptest.NewRequest(http.MethodGet, "/pick-props", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}
