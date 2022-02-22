## Release notes for next release:

### Breaking changes:
- The default plans for PostgreSQL have been removed. In order to successfully deploy a broker, plans must be defined via the `GSB_SERVICE_CSB_GOOGLE_POSTGRES_PLANS` environment variable.

### New feature:
- The `BROKERPAK_UPDATES_ENABLED` feature flag is always turned on, so the HCL used when managing a service instance is always the latest taken from the brokerpak.

### Fix:
- All service offerings and plans are highlighted to be at a Beta lifecycle state.
