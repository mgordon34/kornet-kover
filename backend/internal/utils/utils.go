package utils

import (
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

func NormalizeString(s string) (string, error) {
    t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
    result, _, err := transform.String(t, s)
    if err != nil {
        return "", err
    }

    return result, nil
}