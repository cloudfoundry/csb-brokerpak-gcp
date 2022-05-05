## Release notes for next release:

### New feature:
- PostgreSQL is no longer in beta, and can be used in production
- BREAKING: Due to new features in the PostgreSQL service offering, it is not possible to upgrade from
  a previous (Beta) version to this version. You should either delete existing PostgreSQL instances before upgrade, or
  run "cf purge-service-instance" on them to remove them from CloudFoundry management.
- PostgreSQL backups are enabled by default and can be configured via the new `backups_retain_number`, `backups_location`, `backups_start_time` and `backups_point_in_time_log_retain_days` properties
- PostgreSQL password stored using `scram-sha-256` for additional security
- PostgreSQL properties can now be updated: cores, storage_gb, credentials, authorized_network, authorized_network_id, authorized_networks_cidrs, public_ip
- PostgreSQL connections must be via TLS
- Google SQL service tiers are now exposed when provisioning, or updating an instance. The previous 'cores' abstraction has been removed, in favour of using the underlying Google tier.


### Fix:
- minimum constraints on MySQL, PostreSQL, and Spanner storage_gb are now enforced

