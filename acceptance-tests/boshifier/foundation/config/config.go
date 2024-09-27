package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

const (
	gsbBrokerpakConfig = `{"global_labels":[{"key":"key1","value":"value1"},{"key":"key2","value":"value2"}]}`
)

type Config struct {
	Google struct {
		GoogleProject     string `env:"GOOGLE_PROJECT"`
		GoogleCredentials string `env:"GOOGLE_CREDENTIALS"`
		GCPPasNetwork     string `env:"GCP_PAS_NETWORK"`
	}
	Toolsmiths struct {
		EnvLockMetadata string `env:"ENVIRONMENT_LOCK_METADATA"`
	}
	Bosh struct {
		BoshEnvName      string `env:"BOSH_ENV_NAME"`
		BoshClient       string `env:"BOSH_CLIENT"`
		BoshEnvironment  string `env:"BOSH_ENVIRONMENT"`
		BoshClientSecret string `env:"BOSH_CLIENT_SECRET"`
		BoshCaCert       string `env:"BOSH_CA_CERT"`
		BoshAllProxy     string `env:"BOSH_ALL_PROXY"`
		BoshDeployment   string `env:"BOSH_DEPLOYMENT"`
	}
	Credhub struct {
		CredhubServer string `env:"CREDHUB_SERVER"`
		CredhubProxy  string `env:"CREDHUB_PROXY"`
		CredhubClient string `env:"CREDHUB_CLIENT"`
		CredhubSecret string `env:"CREDHUB_SECRET"`
		CredhubCACert string `env:"CREDHUB_CA_CERT"`
	}
	Brokerpak struct {
		GSBProvisionDefaults                  string `env:"GSB_PROVISION_DEFAULTS"`
		GSBServiceCsbGooglePostgresPlans      string `env:"GSB_SERVICE_CSB_GOOGLE_POSTGRES_PLANS"`
		GSBServiceCsbGoogleMysqlPlans         string `env:"GSB_SERVICE_CSB_GOOGLE_MYSQL_PLANS"`
		GSBServiceCsbGoogleStorageBucketPlans string `env:"GSB_SERVICE_CSB_GOOGLE_STORAGE_BUCKET_PLANS"`
		GSBBrokerpakConfig                    string `env:"GSB_BROKERPAK_CONFIG"`
	}
	UAA struct {
		CHUAAClientName   string `env:"CH_UAA_CLIENT_NAME"`
		CHUAAClientSecret string `env:"CH_UAA_CLIENT_SECRET"`
		CHUAAURL          string `env:"CH_UAA_URL"`
		CHCredHubURL      string `env:"CH_CRED_HUB_URL"`
	}
}

func Parse() (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, fmt.Errorf("failed reading env variables in cfg: %v", err)
	}
	cfg.Brokerpak.GSBBrokerpakConfig = gsbBrokerpakConfig

	return cfg, nil
}
