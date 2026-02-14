-- Dry-run preview for backend integration-test cleanup.
--
-- Usage example:
--   psql "$DB_URL" -f backend/scripts/cleanup_integration_test_data_dry_run.sql
--
-- This script DOES NOT delete data. It only shows what would be deleted by
-- cleanup_integration_test_data.sql.
--
-- NOTE: The paired cleanup script runs in safe mode and does NOT delete rows
-- from `players` or `teams` to avoid expensive FK validation scans.

BEGIN;

-- 1) Identify strategy/user test records
CREATE TEMP TABLE it_users ON COMMIT DROP AS
SELECT id, email, name
FROM users
WHERE email IN ('test-user@example.com', 'picks-user@example.com');

CREATE TEMP TABLE it_strategies ON COMMIT DROP AS
SELECT id, user_id, name
FROM strategies
WHERE user_id IN (SELECT id FROM it_users);

CREATE TEMP TABLE it_strategy_filters ON COMMIT DROP AS
SELECT id
FROM strategy_filters
WHERE false;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = current_schema()
      AND table_name = 'strategy_filters'
      AND column_name = 'strategy_id'
  ) THEN
    EXECUTE 'INSERT INTO it_strategy_filters (id) SELECT id FROM strategy_filters WHERE strategy_id IN (SELECT id FROM it_strategies)';
  ELSIF EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = current_schema()
      AND table_name = 'strategy_filters'
      AND column_name = 'stategy_id'
  ) THEN
    EXECUTE 'INSERT INTO it_strategy_filters (id) SELECT id FROM strategy_filters WHERE stategy_id IN (SELECT id FROM it_strategies)';
  END IF;
END $$;

-- 2) Identify team/player test markers
CREATE TEMP TABLE it_teams ON COMMIT DROP AS
SELECT index, name
FROM teams
WHERE index IN ('TSTH', 'TSTA', 'ITM1', 'ITM2', 'PKH')
   OR index ~ '^(H|A)[0-9]{6}$'
   OR index ~ '^MLB_(H|A)[0-9]{6}$';

CREATE TEMP TABLE it_players ON COMMIT DROP AS
SELECT index, sport, name
FROM players
WHERE index IN ('oddsit01', 'picksit01', 'mlbtest01')
   OR index ~ '^(sfx|nbaa|nbab|nbac|mlbb|mlbp|new)[0-9]{6}$';

-- 3) Identify game/line records tied to integration markers
CREATE TEMP TABLE it_games ON COMMIT DROP AS
SELECT id, sport, home_index, away_index, date
FROM games
WHERE home_index IN (SELECT index FROM it_teams)
   OR away_index IN (SELECT index FROM it_teams)
   OR date >= DATE '2099-01-01';

CREATE TEMP TABLE it_lines ON COMMIT DROP AS
SELECT id, sport, player_index, timestamp, stat, side, type, line, odds
FROM player_lines
WHERE player_index IN (SELECT index FROM it_players)
   OR (timestamp::date >= DATE '2099-01-01' AND player_index IN ('oddsit01', 'picksit01'));

-- Summary counts (what would be deleted)
SELECT 'users' AS table_name, COUNT(*) AS rows_to_delete FROM users WHERE id IN (SELECT id FROM it_users)
UNION ALL
SELECT 'strategies', COUNT(*) FROM strategies WHERE id IN (SELECT id FROM it_strategies)
UNION ALL
SELECT 'strategy_filters', COUNT(*) FROM strategy_filters WHERE id IN (SELECT id FROM it_strategy_filters)
UNION ALL
SELECT 'prop_picks(strat_id)', COUNT(*) FROM prop_picks WHERE strat_id IN (SELECT id FROM it_strategies)
UNION ALL
SELECT 'games', COUNT(*) FROM games WHERE id IN (SELECT id FROM it_games)
UNION ALL
SELECT 'player_lines', COUNT(*) FROM player_lines WHERE id IN (SELECT id FROM it_lines)
UNION ALL
SELECT 'prop_picks(line_id)', COUNT(*) FROM prop_picks WHERE line_id IN (SELECT id FROM it_lines)
UNION ALL
SELECT 'nba_player_games', COUNT(*) FROM nba_player_games WHERE game IN (SELECT id FROM it_games) OR player_index IN (SELECT index FROM it_players)
UNION ALL
SELECT 'wnba_player_games', COUNT(*) FROM wnba_player_games WHERE game IN (SELECT id FROM it_games) OR player_index IN (SELECT index FROM it_players)
UNION ALL
SELECT 'mlb_player_games_batting', COUNT(*) FROM mlb_player_games_batting WHERE game IN (SELECT id FROM it_games) OR player_index IN (SELECT index FROM it_players)
UNION ALL
SELECT 'mlb_player_games_pitching', COUNT(*) FROM mlb_player_games_pitching WHERE game IN (SELECT id FROM it_games) OR player_index IN (SELECT index FROM it_players)
UNION ALL
SELECT 'mlb_play_by_plays', COUNT(*) FROM mlb_play_by_plays WHERE game IN (SELECT id FROM it_games) OR batter_index IN (SELECT index FROM it_players) OR pitcher_index IN (SELECT index FROM it_players)
UNION ALL
SELECT 'nba_pip_predictions', COUNT(*) FROM nba_pip_predictions WHERE player_index IN (SELECT index FROM it_players) OR date >= DATE '2099-01-01'
UNION ALL
SELECT 'wnba_pip_predictions', COUNT(*) FROM wnba_pip_predictions WHERE player_index IN (SELECT index FROM it_players) OR date >= DATE '2099-01-01'
UNION ALL
SELECT 'nba_pip_factors', COUNT(*) FROM nba_pip_factors WHERE player_index IN (SELECT index FROM it_players) OR other_index IN (SELECT index FROM it_players)
UNION ALL
SELECT 'active_rosters', COUNT(*) FROM active_rosters WHERE player_index IN (SELECT index FROM it_players) OR team_index IN (SELECT index FROM it_teams) OR last_updated >= DATE '2099-01-01';

-- Marker entities found but not deleted by safe-mode cleanup script
SELECT 'teams_not_deleted_in_safe_mode' AS table_name, COUNT(*) AS rows_matched
FROM teams
WHERE index IN (SELECT index FROM it_teams)
UNION ALL
SELECT 'players_not_deleted_in_safe_mode', COUNT(*)
FROM players
WHERE index IN (SELECT index FROM it_players);

-- Detailed record previews (remove sections if too verbose)
SELECT 'users' AS section, row_to_json(u) AS record
FROM (SELECT * FROM users WHERE id IN (SELECT id FROM it_users) ORDER BY id) u;

SELECT 'strategies' AS section, row_to_json(s) AS record
FROM (SELECT * FROM strategies WHERE id IN (SELECT id FROM it_strategies) ORDER BY id) s;

SELECT 'strategy_filters' AS section, row_to_json(sf) AS record
FROM (SELECT * FROM strategy_filters WHERE id IN (SELECT id FROM it_strategy_filters) ORDER BY id) sf;

SELECT 'prop_picks_by_strategy' AS section, row_to_json(pp) AS record
FROM (SELECT * FROM prop_picks WHERE strat_id IN (SELECT id FROM it_strategies) ORDER BY id) pp;

SELECT 'teams_not_deleted_in_safe_mode' AS section, row_to_json(t) AS record
FROM (SELECT * FROM teams WHERE index IN (SELECT index FROM it_teams) ORDER BY index) t;

SELECT 'players_not_deleted_in_safe_mode' AS section, row_to_json(p) AS record
FROM (SELECT * FROM players WHERE index IN (SELECT index FROM it_players) ORDER BY index) p;

SELECT 'games' AS section, row_to_json(g) AS record
FROM (SELECT * FROM games WHERE id IN (SELECT id FROM it_games) ORDER BY id) g;

SELECT 'player_lines' AS section, row_to_json(pl) AS record
FROM (SELECT * FROM player_lines WHERE id IN (SELECT id FROM it_lines) ORDER BY id) pl;

SELECT 'prop_picks_by_line' AS section, row_to_json(pp) AS record
FROM (SELECT * FROM prop_picks WHERE line_id IN (SELECT id FROM it_lines) ORDER BY id) pp;

SELECT 'nba_player_games' AS section, row_to_json(npg) AS record
FROM (
  SELECT * FROM nba_player_games
  WHERE game IN (SELECT id FROM it_games) OR player_index IN (SELECT index FROM it_players)
  ORDER BY id
) npg;

SELECT 'wnba_player_games' AS section, row_to_json(wpg) AS record
FROM (
  SELECT * FROM wnba_player_games
  WHERE game IN (SELECT id FROM it_games) OR player_index IN (SELECT index FROM it_players)
  ORDER BY id
) wpg;

SELECT 'mlb_player_games_batting' AS section, row_to_json(mb) AS record
FROM (
  SELECT * FROM mlb_player_games_batting
  WHERE game IN (SELECT id FROM it_games) OR player_index IN (SELECT index FROM it_players)
  ORDER BY id
) mb;

SELECT 'mlb_player_games_pitching' AS section, row_to_json(mp) AS record
FROM (
  SELECT * FROM mlb_player_games_pitching
  WHERE game IN (SELECT id FROM it_games) OR player_index IN (SELECT index FROM it_players)
  ORDER BY id
) mp;

SELECT 'mlb_play_by_plays' AS section, row_to_json(mpbp) AS record
FROM (
  SELECT * FROM mlb_play_by_plays
  WHERE game IN (SELECT id FROM it_games)
     OR batter_index IN (SELECT index FROM it_players)
     OR pitcher_index IN (SELECT index FROM it_players)
  ORDER BY id
) mpbp;

SELECT 'nba_pip_predictions' AS section, row_to_json(np) AS record
FROM (
  SELECT * FROM nba_pip_predictions
  WHERE player_index IN (SELECT index FROM it_players) OR date >= DATE '2099-01-01'
  ORDER BY id
) np;

SELECT 'wnba_pip_predictions' AS section, row_to_json(wp) AS record
FROM (
  SELECT * FROM wnba_pip_predictions
  WHERE player_index IN (SELECT index FROM it_players) OR date >= DATE '2099-01-01'
  ORDER BY id
) wp;

SELECT 'nba_pip_factors' AS section, row_to_json(nf) AS record
FROM (
  SELECT * FROM nba_pip_factors
  WHERE player_index IN (SELECT index FROM it_players)
     OR other_index IN (SELECT index FROM it_players)
  ORDER BY id
) nf;

SELECT 'active_rosters' AS section, row_to_json(ar) AS record
FROM (
  SELECT * FROM active_rosters
  WHERE player_index IN (SELECT index FROM it_players)
     OR team_index IN (SELECT index FROM it_teams)
     OR last_updated >= DATE '2099-01-01'
  ORDER BY id
) ar;

ROLLBACK;
