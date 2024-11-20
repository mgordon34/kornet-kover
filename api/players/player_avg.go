package players

func getStatPchange(controlStat float32, newStat float32) float32 {
    return (newStat - controlStat) / controlStat
}

type PlayerAvg interface {
    IsValid() bool
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
