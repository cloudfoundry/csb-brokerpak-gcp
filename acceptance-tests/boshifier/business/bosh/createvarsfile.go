package bosh

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"text/template"

	"boshifier/foundation/config"
)

func CreateVarsFile(cfg config.Config, dbBlock DBBlock, varsTemplateFile, varsFile string) error {
	varsTemplate, err := os.ReadFile(varsTemplateFile)
	if err != nil {
		return fmt.Errorf("failed to read vars template file: %v", err)
	}

	dbData, err := json.Marshal(dbBlock)
	if err != nil {
		return fmt.Errorf("failed to marshal DB block: %v", err)
	}

	data := struct {
		config.Config
		DBData string
	}{
		Config: cfg,
		DBData: string(dbData),
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
