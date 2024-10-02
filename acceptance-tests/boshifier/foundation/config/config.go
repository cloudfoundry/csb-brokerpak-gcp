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
		GoogleProject     string `env:"GOOGLE_PROJECT,required"`
		GoogleCredentials string `env:"GOOGLE_CREDENTIALS,required"`
		GCPPasNetwork     string `env:"GCP_PAS_NETWORK,required"`
	}
	Toolsmiths struct {
		EnvLockMetadata string `env:"ENVIRONMENT_LOCK_METADATA,required"`
	}
	Bosh struct {
		BoshEnvName      string `env:"BOSH_ENV_NAME,required"`
		BoshClient       string `env:"BOSH_CLIENT,required"`
		BoshEnvironment  string `env:"BOSH_ENVIRONMENT,required"`
		BoshClientSecret string `env:"BOSH_CLIENT_SECRET,required"`
		BoshCaCert       string `env:"BOSH_CA_CERT,required"`
		BoshAllProxy     string `env:"BOSH_ALL_PROXY,required"`
		BoshDeployment   string `env:"BOSH_DEPLOYMENT,required"`
	}
	Credhub struct {
		CredhubServer string `env:"CREDHUB_SERVER,required"`
		CredhubProxy  string `env:"CREDHUB_PROXY,required"`
		CredhubClient string `env:"CREDHUB_CLIENT,required"`
		CredhubSecret string `env:"CREDHUB_SECRET,required"`
		CredhubCACert string `env:"CREDHUB_CA_CERT,required"`
	}
	Brokerpak struct {
		GSBProvisionDefaults                  string `env:"GSB_PROVISION_DEFAULTS,required"`
		GSBServiceCsbGooglePostgresPlans      string `env:"GSB_SERVICE_CSB_GOOGLE_POSTGRES_PLANS,required"`
		GSBServiceCsbGoogleMysqlPlans         string `env:"GSB_SERVICE_CSB_GOOGLE_MYSQL_PLANS,required"`
		GSBServiceCsbGoogleStorageBucketPlans string `env:"GSB_SERVICE_CSB_GOOGLE_STORAGE_BUCKET_PLANS,required"`
		GSBBrokerpakConfig                    string `env:"GSB_BROKERPAK_CONFIG"`
	}
	UAA struct {
		CHUAAClientName   string `env:"CH_UAA_CLIENT_NAME,required"`
		CHUAAClientSecret string `env:"CH_UAA_CLIENT_SECRET,required"`
		CHUAAURL          string `env:"CH_UAA_URL,required"`
		CHCredHubURL      string `env:"CH_CRED_HUB_URL,required"`
	}
	CF struct {
		DeploymentGUID string `env:"CF_DEPLOYMENT_GUID,required"`
		APIPass        string `env:"CF_API_PASS,required"`
		APIDomain      string `env:"CF_API_DOMAIN,required"`
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

func Check() error {
	if _, err := Parse(); err != nil {
		return fmt.Errorf("failed reading env variables in cfg: %v", err)
	}

	return nil
}
