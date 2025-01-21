package players

func getStatPchange(controlStat float32, newStat float32) float32 {
    return (newStat - controlStat) / controlStat
}

type PlayerAvg interface {
    IsValid() bool
    AddAvg(PlayerAvg) PlayerAvg
    CompareAvg(PlayerAvg) PlayerAvg
    PredictStats(PlayerAvg) PlayerAvg
    ConvertToPer() PlayerAvg
    ConvertToStats() PlayerAvg
    GetStats() map[string]float32
}

type NBAAvg struct {
    NumGames     int         `json:"num_minutes"`
    Minutes      float32     `json:"avg_minutes"`
    Points       float32     `json:"avg_points"`
    Rebounds     float32     `json:"avg_rebounds"`
    Assists      float32     `json:"avg_assists"`
    Threes       float32     `json:"avg_threes"`
    Usg          float32     `json:"avg_usg"`
    Ortg         float32     `json:"avg_drtg"`
    Drtg         float32     `json:"avg_ortg"`
}

func (n NBAAvg) IsValid() bool {
    return n.NumGames > 0
}

func (n NBAAvg) GetStats() map[string]float32 {
    return map[string]float32{
        "minutes": n.Minutes,
        "points": n.Points,
        "rebounds": n.Rebounds,
        "assists": n.Assists,
        "threes": n.Threes,
        "usg": n.Usg,
        "ortg": n.Ortg,
        "drtg": n.Drtg,
    }
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
        Threes: (n.Threes * float32(n.NumGames) + nba.Threes * float32(nba.NumGames)) / total_games,
        Usg: (n.Usg * float32(n.NumGames) + nba.Usg * float32(nba.NumGames)) / total_games,
        Ortg: (n.Ortg * float32(n.NumGames) + nba.Ortg * float32(nba.NumGames)) / total_games,
        Drtg: (n.Drtg * float32(n.NumGames) + nba.Drtg * float32(nba.NumGames)) / total_games,
    }
}

func (n NBAAvg) CompareAvg(controlAvg PlayerAvg) PlayerAvg {
    if !n.IsValid() {
        return n
    }
    nbaControl := controlAvg.(NBAAvg)
    return NBAAvg{
        NumGames: n.NumGames,
        Minutes: getStatPchange(nbaControl.Minutes, n.Minutes),
        Points: getStatPchange(nbaControl.Points, n.Points),
        Rebounds: getStatPchange(nbaControl.Rebounds, n.Rebounds),
        Assists: getStatPchange(nbaControl.Assists, n.Assists),
        Threes: getStatPchange(nbaControl.Threes, n.Threes),
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
            Threes: n.Threes / n.Minutes,
            Usg: n.Usg / n.Minutes,
            Ortg: n.Ortg / n.Minutes,
            Drtg: n.Drtg / n.Minutes,
        }
    } else {
        return n
    }
}

func (n NBAAvg) ConvertToStats() PlayerAvg { 
    if n.IsValid() {
        return NBAAvg{
            NumGames: n.NumGames,
            Minutes: n.Minutes,
            Points: n.Points * n.Minutes,
            Rebounds: n.Rebounds * n.Minutes,
            Assists: n.Assists * n.Minutes,
            Threes: n.Threes * n.Minutes,
            Usg: n.Usg * n.Minutes,
            Ortg: n.Ortg * n.Minutes,
            Drtg: n.Drtg * n.Minutes,
        }
    } else {
        return n
    }
}

func (n NBAAvg) PredictStats(pipFactor PlayerAvg) PlayerAvg {
    nbaPip := pipFactor.(NBAAvg)
    predictedMinutes := n.Minutes + n.Minutes * nbaPip.Minutes

    return NBAAvg{
        NumGames: nbaPip.NumGames,
        Minutes: predictedMinutes,
        Points: (n.Points + n.Points * nbaPip.Points) * predictedMinutes,
        Rebounds: (n.Rebounds + n.Rebounds * nbaPip.Rebounds) * predictedMinutes,
        Assists: (n.Assists + n.Assists * nbaPip.Assists) * predictedMinutes,
        Threes: (n.Threes + n.Threes * nbaPip.Threes) * predictedMinutes,
        Usg: (n.Usg + n.Usg * nbaPip.Usg) * predictedMinutes,
        Ortg: (n.Ortg + n.Ortg * nbaPip.Ortg) * predictedMinutes,
        Drtg: (n.Drtg + n.Drtg * nbaPip.Drtg) * predictedMinutes,
    }
}
