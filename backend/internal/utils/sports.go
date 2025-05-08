package utils

type Sport string

const (
	NBA Sport = "nba"
	MLB Sport = "mlb"
)

type SportConfig struct {
	Domain      string
	BoxScoreURL string
	OddsName    string
}

var SportConfigs = map[Sport]SportConfig{
	NBA: {
		Domain:      "https://www.basketball-reference.com",
		BoxScoreURL: "/boxscores",
		OddsName:    "basketball_nba",
	},
	MLB: {
		Domain:      "https://www.baseball-reference.com",
		BoxScoreURL: "/boxes",
		OddsName:    "baseball_mlb",
	},
}
