package credentials

import (
	b64 "encoding/base64"
	"fmt"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/mitchellh/mapstructure"
)

type DataprocCredentials struct {
	PrivateKey  string `mapstructure:"private_key"`
	BucketName  string `mapstructure:"bucket_name"`
	ClusterName string `mapstructure:"cluster_name"`
	ProjectID   string `mapstructure:"project_id"`
	Region      string `mapstructure:"region"`
	Credentials []byte
}

func Read() (DataprocCredentials, error) {
	app, err := cfenv.Current()
	if err != nil {
		return DataprocCredentials{}, fmt.Errorf("error reading app env: %w", err)
	}
	svs, err := app.Services.WithTag("dataproc")
	if err != nil {
		return DataprocCredentials{}, fmt.Errorf("error reading Dataproc service details")
	}

	var r DataprocCredentials
	if err := mapstructure.Decode(svs[0].Credentials, &r); err != nil {
		return DataprocCredentials{}, fmt.Errorf("failed to decode binding credentials: %w", err)
	}

	if r.PrivateKey == "" || r.BucketName == "" || r.ClusterName == "" || r.ProjectID == "" || r.Region == "" {
		return DataprocCredentials{}, fmt.Errorf("parsed credentials are not valid: %v", r)
	}

	credential, err := b64.StdEncoding.DecodeString(r.PrivateKey)
	if err != nil {
		return DataprocCredentials{}, fmt.Errorf("failed to decode PrivateKey: %w", err)
	}

	r.Credentials = credential
	return r, nil
}
