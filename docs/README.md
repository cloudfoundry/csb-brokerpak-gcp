# Cloud Service Broker Documentation

## AI access families

This brokerpak now contains two distinct AI access families:

- `google_vertex_identity`: provisions Vertex AI resources and returns
  identity-backed runtime access as a service account key JSON during bind.
- `google_gemini_key`: provisions a Gemini API key restricted to the Generative
  Language API and returns that key directly during bind.

Keep these families separate in operator guidance and consumer integrations.

## Deploying Service Broker

### CloudFoundry

The service broker is deployed to CloudFoundry as a `cf push`ed application.

- [General configuration](./configuration.md)
- [GCP](./installation.md)

## Cloud Service Broker General

- Consuming Services are documented in Tanzu CSB for GCP official docs.

## Brokerpak Specifications

- [Brokerpak Intro](https://github.com/cloudfoundry/cloud-service-broker/tree/main/docs/brokerpak-intro.md)
- [Brokerpak Specification](https://github.com/cloudfoundry/cloud-service-broker/tree/main/docs/brokerpak-specification.md)

## Brokerpak Development

- [Brokerpak Dissection](https://github.com/cloudfoundry/cloud-service-broker/tree/main/docs/brokerpak-dissection.md)

For service-specific installation and configuration details, use the linked GCP docs above.

