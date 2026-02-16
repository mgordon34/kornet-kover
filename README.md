# Kornet Kover

Kornet Kover is a full-stack sports analytics and prop-betting research platform. It ingests historical and live sportsbook odds, game and player data across multiple sports (NBA, MLB, WNBA), runs statistical analysis and backtests betting strategies, and exposes the results via a Go API with a Next.js frontend.

This repo is organized as a Go backend for data ingestion, modeling, and analysis, and a modern React/Next.js frontend for exploration and visualization.

## What This Does

- Scrapes and normalizes game, roster, player, and odds data from sportsbooks and public sources
- Stores historical data in a relational database
- Runs player prop analysis and strategy backtests
- Exposes JSON APIs for picks, odds, games, players, and strategies
- Provides a frontend UI for interacting with picks and strategy outputs

This project is research-focused and intended for analysis and experimentation, not automated wagering.

## Repository Structure

- `backend/` Go API, data ingestion, analysis, and backtesting
- `backend/api/` HTTP controllers and API models
- `backend/internal/analysis/` prop selection logic, predictors, and analysis routines
- `backend/internal/backtesting/` historical backtesting engine
- `backend/internal/scraper/` data scraping and update jobs
- `backend/internal/sports/` data-first sport config registry and provider interfaces (NBA, MLB, WNBA)
- `backend/internal/sportsbook/` sportsbook odds ingestion and prop handling
- `backend/internal/storage/` database initialization and access
- `frontend/` Next.js app with Tailwind, Prisma, and UI components
- `openspec/` design docs, proposals, and longer-term planning

## Tech Stack

Backend
- Go
- Gin (HTTP server)
- SQLite / relational DB (via internal storage layer)
- Custom statistical models and backtesting engine

Frontend
- Next.js (App Router)
- TypeScript
- Tailwind CSS
- Prisma

## Running the Backend

Requirements:
- Go 1.21+

From the repo root:

```bash
cd backend
go run .
```

The API server starts on `http://localhost:8080`.

Key routes:
- `GET /update-games` refresh games for supported sports
- `GET /update-players` refresh active rosters
- `GET /update-lines` refresh sportsbook odds
- `GET /pick-props` run prop analysis
- `GET /strategies` list configured strategies
- `GET /prop-picks` return generated prop picks

Some update and backtest routines are intentionally commented out in `backend/main.go` and can be run manually for research or experimentation.

## Running the Frontend

Requirements:
- Node.js 18+

From the repo root:

```bash
cd frontend
npm install
npm run dev
```

The frontend runs at `http://localhost:3000` and is configured to talk to the local backend.

## Environment Configuration

Environment variables are expected via `.env` files (not committed):

- `backend/.env` database and scraper configuration
- `frontend/.env.development` frontend-specific config

See existing `.envrc` and example files for expected values.

## Development Notes

- The backend is intentionally modular; most business logic lives under `backend/internal/`
- Sport-specific config is provided through `sports.ConfigProvider`; services should take it via constructor deps instead of package-global lookups
- Strategy definitions are code-driven via `analysis.PropSelector`
- Backtests are reproducible by date range and strategy configuration
- The frontend is still evolving and primarily focused on internal tooling

## Disclaimer

This project is for educational and analytical purposes only. It does not place bets or provide gambling advice.
