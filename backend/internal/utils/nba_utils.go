package utils

import (
	"time"
)

func DateToNBAYear(date time.Time) int {
    if date.Month() > 9 {
        return date.Year() + 1
    }
    return date.Year()
}
