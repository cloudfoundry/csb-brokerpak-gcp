## Release notes for next release:

## Features:

- The default TF version is updated to v1.2.4. Service instances can be upgraded to the latest version. 
  - Notes: If you have Redis or Dataproc instances created without defining `instance_id` for Redis or `name` for Dataproc you have to update those instances before installing this version. 
  For Redis instances run `cf update-service your-redis-si -c '{"instance_id":"<current-name-from-GCP-console>"}'`. 
  For Dataproc instances run `cf update-service your-dataproc-si -c '{"name":"<current-name-from-GCP-console>"}'`.

## Fixes:

- Adds lifecycle.prevent_destroy to all data services to provide extra layer of protection against data loss.
- Adds prohibit_update property to avoid updating region in BigQuery and Storage services because it can result in the
  recreation of the service instance and lost data.
- Redis and Dataproc names in the GCP console now rely on the request instance ID. It was previously relying on a
  timestamp that was causing updates to destroy the instance.
