package sports

import "fmt"

func GetConfig(sport Sport) (SportConfig, error) {
	config, ok := Configs[sport]
	if !ok {
		return SportConfig{}, fmt.Errorf("%w: %s", ErrUnsupportedSport, sport)
	}
	return config, nil
}
