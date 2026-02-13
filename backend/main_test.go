package main

import (
	"slices"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mgordon34/kornet-kover/api/players"
)

func TestConvertPlayerMaptoPlayerRosters(t *testing.T) {
	in := []players.Player{{Index: "a"}, {Index: "b"}}
	out := convertPlayerMaptoPlayerRosters(in)

	if len(out) != 2 {
		t.Fatalf("len = %d, want 2", len(out))
	}
	if out[0].PlayerIndex != "a" || out[0].Status != "Available" {
		t.Fatalf("unexpected first roster: %+v", out[0])
	}
}

func TestNewRouterRegistersExpectedRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := newRouter()
	routes := r.Routes()

	var paths []string
	for _, route := range routes {
		paths = append(paths, route.Path)
	}

	expected := []string{
		"/update-games",
		"/update-players",
		"/update-lines",
		"/pick-props",
		"/strategies",
		"/prop-picks",
		"/prop-picks/bettor",
	}

	for _, path := range expected {
		if !slices.Contains(paths, path) {
			t.Fatalf("expected route %s to be registered; routes=%v", path, paths)
		}
	}
}
