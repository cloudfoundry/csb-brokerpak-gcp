## Release notes for next release:

## Fixes:
- adds lifecycle.prevent_destroy to all data services to provide extra layer of protection against data loss
- adds prohibit_update property to avoid updating region in BigQuery and Storage services because it can result in the recreation of the service instance and lost data.

