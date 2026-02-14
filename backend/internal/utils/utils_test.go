package utils

import "testing"

func TestNormalizeString_RemovesDiacriticsAndNBSP(t *testing.T) {
	in := "Nikola\u00a0Jokic\u0301"

	out, err := NormalizeString(in)
	if err != nil {
		t.Fatalf("NormalizeString() error = %v", err)
	}

	if out != "Nikola Jokic" {
		t.Fatalf("NormalizeString() = %q, want %q", out, "Nikola Jokic")
	}
}
