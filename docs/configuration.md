# Broker Configuration
The broker can be configured though environment variables or configuration files or a combo of both.

## Configuration File
A configuration file can be provided at run time to the broker.
```bash
cloud-service-broker serve --config <config file name>
```

A configuration file can be YAML or JSON. Config file values that are `.` delimited represent hierarchy in the config file.

Example:
```
db:
  host: hostname
```
represents a config file value of `db.host`

## Database Configuration Properties

Connection details for the backing database for the service broker.

You can configure the following values:

| Environment Variable | Config File Value | Type | Description |
|----------------------|------|-------------|------------------|
| <tt>DB_HOST</tt> <b>*</b> | db.host | string | <p>Database host </p>|
| <tt>DB_USERNAME</tt> | db.user | string | <p>Database username </p>|
| <tt>DB_PASSWORD</tt> | db.password | secret | <p>Database password </p>|
| <tt>DB_PORT</tt> <b>*</b> | db.port | string | <p>Database port (defaults to 3306)  Default: <code>3306</code></p>|
| <tt>DB_NAME</tt> <b>*</b> | db.name | string | <p>Database name  Default: <code>servicebroker</code></p>|
| <tt>CA_CERT</tt> | db.ca.cert | text | <p>Server CA cert </p>|
| <tt>CLIENT_CERT</tt> | db.client.cert | text | <p>Client cert </p>|
| <tt>CLIENT_KEY</tt> | db.client.key | text | <p>Client key </p>|

## Broker Service Configuration

Broker service configuration values:
| Environment Variable | Config File Value | Type | Description |
|----------------------|------|-------------|------------------|
| <tt>SECURITY_USER_NAME</tt> <b>*</b> | api.user | string | <p>Broker authentication username</p>|
| <tt>SECURITY_USER_PASSWORD</tt> <b>*</b> | api.password | string | <p>Broker authentication password</p>|
| <tt>PORT</tt> | api.port | string | <p>Port to bind broker to</p>|

## Credhub Configuration
The broker supports passing credentials to apps via [credhub references](https://github.com/cloudfoundry-incubator/credhub/blob/master/docs/secure-service-credentials.md#service-brokers), thus keeping them private to the application (they won't show up in `cf env app_name` output.)

| Environment Variable | Config File Value | Type | Description |
|----------------------|------|-------------|------------------|
| CH_CRED_HUB_URL           |credhub.url    | URL | credhub service URL - usually `https://credhub.service.cf.internal:8844`|
| CH_UAA_URL                |credhub.uaa_url | URL | uaa service URL - usually `https://uaa.service.cf.internal:8443`|
| CH_UAA_CLIENT_NAME        |credhub.uaa_client_name| string | uaa username - usually `credhub_admin_client`|
| CH_UAA_CLIENT_SECRET      |credhub.uaa_client_secret| string | uaa client secret - "*Credhub Admin Client Credentials*" from *Operations Manager > PAS > Credentials* tab. |
| CH_SKIP_SSL_VALIDATION    |credhub.skip_ssl_validation| boolean | skip SSL validation if true | 
| CH_CA_CERT_FILE           |credhub.ca_cert_file| path | path to cert file |


## Brokerpak Configuration

Brokerpak configuration values:
| Environment Variable | Config File Value | Type | Description |
|----------------------|------|-------------|------------------|
|<tt>GSB_BROKERPAK_BUILTIN_PATH</tt> | brokerpak.builtin.path | string | <p>Path to search for .brokerpak files, default: <code>./</code></p>|
|<tt>GSB_BROKERPAK_CONFIG</tt>|brokerpak.config| string | JSON global config for broker pak services|
|<tt>GSB_PROVISION_DEFAULTS</tt>|provision.defaults| string | JSON global provision defaults|
|<tt>GSB_SERVICE_*SERVICE_NAME*_PROVISION_DEFAULTS</tt>|service.*service-name*.provision.defaults| string | JSON provision defaults override for *service-name*|
|<tt>GSB_SERVICE_*SERVICE_NAME*_PLANS</tt>|service.*service-name*.plans| string | JSON plan collection to augment plans for *service-name*|
|<tt>GSB_COMPATIBILITY_ENABLE_BETA_SERVICES</tt>| compatibility.enable-beta-services | bool | Enable services tagged with `beta`. Default: `false` |


## Google Configuration

The GCP brokerpak supports default values for tenant, subscription and service principal credentials.

| Environment Variable | Config File Value | Type | Description |
|----------------------|-------------------|------|-------------|
| GOOGLE_CREDENTIALS   | gcp.credentials   | string | the string version of the credentials file created for the Owner level Service Account |
| GOOGLE_PROJECT       | gcp.project       | string | gcp project id |


### config file example
```
gcp:
  credentials: |
    <credentials json as string>
  project: your-project-id
db:
  host: your mysql host
  password: your mysql password
  user: your mysql username
api:
  user: someusername
  password: somepassword
credhub:
  url: ...
  uaa_url: ...
  uaa_client_name: ...
  uaa_client_secret: ...
 ```
 
### Global Config Example

Services for a given IaaS should have common parameter names for service wide platform resources (like location)

GCP services support global region and authorized_network parameters:

```yaml
provision:
  defaults: '{
    "region": "europe-west1", 
    "authorized_network": "pcf-env-network"
  }'
```


### Plans Example

Plans should be added to the brokerpak configuration:

```yaml
service:
  csb-csb-google-postgres:
    plans: '[
      {
        "name":"small",
        "id":"85b27a04-8695-11ea-818a-274131861b81",
        "description":"PostgreSQL with default version, shared CPU, minumum 0.6GB ram, 10GB storage",
        "display_name":"small",
        "cores":0.6,
        "storage_gb":10
      }
    ]'
```
