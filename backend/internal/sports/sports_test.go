package sports

import "testing"

func TestNew_ReturnsExpectedConfigAndCaches(t *testing.T) {
	configCache = make(map[Sport]Config)

	first := New(NBA)
	second := New(NBA)

	if first == nil || second == nil {
		t.Fatalf("New(NBA) returned nil config")
	}

	if first != second {
		t.Fatalf("expected cached config pointer to match")
	}

	if first.GetSport() != NBA {
		t.Fatalf("GetSport() = %s, want %s", first.GetSport(), NBA)
	}
}

func TestNew_UnsupportedSport(t *testing.T) {
	configCache = make(map[Sport]Config)

	got := New(NHL)
	if got != nil {
		t.Fatalf("New(NHL) = %#v, want nil", got)
	}
}

func TestConvenienceGetters(t *testing.T) {
	configCache = make(map[Sport]Config)

	if GetSportsbook(MLB) == nil {
		t.Fatalf("GetSportsbook(MLB) returned nil")
	}
	if GetScraper(WNBA) == nil {
		t.Fatalf("GetScraper(WNBA) returned nil")
	}
	if GetAnalysis(NBA) == nil {
		t.Fatalf("GetAnalysis(NBA) returned nil")
	}
}
