## Release notes for next release:

### Breaking changes:
- The default plans for PostgreSQL have been removed. In order to successfully deploy a broker, plans must be defined via the `GSB_SERVICE_CSB_GOOGLE_POSTGRES_PLANS` environment variable.
- Updating instance properties is not supported for postgres

### New feature:
- The `BROKERPAK_UPDATES_ENABLED` feature flag is always turned on, so the HCL used when managing a service instance is always the latest taken from the brokerpak.
- Beta tagged services are now disabled by CSB by default. To enable these services the following env var must be set: `GSB_COMPATIBILITY_ENABLE_BETA_SERVICES=true`

### Fix:
- All service offerings and plans are highlighted to be at a Beta lifecycle state.
- Fixed typo `defualt` => `default` in the postgres service, now the size of storage volume should default to 10 GB
