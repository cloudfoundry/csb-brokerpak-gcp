# Installing the broker on GCP

The broker service and the GCP brokerpak can be pushed and registered on a foundation running on GCP.

Documentation for broker configuration can be found [here](./configuration.md).

## Requirements

### CloudFoundry running on GCP.
The GCP brokerpak services are provisioned with firewall rules that only allow internal connectivity. 
This allows `cf push`ed applications access, while denying any public access.

### GCP Service Credentials

#### [Set up a GCP Project](#project)

1. Go to the [Google Cloud Console](https://console.cloud.google.com) and sign up, walking through the setup wizard.
1. A page then displays with a collection of options. Select "Create Project" option.
1. Give your project a name and click "Create".
1. The dashboard for the newly created project will be displayed.

#### [Enable APIs](#apis)

Enable the following services in **[APIs and services > Library](https://console.cloud.google.com/apis/library)**.

1. Enable the [Google Cloud Resource Manager API](https://console.cloud.google.com/apis/api/cloudresourcemanager.googleapis.com/overview)
1. Enable the [Google Identity and Access Management (IAM) API](https://console.cloud.google.com/apis/api/iam.googleapis.com/overview)
1. If you want to enable CloudSQL as a service (MySQL and PostgreSQL), enable the [CloudSQL API](https://console.cloud.google.com/apis/library/sql-component.googleapis.com), [CloudSQL Admin API](https://console.developers.google.com/apis/api/sqladmin.googleapis.com/overview), and [Service Networking API](https://console.cloud.google.com/apis/library/servicenetworking.googleapis.com)
1. If you want to enable BigQuery as a service, enable the [BigQuery API](https://console.cloud.google.com/apis/api/bigquery/overview)
1. If you want to enable Cloud Storage as a service, enable the [Cloud Storage API](https://console.cloud.google.com/apis/api/storage_component/overview)
1. If you want to enable Pub/Sub as a service, enable the [Cloud Pub/Sub API](https://console.cloud.google.com/apis/library/pubsub.googleapis.com)
1. If you want to enable Bigtable as a service, enable the [Bigtable Admin API](https://console.cloud.google.com/apis/library/bigtable.googleapis.com)
1. If you want to enable Datastore as a service, enable the [Datastore API](https://console.cloud.google.com/apis/api/datastore.googleapis.com/overview)
1. If you want to enable Redis as a service, enable the [Redis API](https://console.cloud.google.com/apis/library/redis.googleapis.com)
1. If you want to enable Dataproc as a service, enable the [Dataproc API](https://console.developers.google.com/apis/api/dataproc.googleapis.com/overview)
1. If you want to enable Cloud Spanner as a service, enable the [Cloud Spanner API](https://console.developers.google.com/apis/api/spanner.googleapis.com/overview)

#### [Create a root service account](#service-account)

1. From the GCP console, navigate to **IAM & Admin > Service accounts** and click **Create Service Account**.
1. Enter a **Service account name**.
1. In the **Project Role** dropdown, choose **Project > Owner**.
1. Select the checkbox to **Furnish a new Private Key**, make sure the **JSON** key type is specified.
1. Click **Save** to create the account, key and grant it the owner permission.
1. Save the automatically downloaded key file to a secure location.

### MySQL Database for Broker State
The broker keeps service instance and binding information in a MySQL database. 

#### Binding a MySQL Database
If there is an existing broker in the foundation that can provision a MySQL instance use `cf create-service` to create a new MySQL instance. Then use `cf bind-service` to bind that instance to the service broker.

#### Scripted
Use [scripts/gcp-create-mysql-db.sh](../scripts/gcp-create-mysql-db.sh) to create a GCP mysql instance. It will report the DB_HOST (ip address) username, password and db name upon completion.

It requires the [gcloud](https://cloud.google.com/sdk/gcloud) cli be installed.
#### Manually Provisioning a MySQL Database

The GCP Service Broker stores the state of provisioned resources in a MySQL database.
You may use any database compatible with the MySQL protocol.
We recommend a second generation GCP CloudSQL instance with automatic backups, high availability and automatic maintenance.
The service broker does not require much disk space, but we do recommend an SSD for faster interactions with the broker.

1. Create new MySQL instance.
1. **CloudSQL Only** Make sure that the database can be accessed, add `0.0.0.0/0` as an authorized network.
1. Run `CREATE DATABASE servicebroker;`
1. Run `CREATE USER '<username>'@'%' IDENTIFIED BY '<password>';`
1. Run `GRANT ALL PRIVILEGES ON servicebroker.* TO '<username>'@'%' WITH GRANT OPTION;`
1. **CloudSQL Only** (Optional) create SSL certs for the database and save them somewhere secure.

The following configuration parameters will be needed:
- `DB_HOST`
- `DB_USERNAME`
- `DB_PASSWORD`


#### [Set required environment variables](#required-env)

Add these to the `env` section of `manifest.yml`

* `GOOGLE_CREDENTIALS` - the string version of the credentials file created for the Owner level Service Account.
* `SECURITY_USER_NAME` - the username to authenticate broker requests - the same one used in `cf create-service-broker`.
* `SECURITY_USER_PASSWORD` - the password to authenticate broker requests - the same one used in `cf create-service-broker`.
* `DB_HOST` - the host for the database to back the service broker.
* `DB_USERNAME` - the database username for the service broker to use.
* `DB_PASSWORD` - the database password for the service broker to use.

### Create Private Service Connection in GCP

To allow CF applications to connect to service instances created by CSB, follow [these instructions](https://cloud.google.com/vpc/docs/configure-private-services-access) to enable private service access to the VPC network that your foundation is running in.

To peer the service network (that mysql and postgres instances are connected to) and your VPC, the following commands need to be run once. Note that the `prefix-length` value depends on how many databases are created. If you run out of available IP addresses then consider using a lower number.

```bash
VPC_NETWORK_NAME=[the name of your VCP network]
PROJECT=[your GCP project id]
gcloud compute addresses create google-managed-services-mysql-${VPC_NETWORK_NAME} \
    --global \
    --purpose=VPC_PEERING \
    --prefix-length=23 \
    --network=${VPC_NETWORK_NAME} \
    --project=${PROJECT}

gcloud services vpc-peerings connect \
    --service=servicenetworking.googleapis.com \
    --ranges=google-managed-services-mysql-${VPC_NETWORK_NAME} \
    --network=${VPC_NETWORK_NAME} \
    --project=${PROJECT}
```
> if you use *scripts/gcp-create-mysql-db.sh* to create the mysql metadata database for the broker, these steps are already done.

### Authorized Network ID
When using private service connections, the ID for the VPC network must provided in the `authorized_network_id` parameter when creating service instances. To get the ID of the given network, use 

```
gcloud compute networks list --filter="name=$GCP_PAS_NETWORK" --uri
```
where GCP_PAS_NETWORK is the name of the network used when creating the private service network peering above.

### Fetch A Broker and GCP Brokerpak

Download a release from https://github.com/pivotal/cloud-service-broker/releases. Find the latest release matching the name pattern `sb-0.1.0-rc.XXX-gcp-0.0.1-rc.YY`. This will have a broker and brokerpak that have been tested together. Follow the hyperlink into that release and download `cloud-servic-broker` and `gcp-services-0.1.0-rc.YY.brokerpak` into the same directory on your workstation.

### Create a MySQL instance with GCP broker
The following command will create a basic MySQL database instance named `csb-sql`
```bash
cf create-service google-cloudsql-mysql basic csb-sql
```

### Build Config File
To avoid putting any sensitive information in environment variables, a config file can be used.

Create a file named `config.yml` in the same directory the broker and brokerpak have been downloaded to. Its contents should be:

```yaml
gcp:
  google_credentials: the string version of the credentials file created for the Owner level Service Account
  google_project: Give your project a name 
```

### Push and Register the Broker

Push the broker as a binary application:

```bash
make push-broker
```

Once this completes, the output from `cf marketplace` should include:

```
csb-google-mysql            small, medium, large   Mysql is a fully managed service for the Google Cloud Platform.

csb-google-postgres         small, medium, large   PostgreSQL is a fully managed service for the Google Cloud Platform.

csb-google-redis            basic, ha              Cloud Memorystore for Redis is a fully managed Redis service for the Google Cloud Platform. 

csb-google-storage-bucket   private, public-read   Google Cloud Storage that uses the Terraform back-end and grants service accounts IAM permissions directly on the bucket.      

csb-google-bigquery         standard               A fast, economical and fully managed data warehouse for large-scale data analytics.   

csb-google-dataproc         standard, ha           Dataproc is a fully-managed service for running Apache Spark and Apache Hadoop clusters in a simpler, more cost-efficient way.   

csb-google-spanner          small, medium, large   Fully managed, scalable, relational database service for regional and global application data.  
```


## Step By Step From a Pre-built Release with a Manually Provisioned MySQL Instance

Fetch a pre-built broker and brokerpak and configure with a manually provisioned MySQL instance.

Requirements and assumptions are the same as above. Follow instructions above to [fetch the broker and brokerpak](#Fetch-A-Broker-and-GCP-Brokerpak)

### Create a MySQL Database
Its an exercise for the reader to create a MySQL server somewhere that a `cf push`ed app can access. The database connection values (hostname, user name and password) will be needed in the next step. It is also necessary to create a database named `servicebroker` within that server (use your favorite tool to connect to the MySQL server and issue `CREATE DATABASE servicebroker;`).

### Build Config File
To avoid putting any sensitive information in environment variables, a config file can be used.

Create a file named `config.yml` in the same directory the broker and brokerpak have been downloaded to. Its contents should be:

```yaml
gcp:
  google_credentials: the string version of the credentials file created for the Owner level Service Account
  google_project: Give your project a name

db:
  host: your mysql host
  password: your mysql password
  user: your mysql username

api:
  user: someusername
  password: somepassword
```

### Push and Register the Broker

Push the broker as a binary application and register it as a broker:

```bash
make push-broker
```

Once these steps are complete, the output from `cf marketplace` should resemble the same as above.

## Step By Step From Source with Bound MySQL
Grab the source code, build and deploy.

### Requirements

The following tools are needed on your workstation:
- [go 1.18](https://golang.org/dl/)
- make
- [cf cli](https://docs.cloudfoundry.org/cf-cli/install-go-cli.html)

The Cloud Service Broker for GCP must be installed in your foundation.

### Assumptions

The `cf` CLI has been used to authenticate with a foundation (`cf api` and `cf login`,) and an org and space have been targeted (`cf target`)

### Clone the Repo

The following commands will clone the service broker repository and cd into the resulting directory.
```bash
git clone https://github.com/cloudfoundry/cloud-service-broker.git
cd cloud-service-broker
```
### Set Required Environment Variables

Collect the GCP Service Account credentials for your account and set them as environment variables:
```bash
export ROOT_SERVICE_ACCOUNT_JSON=the string version of the credentials file created for the Owner level Service Account

```
Generate username and password for the broker - Cloud Foundry will use these credentials to authenticate API calls to the service broker.
```bash
export SECURITY_USER_NAME=someusername
export SECURITY_USER_PASSWORD=somepassword
```

### Create a MySQL instance

The following command will create a basic MySQL database instance named `csb-sql`
```bash
cf create-service google-cloudsql-mysql basic csb-sql
```
### Use the Makefile to Deploy the Broker
There is a make target that will build the broker and brokerpak and deploy to and register with Cloud Foundry as a space scoped broker. This will be local and private to the org and space your `cf` CLI is targeting.

```bash
make push-broker
```

Once these steps are complete, the output from `cf marketplace` should resemble the same as above.

## Step By Step Slightly Harder Way

Requirements and assumptions are the same as above. Follow instructions for the first two steps above ([Clone the Repo](#Clone-the-Repo) and [Set Required Environment Variables](Set-Required-Environment-Variables))

### Create a MySQL Database
Its an exercise for the reader to create a MySQL server somewhere that a `cf push`ed app can access. It is also necessary to create a database named `servicebroker` within that server (use your favorite tool to connect to the MySQL server and issue `CREATE DATABASE servicebroker;`). Set the following environment variables with information about that MySQL instance:
```bash
export DB_HOST=mysql server host
export DB_USERNAME=mysql server username
export DB_PASSWORD=mysql server password
```

### Build the Broker and Brokerpak
Use the makefile to build the broker executable and brokerpak.
```bash
make build-gcp-brokerpak
```
## Uninstalling the Broker
First, make sure there are all service instances created with `cf create-service` have been destroyed with `cf delete-service` otherwise removing the broker will fail.

### Unregister the Broker
```bash
cf delete-service-broker csb-$USER
```

### Uninstall the Broker
```bash
cf delete cloud-service-broker
```
