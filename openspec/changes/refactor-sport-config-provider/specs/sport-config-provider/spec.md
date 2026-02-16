## ADDED Requirements

### Requirement: Data-first sport config registry
The system SHALL store sport-specific sportsbook, scraper, and analysis configuration in a data-first registry keyed by sport.

#### Scenario: Supported sport config is requested
- **WHEN** a caller requests configuration for a supported sport
- **THEN** the registry returns the configured values for that sport
- **AND** the values are sourced from the shared registry definition

#### Scenario: Unsupported sport config is requested
- **WHEN** a caller requests configuration for an unsupported sport
- **THEN** the lookup returns an explicit error indicating the sport is unsupported

### Requirement: Injectable sport config provider for services
Backend services that consume sport configuration SHALL depend on an injected sport config provider abstraction rather than package-global config retrieval.

#### Scenario: Service is created without explicit provider
- **WHEN** a service constructor is called without a custom config provider
- **THEN** the constructor injects the default registry-backed provider

#### Scenario: Service tests supply fake provider
- **WHEN** a test constructs a service with a fake config provider
- **THEN** the service uses the injected fake provider for config retrieval
- **AND** test behavior does not require process-global config mutation

### Requirement: Incremental compatibility for existing callers
The migration SHALL allow existing in-repo callers to continue working while provider injection is rolled out incrementally.

#### Scenario: Legacy helper function is used during migration
- **WHEN** an existing caller uses a legacy helper in `internal/sports`
- **THEN** the helper delegates to the new provider/registry path
- **AND** behavior remains compatible until the caller is migrated
