package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/caarlos0/env/v11"
	"github.com/google/uuid"
)

var (
	varsTemplateFile, varsFile, manifest, releaseRepoPath string
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
		GSBServiceCsbGoogleRedisPlans         string `env:"GSB_SERVICE_CSB_GOOGLE_REDIS_PLANS"`
		GSBBrokerpakConfig                    string `env:"GSB_BROKERPAK_CONFIG"`
	}
	UAA struct {
		CHUAAClientName   string `env:"CH_UAA_CLIENT_NAME"`
		CHUAAClientSecret string `env:"CH_UAA_CLIENT_SECRET"`
		CHUAAURL          string `env:"CH_UAA_URL"`
		CHCredHubURL      string `env:"CH_CRED_HUB_URL"`
	}
}

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("failed to get home directory: %v", err)
	}

	varsTemplateFile = filepath.Join(homeDir, "workspace/csb/csb-brokerpak-gcp/acceptance-tests/assets/vars-template.yml")
	varsFile = filepath.Join(homeDir, "workspace/csb/csb-brokerpak-gcp/acceptance-tests/assets/vars.yml")
	manifest = filepath.Join(homeDir, "workspace/csb/csb-brokerpak-gcp/acceptance-tests/assets/manifest.yml")
	releaseRepoPath = filepath.Join(homeDir, "workspace/csb/csb-brokerpak-gcp/../csb-gcp-release/")
}

func main() {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to create cfg: %v", err)
	}
	cfg.Brokerpak.GSBBrokerpakConfig = gsbBrokerpakConfig

	org := "pivotal"
	space := "broker-cf-test"
	dbPassword := uuid.NewString()
	opsManagerEnvVars(cfg.Toolsmiths.EnvLockMetadata)
	cfLogin(cfg.Toolsmiths.EnvLockMetadata, org, space)
	createServiceKey("csb-sql", "csb-sql")
	csbDBData := extractServiceKeyData("csb-sql", "csb-sql", dbPassword)
	capiData := cfAPIData()
	createVarsFile(cfg, csbDBData, capiData)
	createTMPBoshManifest()
}

func opsManagerEnvVars(filePath string) {
	cmd := exec.Command("smith", "om", "-l", filePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("failed to source environment metadata: %v", err)
	}
}

func cfLogin(filePath, org, space string) {
	cmd := exec.Command("smith", "-l", filePath, "cf-login")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("failed to login to Cloud Foundry: %v", err)
	}

	cmd = exec.Command("cf", "target", "-o", org, "-s", space)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("failed to target org: %v", err)
	}
}

func createServiceKey(service, key string) {
	cmd := exec.Command("cf", "service-key", service, key)
	if err := cmd.Run(); err != nil {
		log.Printf("service key does not exist, creating it %v", err)
		cmd = exec.Command("cf", "create-service-key", service, key)
		if err := cmd.Run(); err != nil {
			log.Fatalf("failed to create service key: %v", err)
		}
	}
}

func extractServiceKeyData(service, key, dbPassword string) string {
	cfCmd := exec.Command("cf", "service-key", service, key)
	tailCmd := exec.Command("tail", "-n+2")
	// Create a pipe to connect the output of cfCmd to the input of tailCmd
	pipe, err := cfCmd.StdoutPipe()
	if err != nil {
		log.Fatalf("failed to create pipe: %v", err)
	}
	tailCmd.Stdin = pipe
	var output bytes.Buffer
	tailCmd.Stdout = &output

	if err := cfCmd.Start(); err != nil {
		log.Fatalf("failed to start cf command: %v", err)
	}

	if err := tailCmd.Start(); err != nil {
		log.Fatalf("failed to start tail command: %v", err)
	}

	if err := cfCmd.Wait(); err != nil {
		log.Fatalf("cf command failed: %v", err)
	}

	if err := tailCmd.Wait(); err != nil {
		log.Fatalf("tail command failed: %v", err)
	}

	csbDbDataRaw := strings.TrimSpace(output.String())

	var data struct {
		Credentials struct {
			Hostname string `json:"hostname"`
			TLS      struct {
				Cert struct {
					CA string `json:"ca"`
				} `json:"cert"`
			} `json:"tls"`
			Username string `json:"username"`
			Password string `json:"password"`
			Port     int    `json:"port"`
		} `json:"credentials"`
	}

	if err := json.Unmarshal([]byte(csbDbDataRaw), &data); err != nil {
		log.Fatalf("failed to parse service-key JSON data: %v", err)
	}

	csbDbData := fmt.Sprintf(
		`{
	"host": "%s",
	"encryption": { "enabled": true, "passwords": [{"password": {"secret": "%s"}, "label": "first-encryption", "primary": true}] },
	"ca": { "cert": "%s" },
	"name": "service_instance_db",
	"user": "%s",
	"password": "%s",
	"port": %d
}`,
		data.Credentials.Hostname,
		dbPassword,
		strings.ReplaceAll(data.Credentials.TLS.Cert.CA, "\n", "\\n"),
		data.Credentials.Username,
		data.Credentials.Password,
		data.Credentials.Port,
	)
	return csbDbData
}

func cloudFoundryDeploymentGUID() string {
	cmd := exec.Command("om", "-k", "curl", "-s", "-p", "/api/v0/staged/products")
	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("failed to get product information: %v", err)
	}
	var products []map[string]interface{}
	if err := json.Unmarshal(output, &products); err != nil {
		log.Fatalf("failed to parse product information: %v", err)
	}
	for _, product := range products {
		if product["type"] == "cf" {
			return product["guid"].(string)
		}
	}

	log.Fatalf("failed to find product information: %v", err)
	return ""
}

type CFAPI struct {
	CFAPIPass   string
	CFAPIDomain string
}

func cfAPIPass(cfDeploymentGUID string) (string, error) {
	cmd := exec.Command(
		"credhub",
		"get",
		"--key",
		"password",
		"-n",
		fmt.Sprintf("/opsmgr/%s/uaa/admin_credentials", cfDeploymentGUID),
		"-j",
	)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get CF_API_PASS: %v", err)
	}

	return strings.TrimSpace(string(output)), nil
}

func cfAPIDomain() (string, error) {
	cmd := exec.Command("cf", "api")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get CF_API_DOMAIN: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 1 {
		return "", fmt.Errorf("unexpected output from cf api command")
	}

	parts := strings.Split(lines[0], "/")
	if len(parts) < 3 {
		return "", fmt.Errorf("unexpected output from cf api command")
	}

	return parts[2], nil
}

func cfAPIData() CFAPI {
	pass, err := cfAPIPass(cloudFoundryDeploymentGUID())
	if err != nil {
		log.Fatalf("failed to get CF API Pass: %v", err)
	}

	domain, err := cfAPIDomain()
	if err != nil {
		log.Fatalf("failed to get CF API Domain: %v", err)
	}

	return CFAPI{
		CFAPIPass:   pass,
		CFAPIDomain: domain,
	}
}

func createVarsFile(cfg Config, csbDbData string, cfAPI CFAPI) {
	varsTemplate, err := os.ReadFile(varsTemplateFile)
	if err != nil {
		log.Fatalf("failed to read vars template file: %v", err)
	}

	data := struct {
		Config
		CsbDbData   string
		CFAPIPass   string
		CFAPIDomain string
	}{
		Config:      cfg,
		CsbDbData:   csbDbData,
		CFAPIPass:   cfAPI.CFAPIPass,
		CFAPIDomain: cfAPI.CFAPIDomain,
	}

	tmpl, err := template.New("vars").Parse(string(varsTemplate))
	if err != nil {
		log.Fatalf("failed to parse template: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		log.Fatalf("failed to execute template: %v", err)
	}

	if err := os.WriteFile(varsFile, buf.Bytes(), 0644); err != nil {
		log.Fatalf("failed to create vars file: %v", err)
	}
}

func createTMPBoshManifest() {
	cmd := exec.Command(
		"bosh",
		"int",
		manifest,
		"-l",
		varsFile,
		"-v",
		fmt.Sprintf("release_repo_path=%s", releaseRepoPath),
	)
	tmpManifestFile, err := os.Create("/tmp/tmp-manifest.yml")
	if err != nil {
		log.Fatalf("failed to create temporary manifest file: %v", err)
	}
	defer tmpManifestFile.Close()

	cmd.Stdout = tmpManifestFile
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("failed to create temporary BOSH manifest: %v", err)
	}
}
