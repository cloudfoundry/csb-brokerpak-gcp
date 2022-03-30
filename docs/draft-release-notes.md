## Release notes for next release:

### Breaking changes:
- The default plans for PostgreSQL have been removed. In order to successfully deploy a broker, plans must be defined via the `GSB_SERVICE_CSB_GOOGLE_POSTGRES_PLANS` environment variable.
- Updating instance properties is not supported for PostgreSQL
- Default version os PostgreSQL has been changed to 13, to match GCP defaults
- Bind users are now deleted on unbind operation and the ownership of the objects they created is passed on to a "provision_user". As a result bindings created with previous versions cannot longer be managed by the broker. We recommend deleting the bindings before upgrading. Also, if you would like to continue managing previously created service instances you would need to update them before doing any other operation. 

### New feature:
- The `BROKERPAK_UPDATES_ENABLED` feature flag is always turned on, so the HCL used when managing a service instance is always the latest taken from the brokerpak.
- Beta tagged services are now disabled by CSB by default. To enable these services the following env var must be set: `GSB_COMPATIBILITY_ENABLE_BETA_SERVICES=true`
- Public IPs can be assigned to Postgresql databases on creation. This can be enabled by setting the `public_ip` parameter which defaults to `false`
- List of IP addresses can now be specified to allow connections to a Postgresql database by setting the `authorized_networks_cidrs` parameter
- This repo how has a go.mod file at the top level which contains the version of Cloud Service Broker that is compatible with this release.
Scripts such as make push-broker will use this version rather than always using the very latest Cloud Service Broker.

### Fix:
- All service offerings and plans are highlighted to be at a Beta lifecycle state.
- Fixed typo `defualt` => `default` in the postgres service, now the size of storage volume should default to 10 GB
