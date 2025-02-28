package utils

type Sport string

const (
	NBA Sport = "nba"
	MLB Sport = "mlb"
)

type SportConfig struct {
	Domain      string
	BoxScoreURL string
}

var SportConfigs = map[Sport]SportConfig{
	NBA: {
		Domain:      "basketball-reference.com",
		BoxScoreURL: "/boxscores",
	},
	MLB: {
		Domain:      "baseball-reference.com",
		BoxScoreURL: "/boxes",
	},
}
