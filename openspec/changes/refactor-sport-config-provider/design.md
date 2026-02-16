## Context
The current sport config package models each sport as a constructor-backed config type and caches instances in package-global mutable state. In practice, these configs are static data tables consumed by orchestration services. As backend migration moves toward service structs with explicit dependencies, the current pattern creates avoidable complexity and weaker dependency boundaries.

## Goals / Non-Goals
- Goals:
  - Represent sport-specific configuration as plain data keyed by sport.
  - Eliminate unnecessary constructor/caching indirection in config access.
  - Make sport config consumption explicit via injected provider dependencies in services.
  - Preserve existing runtime behavior while migrating call sites incrementally.
- Non-Goals:
  - Change sportsbook/scraper/analysis product behavior or endpoint contracts.
  - Introduce a DI framework or broad architecture rewrite beyond config access.
  - Add new sport domains as part of this refactor.

## Decisions
- Decision: Replace per-sport constructor types with a registry map keyed by `sports.Sport` containing sportsbook/scraper/analysis config bundles.
- Decision: Introduce a small provider interface for config lookup and inject it into service structs that require sport config.
- Decision: Return explicit lookup errors for unsupported sports instead of relying on nil propagation.
- Decision: Keep temporary wrapper functions to reduce migration risk, then remove wrappers once all call sites are migrated.
- Alternatives considered:
  - Keep current constructor + cache model (rejected: extra abstraction and mutable global state for static data).
  - Embed config retrieval directly in each service (rejected: duplicates config source of truth and weakens consistency).

## Target Pattern
- `internal/sports` is the single source of truth for sport config data.
- Services depend on a `ConfigProvider` abstraction (constructor-injected), not package-global lookups.
- Tests provide fake providers directly to service constructors for deterministic behavior checks.

## Risks / Trade-offs
- Risk: Temporary duplication while wrappers and provider APIs coexist.
  - Mitigation: Track wrapper removal as an explicit follow-up task.
- Risk: Interface updates in service dependencies may require broad test updates.
  - Mitigation: Migrate by package and preserve constructor defaults for compatibility.
- Risk: Unsupported-sport behavior changes from implicit nil handling to explicit error paths.
  - Mitigation: Add tests that codify expected unsupported-sport behavior before rollout.

## Migration Plan
1. Build registry + provider in `internal/sports` and retain wrappers.
2. Inject provider into `internal/scraper` and `internal/sportsbook` dependencies with constructor defaults.
3. Update composition-root wiring and tests.
4. Remove deprecated wrappers after call-site migration is complete.
