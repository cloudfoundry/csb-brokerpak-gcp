package brokers

import (
	"csbbrokerpakgcp/acceptance-tests/helpers/apps"
	"fmt"
	"os"

	"github.com/onsi/ginkgo/v2"
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
		apps.EnvVar{Name: "GSB_SERVICE_CSB_GOOGLE_POSTGRES_PLANS", Value: `[{"name":"small","id":"5b45de36-cb90-11ec-a755-77f8be95a49d","description":"PostgreSQL with default version, shared CPU, minimum 0.6GB ram, 10GB storage","display_name":"small","tier":"db-f1-micro","storage_gb":10},{"name":"medium","id":"a3359fa6-cb90-11ec-bcb6-cb68544eda78","description":"PostgreSQL with default version, shared CPU, minimum 1.7GB ram, 20GB storage","display_name":"medium","tier":"db-g1-small","storage_gb":20},{"name":"large","id":"cd95c5b4-cb90-11ec-a5da-df87b7fb7426","description":"PostgreSQL with default version, minimum 8 cores, minimum 8GB ram, 50GB storage","display_name":"large","tier":"db-custom-8-8192","storage_gb":50}]`},
	)

	return append(result, b.envExtras...)
}

func (b Broker) releaseEnv() []apps.EnvVar {
	return []apps.EnvVar{
		{Name: "GSB_SERVICE_CSB_GOOGLE_POSTGRES_PLANS", Value: `[{"name":"small","id":"85b27a04-8695-11ea-818a-274131861b81","description":"PostgreSQL v11, shared CPU, minumum 0.6GB ram, 10GB storage","display_name":"small","cores":0.6,"postgres_version":"POSTGRES_11","storage_gb":10},{"name":"medium","id":"b41ee300-8695-11ea-87df-cfcb8aecf3bc","description":"PostgreSQL v11, shared CPU, minumum 1.7GB ram, 20GB storage","display_name":"medium","cores":1.7,"postgres_version":"POSTGRES_11","storage_gb":20},{"name":"large","id":"2a57527e-b025-11ea-b643-bf3bcf6d055a","description":"PostgreSQL v11, minumum 8 cores, minumum 8GB ram, 50GB storage","display_name":"large","cores":8,"postgres_version":"POSTGRES_11","storage_gb":50}]`},
	}
}

func (b Broker) latestEnv() []apps.EnvVar {
	return []apps.EnvVar{
		{Name: "GSB_SERVICE_CSB_GOOGLE_POSTGRES_PLANS", Value: `[{"name":"small","id":"5b45de36-cb90-11ec-a755-77f8be95a49d","description":"PostgreSQL with default version, shared CPU, minimum 0.6GB ram, 10GB storage","display_name":"small","tier":"db-f1-micro","storage_gb":10},{"name":"medium","id":"a3359fa6-cb90-11ec-bcb6-cb68544eda78","description":"PostgreSQL with default version, shared CPU, minimum 1.7GB ram, 20GB storage","display_name":"medium","tier":"db-g1-small","storage_gb":20},{"name":"large","id":"cd95c5b4-cb90-11ec-a5da-df87b7fb7426","description":"PostgreSQL with default version, minimum 8 cores, minimum 8GB ram, 50GB storage","display_name":"large","tier":"db-custom-8-8192","storage_gb":50}]`},
	}
}
