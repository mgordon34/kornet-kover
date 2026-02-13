package utils

import (
	"testing"
	"time"
)

func TestDateToNBAYear(t *testing.T) {
	tests := []struct {
		name string
		date time.Time
		want int
	}{
		{name: "October rolls to next season", date: time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC), want: 2026},
		{name: "September stays in same year", date: time.Date(2025, 9, 30, 0, 0, 0, 0, time.UTC), want: 2025},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DateToNBAYear(tt.date)
			if got != tt.want {
				t.Fatalf("DateToNBAYear() = %d, want %d", got, tt.want)
			}
		})
	}
}
