## 1. Sports config registry refactor
- [x] 1.1 Define a data-first sport config registry keyed by `sports.Sport` for sportsbook, scraper, and analysis config values.
- [x] 1.2 Replace constructor/cache-driven lookup paths with typed registry lookups that return explicit errors for unsupported sports.
- [x] 1.3 Migrate all in-repo callers to provider-backed lookups (no legacy helper wrappers required).

## 2. Service-struct integration
- [x] 2.1 Add an injected sport-config provider dependency to services that consume sport config (starting with `internal/scraper` and `internal/sportsbook`).
- [x] 2.2 Update default dependency wiring so constructors use the default provider when none is injected.
- [x] 2.3 Update composition-root wiring in `backend/main.go` to construct services with explicit config-provider dependencies.

## 3. Test updates and behavior verification
- [x] 3.1 Replace sports package tests that assert constructor/cache behavior with tests that assert registry lookup and unsupported-sport handling.
- [x] 3.2 Update scraper and sportsbook tests to verify they consume injected config providers (no package-global mutation required).
- [x] 3.3 Run backend tests for touched packages and full backend test suite (`go test ./...`) and resolve regressions.

## 4. Migration cleanup follow-up
- [x] 4.1 Remove deprecated compatibility wrappers after all in-repo call sites are migrated.
- [x] 4.2 Document the final sport-config provider pattern in backend developer docs.
