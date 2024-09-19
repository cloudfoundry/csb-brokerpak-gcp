package bosh

import (
	"boshifier/business/capi"
	"boshifier/foundation/config"
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/google/uuid"
)

func CreateVarsFile(cfg config.Config, cfAPI capi.Data, sk capi.ServiceKey, varsTemplateFile, varsFile string) error {
	varsTemplate, err := os.ReadFile(varsTemplateFile)
	if err != nil {
		return fmt.Errorf("failed to read vars template file: %v", err)
	}

	data := struct {
		config.Config
		CSBDBData   string
		CFAPIPass   string
		CFAPIDomain string
	}{
		Config:      cfg,
		CSBDBData:   createCSBDBManifestBlock(sk),
		CFAPIPass:   cfAPI.CFAPIPass,
		CFAPIDomain: cfAPI.CFAPIDomain,
	}

	tmpl, err := template.New("vars").Parse(string(varsTemplate))
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	if err := os.WriteFile(varsFile, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to create vars file: %v", err)
	}
	return nil
}

func createCSBDBManifestBlock(sk capi.ServiceKey) string {

	return fmt.Sprintf(
		`{
	"host": "%s",
	"encryption": { "enabled": true, "passwords": [{"password": {"secret": "%s"}, "label": "first-encryption", "primary": true}] },
	"ca": { "cert": "%s" },
	"name": "service_instance_db",
	"user": "%s",
	"password": "%s",
	"port": %d
}`,
		sk.Credentials.Hostname,
		uuid.NewString(),
		strings.ReplaceAll(sk.Credentials.TLS.Cert.CA, "\n", "\\n"),
		sk.Credentials.Username,
		sk.Credentials.Password,
		sk.Credentials.Port,
	)
}
