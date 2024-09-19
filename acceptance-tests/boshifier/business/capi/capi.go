package capi

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type Data struct {
	CFAPIPass               string
	CFAPIDomain             string
	envLockMetadataFilePath string
	org                     string
	space                   string
}

func New(envLockMetadataFilePath, org, space string) (Data, error) {
	c := Data{
		envLockMetadataFilePath: envLockMetadataFilePath,
		org:                     org,
		space:                   space,
	}
	if err := Login(c.envLockMetadataFilePath, c.org, c.space); err != nil {
		return Data{}, err
	}
	return c, nil
}

func (c Data) Data() (Data, error) {
	deploymentGUID, err := c.cloudFoundryDeploymentGUID()
	if err != nil {
		return Data{}, fmt.Errorf("failed to get CF API Data: %v", err)
	}
	pass, err := c.cfAPIPass(deploymentGUID)
	if err != nil {
		return Data{}, fmt.Errorf("failed to get CF API Pass: %v", err)
	}

	domain, err := c.cfAPIDomain()
	if err != nil {
		return Data{}, fmt.Errorf("failed to get CF API Domain: %v", err)
	}

	return Data{
		CFAPIPass:   pass,
		CFAPIDomain: domain,
	}, nil
}

func (c Data) cfAPIPass(cfDeploymentGUID string) (string, error) {
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

func (c Data) cfAPIDomain() (string, error) {
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

func (c Data) cloudFoundryDeploymentGUID() (string, error) {
	cmd := exec.Command("om", "-k", "curl", "-s", "-p", "/api/v0/staged/products")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get product information: %v", err)
	}
	var products []map[string]interface{}
	if err := json.Unmarshal(output, &products); err != nil {
		return "", fmt.Errorf("failed to parse product information: %v", err)
	}
	for _, product := range products {
		if product["type"] == "cf" {
			return product["guid"].(string), nil
		}
	}

	return "", fmt.Errorf("failed to find product information")
}
