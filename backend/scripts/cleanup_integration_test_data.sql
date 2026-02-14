-- Cleanup script for backend integration-test records (safe mode).
--
-- Usage example:
--   psql "$DB_URL" -f backend/scripts/cleanup_integration_test_data.sql
--
-- This script targets only known integration-test markers/patterns used in
-- *_integration_test.go files.
--
-- NOTE: In production-sized databases, deleting from `players`/`teams` can be
-- slow due to FK checks against very large referencing tables. This safe-mode
-- script deletes dependent test data but leaves marker players/teams in place.

BEGIN;

-- 1) Users/strategies created by integration tests
CREATE TEMP TABLE it_users ON COMMIT DROP AS
SELECT id
FROM users
WHERE email IN ('test-user@example.com', 'picks-user@example.com');

CREATE TEMP TABLE it_strategies ON COMMIT DROP AS
SELECT id
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

DELETE FROM prop_picks
WHERE strat_id IN (SELECT id FROM it_strategies);

DELETE FROM strategy_filters
WHERE id IN (SELECT id FROM it_strategy_filters);

DELETE FROM strategies
WHERE id IN (SELECT id FROM it_strategies);

DELETE FROM users
WHERE id IN (SELECT id FROM it_users);

-- 2) Team/player markers used by integration tests
CREATE TEMP TABLE it_teams ON COMMIT DROP AS
SELECT index
FROM teams
WHERE index IN ('TSTH', 'TSTA', 'ITM1', 'ITM2', 'PKH')
   OR index ~ '^(H|A)[0-9]{6}$'
   OR index ~ '^MLB_(H|A)[0-9]{6}$';

CREATE TEMP TABLE it_players ON COMMIT DROP AS
SELECT index
FROM players
WHERE index IN ('oddsit01', 'picksit01', 'mlbtest01')
   OR index ~ '^(sfx|nbaa|nbab|nbac|mlbb|mlbp|new)[0-9]{6}$';

-- 3) Game/line records tied to integration markers
CREATE TEMP TABLE it_games ON COMMIT DROP AS
SELECT id
FROM games
WHERE home_index IN (SELECT index FROM it_teams)
   OR away_index IN (SELECT index FROM it_teams)
   OR date >= DATE '2099-01-01';

CREATE TEMP TABLE it_lines ON COMMIT DROP AS
SELECT id
FROM player_lines
WHERE player_index IN (SELECT index FROM it_players)
   OR (timestamp::date >= DATE '2099-01-01' AND player_index IN ('oddsit01', 'picksit01'));

-- 4) Remove dependents first (FK-safe order)
DELETE FROM prop_picks
WHERE line_id IN (SELECT id FROM it_lines);

DELETE FROM mlb_play_by_plays
WHERE game IN (SELECT id FROM it_games)
   OR batter_index IN (SELECT index FROM it_players)
   OR pitcher_index IN (SELECT index FROM it_players);

DELETE FROM mlb_player_games_pitching
WHERE game IN (SELECT id FROM it_games)
   OR player_index IN (SELECT index FROM it_players);

DELETE FROM mlb_player_games_batting
WHERE game IN (SELECT id FROM it_games)
   OR player_index IN (SELECT index FROM it_players);

DELETE FROM wnba_player_games
WHERE game IN (SELECT id FROM it_games)
   OR player_index IN (SELECT index FROM it_players);

DELETE FROM nba_player_games
WHERE game IN (SELECT id FROM it_games)
   OR player_index IN (SELECT index FROM it_players);

DELETE FROM nba_pip_predictions
WHERE player_index IN (SELECT index FROM it_players)
   OR date >= DATE '2099-01-01';

DELETE FROM wnba_pip_predictions
WHERE player_index IN (SELECT index FROM it_players)
   OR date >= DATE '2099-01-01';

DELETE FROM nba_pip_factors
WHERE player_index IN (SELECT index FROM it_players)
   OR other_index IN (SELECT index FROM it_players);

DELETE FROM active_rosters
WHERE player_index IN (SELECT index FROM it_players)
   OR team_index IN (SELECT index FROM it_teams)
   OR last_updated >= DATE '2099-01-01';

DELETE FROM player_lines
WHERE id IN (SELECT id FROM it_lines);

DELETE FROM games
WHERE id IN (SELECT id FROM it_games);

-- Intentionally skipped in safe mode to avoid long FK scans on large DBs:
DELETE FROM players WHERE index IN (SELECT index FROM it_players);
DELETE FROM teams   WHERE index IN (SELECT index FROM it_teams);

COMMIT;
