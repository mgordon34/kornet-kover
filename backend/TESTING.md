# Backend Testing Guide

## Run unit tests

```bash
go test ./...
```

Integration tests are separated with the `integration` build tag and are excluded from this default command.

## Run integration tests

```bash
go test -tags=integration ./...
```

Integration tests may write to the database. Use a dedicated test database via `DB_URL`.

## Run unit tests with coverage

```bash
go test ./... -cover
```

## Enforce per-package coverage thresholds

```bash
./scripts/check_coverage.sh
```

By default each package is expected to reach `100.0` coverage. Temporary exceptions are tracked in `.coverage-exceptions` and must include an owner and expiration comment.

## Adding new tests

- Prefer pure-function tests first.
- Isolate DB/network/time dependencies behind small seams when logic is hard to test.
- Add regression tests for every bug fix.
