# Billing

Cloud Service Broker for GCP automatically labels supported resources with organization GUID, space GUID and instance ID.

When these supported services are provisioned, they will have the following labels populated with information from the request:

 * `pcf-organization-guid`
 * `pcf-space-guid`
 * `pcf-instance-id`

GCP labels have a more restricted character set than the Cloud Service Broker so unsupported characters will be mapped to the underscore character (`_`).

### GCP

On GCP, you can use these labels with the [BigQuery Billing Export](https://cloud.google.com/billing/docs/how-to/bq-examples)
to create reports about which organizations and spaces are incurring cost in your GCP project.
