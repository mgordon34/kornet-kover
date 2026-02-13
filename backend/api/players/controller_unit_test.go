package players

import "testing"

func TestPlayerNameToIndex_HardcodedMappings(t *testing.T) {
	nameMap := map[string]string{}

	tests := []struct {
		name string
		want string
	}{
		{name: "Herb Jones", want: "joneshe01"},
		{name: "Moe Wagner", want: "wagnemo01"},
		{name: "Nicolas Claxton", want: "claxtni01"},
		{name: "Cam Johnson", want: "johnsca02"},
	}

	for _, tt := range tests {
		got, err := PlayerNameToIndex(nameMap, tt.name)
		if err != nil {
			t.Fatalf("PlayerNameToIndex(%q) error = %v", tt.name, err)
		}
		if got != tt.want {
			t.Fatalf("PlayerNameToIndex(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestPlayerNameToIndex_UsesCacheBeforeDBLookup(t *testing.T) {
	nameMap := map[string]string{
		"LeBron James": "jamesle01",
		"Nikola Jokic": "jokicni01",
	}

	got, err := PlayerNameToIndex(nameMap, "LeBron James")
	if err != nil {
		t.Fatalf("unexpected error from cached lookup: %v", err)
	}
	if got != "jamesle01" {
		t.Fatalf("cached result = %q, want jamesle01", got)
	}

	got2, err := PlayerNameToIndex(nameMap, "Nikola\u00a0Jokic")
	if err != nil {
		t.Fatalf("unexpected error from normalized cached lookup: %v", err)
	}
	if got2 != "jokicni01" {
		t.Fatalf("normalized cached result = %q, want jokicni01", got2)
	}
}
