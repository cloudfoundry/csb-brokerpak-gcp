package brokers

import (
	"fmt"
	"os"

	"csbbrokerpakgcp/acceptance-tests/helpers/apps"

	"github.com/onsi/ginkgo/v2"
)

const (
	plansPostgreSQLVar = "GSB_SERVICE_CSB_GOOGLE_POSTGRES_PLANS"
	plansMySQLVar      = "GSB_SERVICE_CSB_GOOGLE_MYSQL_PLANS"
	plansStorageVar    = "GSB_SERVICE_CSB_GOOGLE_STORAGE_BUCKET_PLANS"
	plansPostgreSQL    = `[{"name":"small","id":"5b45de36-cb90-11ec-a755-77f8be95a49d","description":"PostgreSQL with default version, shared CPU, minimum 0.6GB ram, 10GB storage","metadata":{"displayName":"small"},"tier":"db-f1-micro","storage_gb":10},{"name":"medium","id":"a3359fa6-cb90-11ec-bcb6-cb68544eda78","description":"PostgreSQL with default version, shared CPU, minimum 1.7GB ram, 20GB storage","metadata":{"displayName":"medium"},"tier":"db-g1-small","storage_gb":20},{"name":"large","id":"cd95c5b4-cb90-11ec-a5da-df87b7fb7426","description":"PostgreSQL with default version, minimum 8 cores, minimum 8GB ram, 50GB storage","metadata":{"displayName":"large"},"tier":"db-custom-8-8192","storage_gb":50}]`
	plansMySQL         = `[{"name":"default","id":"eec62c9b-b25e-4e65-bad5-6b74d90274bf","description":"Default MySQL v8.0 10GB storage","metadata":{"displayName":"default"},"mysql_version":"MYSQL_8_0","storage_gb":10,"tier":"db-n1-standard-2"}]`
	plansStorage       = `[{"name": "default","id": "2875f0f0-a69f-4fe6-a5ec-5ed7f6e89a01","description": "Cloud Storage Bucket service with default configuration","metadata":{"display_name": "default"}}]`
	oldPlansStorage    = `[{"name": "private","id": "bbc4853e-8a63-11ea-a54e-670ca63cee0b","description": "Private Storage bucket", "region": "us-central1", "storage_class": "STANDARD"},{"name": "public-read","id": "c07f21a6-8a63-11ea-bc1b-d38b123189cb","description": "Public-read Storage bucket", "region": "us-central1", "storage_class": "STANDARD"}]`
)

func (b Broker) env() []apps.EnvVar {
	var result []apps.EnvVar

	for name, required := range map[string]bool{
		"GOOGLE_CREDENTIALS":                     true,
		"GOOGLE_PROJECT":                         true,
		"GCP_PAS_NETWORK":                        true,
		"GSB_BROKERPAK_BUILTIN_PATH":             false,
		"GSB_PROVISION_DEFAULTS":                 false,
		"CH_CRED_HUB_URL":                        false,
		"CH_UAA_URL":                             false,
		"CH_UAA_CLIENT_NAME":                     false,
		"CH_UAA_CLIENT_SECRET":                   false,
		"CH_SKIP_SSL_VALIDATION":                 false,
		"GSB_COMPATIBILITY_ENABLE_BETA_SERVICES": false,
	} {
		val, ok := os.LookupEnv(name)
		switch {
		case ok:
			result = append(result, apps.EnvVar{Name: name, Value: val})
		case !ok && required:
			ginkgo.Fail(fmt.Sprintf("You must set the %s environment variable", name))
		}
	}

	result = append(result,
		apps.EnvVar{Name: "SECURITY_USER_NAME", Value: b.username},
		apps.EnvVar{Name: "SECURITY_USER_PASSWORD", Value: b.password},
		apps.EnvVar{Name: "DB_TLS", Value: "skip-verify"},
		apps.EnvVar{Name: "ENCRYPTION_ENABLED", Value: true},
		apps.EnvVar{Name: "ENCRYPTION_PASSWORDS", Value: b.secrets},
		apps.EnvVar{Name: "BROKERPAK_UPDATES_ENABLED", Value: true},
		apps.EnvVar{Name: "TERRAFORM_UPGRADES_ENABLED", Value: true},
		apps.EnvVar{Name: plansPostgreSQLVar, Value: plansPostgreSQL},
		apps.EnvVar{Name: plansMySQLVar, Value: plansMySQL},
		apps.EnvVar{Name: plansStorageVar, Value: plansStorage},
	)

	return append(result, b.envExtras...)
}

func (b Broker) releasedEnv() []apps.EnvVar {
	return []apps.EnvVar{
		{Name: plansPostgreSQLVar, Value: plansPostgreSQL},
	}
}

func (b Broker) latestEnv() []apps.EnvVar {
	return []apps.EnvVar{
		{Name: plansPostgreSQLVar, Value: plansPostgreSQL},
		{Name: plansMySQLVar, Value: plansMySQL},
		{Name: plansStorageVar, Value: oldPlansStorage},
	}
}
