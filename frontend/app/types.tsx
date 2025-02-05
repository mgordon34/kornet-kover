export interface Strategy {
  id: number;
  name: string;
}

export interface PropPick {
  id: number;
  strat_id: number;
  player_name: string;
  num_games: number;
  side: string;
  stat: string;
  line: number;
  points: number;
  rebounds: number;
  assists: number;
  threes: number;
}

export interface StrategyPicks {
  strat_id: number;
  strat_name: string;
  picks: PropPick[];
}
