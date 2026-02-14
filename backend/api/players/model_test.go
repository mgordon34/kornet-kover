package players

import "testing"

func TestCurrNBAPIPPredVersion(t *testing.T) {
	if got := CurrNBAPIPPredVersion(); got != 1 {
		t.Fatalf("CurrNBAPIPPredVersion() = %d, want 1", got)
	}
}
