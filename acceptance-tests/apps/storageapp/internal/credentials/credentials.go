package credentials

import (
	"fmt"

	"github.com/mitchellh/mapstructure"

	"github.com/cloudfoundry-community/go-cfenv"
	//"cloud.google.com/go/storage"
	//"github.com/mitchellh/mapstructure"
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

	var r StorageCredentials
	if err := mapstructure.Decode(svs[0].Credentials, &r); err != nil {
		return StorageCredentials{}, fmt.Errorf("failed to decode credentials: %w", err)
	}

	if r.Credentials == "" || r.BucketName == "" {
		return StorageCredentials{}, fmt.Errorf("parsed credentials are not valid")
	}

	return r, nil
}
