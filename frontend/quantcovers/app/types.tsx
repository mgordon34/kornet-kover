export interface Strategy {
  id: number;
  name: string;
}

export interface PropPick {
  id: number;
  player_name: string;
  num_games: number;
  side: string;
  stat: string;
  line: number;
  points: number;
  rebounds: number;
  assists: number;
}
