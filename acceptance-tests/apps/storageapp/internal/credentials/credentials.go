package credentials

import (
	"encoding/base64"
	"fmt"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/mitchellh/mapstructure"
)

type StorageCredentials struct {
	Credentials string `mapstructure:"credentials"`
	BucketName  string `mapstructure:"bucket_name"`
}

func Read() (StorageCredentials, error) {
	app, err := cfenv.Current()
	if err != nil {
		return StorageCredentials{}, fmt.Errorf("error reading app env: %w", err)
	}
	svs, err := app.Services.WithTag("storage")
	if err != nil {
		return StorageCredentials{}, fmt.Errorf("error reading Storage service details")
	}

	switch svs[0].Label {
	case "google-storage":
		return readLegacyBrokerBinding(svs[0].Credentials)
	default:
		return readBinding(svs[0].Credentials)
	}
}

func readBinding(creds any) (StorageCredentials, error) {
	var r StorageCredentials
	if err := mapstructure.Decode(creds, &r); err != nil {
		return r, fmt.Errorf("failed to decode credentials: %w", err)
	}

	if r.Credentials == "" || r.BucketName == "" {
		return r, fmt.Errorf("parsed credentials are not valid")
	}

	return r, nil
}

func readLegacyBrokerBinding(creds any) (StorageCredentials, error) {
	type legacyBindingData struct {
		BucketName     string `mapstructure:"bucket_name"`
		PrivateKeyData string `mapstructure:"PrivateKeyData"`
	}
	var (
		r StorageCredentials
		l legacyBindingData
	)
	if err := mapstructure.Decode(creds, &l); err != nil {
		return r, err
	}

	v, err := base64.StdEncoding.DecodeString(l.PrivateKeyData)
	if err != nil {
		return r, fmt.Errorf("error decoding private key in JSON format, base64 encoded: %s", err.Error())
	}

	r.BucketName = l.BucketName
	r.Credentials = string(v)
	return r, nil
}
