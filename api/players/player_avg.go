package players

func getStatPchange(controlStat float32, newStat float32) float32 {
    return (newStat - controlStat) / controlStat
}

type PlayerAvg interface {
    IsValid() bool
    AddAvg(PlayerAvg) PlayerAvg
    CompareAvg(PlayerAvg) PlayerAvg
    ConvertToPer() PlayerAvg
}

type NBAAvg struct {
    NumGames     int         `json:"num_minutes"`
    Minutes      float32     `json:"avg_minutes"`
    Points       float32     `json:"avg_points"`
    Rebounds     float32     `json:"avg_rebounds"`
    Assists      float32     `json:"avg_assists"`
    Usg          float32     `json:"avg_usg"`
    Ortg         float32     `json:"avg_drtg"`
    Drtg         float32     `json:"avg_ortg"`
}

func (n NBAAvg) IsValid() bool {
    return n.NumGames > 0
}

func (n NBAAvg) AddAvg(a PlayerAvg) PlayerAvg {
    if !a.IsValid(){
        return n
    }
    nba := a.(NBAAvg)
    total_games := float32(n.NumGames + nba.NumGames)
    return NBAAvg{
        NumGames: n.NumGames + nba.NumGames,
        Minutes: (n.Minutes * float32(n.NumGames) + nba.Minutes * float32(nba.NumGames)) / total_games,
        Points: (n.Points * float32(n.NumGames) + nba.Points * float32(nba.NumGames)) / total_games,
        Rebounds: (n.Rebounds * float32(n.NumGames) + nba.Rebounds * float32(nba.NumGames)) / total_games,
        Assists: (n.Assists * float32(n.NumGames) + nba.Assists * float32(nba.NumGames)) / total_games,
        Usg: (n.Usg * float32(n.NumGames) + nba.Usg * float32(nba.NumGames)) / total_games,
        Ortg: (n.Ortg * float32(n.NumGames) + nba.Ortg * float32(nba.NumGames)) / total_games,
        Drtg: (n.Drtg * float32(n.NumGames) + nba.Drtg * float32(nba.NumGames)) / total_games,
    }
}

func (n NBAAvg) CompareAvg(controlAvg PlayerAvg) PlayerAvg {
    nbaControl := controlAvg.(NBAAvg)
    return NBAAvg{
        NumGames: n.NumGames,
        Minutes: getStatPchange(nbaControl.Minutes, n.Minutes),
        Points: getStatPchange(nbaControl.Points, n.Points),
        Rebounds: getStatPchange(nbaControl.Rebounds, n.Rebounds),
        Assists: getStatPchange(nbaControl.Assists, n.Assists),
        Usg: getStatPchange(nbaControl.Usg, n.Usg),
        Ortg: getStatPchange(nbaControl.Ortg, n.Ortg),
        Drtg: getStatPchange(nbaControl.Drtg, n.Drtg),
    }
}

func (n NBAAvg) ConvertToPer() PlayerAvg { 
    if n.IsValid() {
        return NBAAvg{
            NumGames: n.NumGames,
            Minutes: n.Minutes,
            Points: n.Points / n.Minutes,
            Rebounds: n.Rebounds / n.Minutes,
            Assists: n.Assists / n.Minutes,
            Usg: n.Usg / n.Minutes,
            Ortg: n.Ortg / n.Minutes,
            Drtg: n.Drtg / n.Minutes,
        }
    } else {
        return n
    }
}
