## Release notes for next release:

## Features:

- Service instances can be upgraded to latest TF v 1.2.3

## Fixes:

- Adds lifecycle.prevent_destroy to all data services to provide extra layer of protection against data loss
- Adds prohibit_update property to avoid updating region in BigQuery and Storage services because it can result in the
  recreation of the service instance and lost data.
