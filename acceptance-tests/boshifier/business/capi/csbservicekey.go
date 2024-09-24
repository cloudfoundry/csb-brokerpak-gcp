package capi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type ServiceKey struct {
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
		Name     string `json:"name"`
	} `json:"credentials"`
}

func CreateCSBServiceKey(service, key string, params map[string]string) (ServiceKey, error) {
	if err := createServiceKey(service, key, params); err != nil {
		return ServiceKey{}, err
	}
	return extractServiceKeyData(service, key)
}

func createServiceKey(service, key string, params map[string]string) error {
	cmdArgs := []string{"create-service-key", service, key}
	if len(params) > 0 {
		paramsJSON, err := json.Marshal(params)
		if err != nil {
			return fmt.Errorf("failed to marshal service key params to JSON: %v", err)
		}
		cmdArgs = append(cmdArgs, "-c", string(paramsJSON))
	}

	// CAPI cf create-service-key returns a zero exit code if the service key already exists
	cmd := exec.Command("cf", cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create service key: %v", err)
	}
	return nil
}

func extractServiceKeyData(service, key string) (ServiceKey, error) {
	cfCmd := exec.Command("cf", "service-key", service, key)
	tailCmd := exec.Command("tail", "-n+2")
	// Create a pipe to connect the output of cfCmd to the input of tailCmd
	pipe, err := cfCmd.StdoutPipe()
	if err != nil {
		return ServiceKey{}, fmt.Errorf("failed to create pipe: %v", err)
	}
	tailCmd.Stdin = pipe
	var output bytes.Buffer
	tailCmd.Stdout = &output

	if err := cfCmd.Start(); err != nil {
		return ServiceKey{}, fmt.Errorf("failed to start cf command: %v", err)
	}

	if err := tailCmd.Start(); err != nil {
		return ServiceKey{}, fmt.Errorf("failed to start tail command: %v", err)
	}

	if err := cfCmd.Wait(); err != nil {
		return ServiceKey{}, fmt.Errorf("cf command failed: %v", err)
	}

	if err := tailCmd.Wait(); err != nil {
		return ServiceKey{}, fmt.Errorf("tail command failed: %v", err)
	}

	rawServiceKey := strings.TrimSpace(output.String())

	var serviceKey ServiceKey
	if err := json.Unmarshal([]byte(rawServiceKey), &serviceKey); err != nil {
		return ServiceKey{}, fmt.Errorf("failed to parse service-key JSON data: %v", err)
	}

	return serviceKey, nil
}
