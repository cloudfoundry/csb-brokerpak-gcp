## Release notes for next release:

### New feature:
- PostgreSQL backups are enabled by default and can be configured via the new `backups_retain_number`, `backups_location`, `backups_start_time` and `backups_point_in_time_log_retain_days` properties
- PostgreSQL password stored using `scram-sha-256` for additional security

### Fix:
- minimum constraints on MySQL, PostreSQL, and Spanner storage_gb are now enforced

