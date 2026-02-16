# Change: Simplify sport config initialization with a data registry and provider injection

## Why
Sport configuration currently uses per-sport constructor types, an interface layer, and a mutable global cache even though the values are mostly static data. This adds indirection, introduces nil/caching edge cases, and does not align cleanly with the newer service-struct pattern that favors explicit dependency injection.

## What Changes
- Replace per-sport config constructors and cache-driven initialization with a data-first registry keyed by `sports.Sport`.
- Introduce a `sports` config provider interface for service consumption, with a default provider backed by the registry.
- Update service-struct call sites that depend on sport config (`scraper`, `sportsbook`) to consume injected provider dependencies instead of package-global config access.
- Keep compatibility wrappers in `internal/sports` during migration so existing callers can transition incrementally.
- Add/adjust tests to verify supported/unsupported sport lookup behavior and injected-provider behavior in consuming services.

## Impact
- Affected specs: `sport-config-provider` (new)
- Affected code:
  - `backend/internal/sports/*`
  - `backend/internal/scraper/*`
  - `backend/internal/sportsbook/*`
  - `backend/main.go`
