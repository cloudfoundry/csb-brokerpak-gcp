package credentials

import (
	"fmt"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/mitchellh/mapstructure"
)

type SpannerCredentials struct {
	Credentials  string `mapstructure:"credentials"`
	DBName       string `mapstructure:"db_name"`
	InstanceName string `mapstructure:"instance"`
	ProjectId    string `mapstructure:"ProjectId"`
	FullDBName   string
}

func Read() (SpannerCredentials, error) {
	app, err := cfenv.Current()
	if err != nil {
		return SpannerCredentials{}, fmt.Errorf("error reading app env: %w", err)
	}

	svs, err := app.Services.WithTag("spanner")
	if err != nil {
		return SpannerCredentials{}, fmt.Errorf("error reading Spanner service details")
	}

	var r SpannerCredentials
	if err := mapstructure.Decode(svs[0].Credentials, &r); err != nil {
		return SpannerCredentials{}, fmt.Errorf("failed to decode credentials: %w", err)
	}

	if r.Credentials == "" || r.DBName == "" || r.InstanceName == "" || r.ProjectId == "" {
		return SpannerCredentials{}, fmt.Errorf("parsed credentials are not valid: %v", r)
	}

	r.FullDBName = fmt.Sprintf("projects/%s/instances/%s/databases/%s", r.ProjectId, r.InstanceName, r.DBName)

	return r, nil
}
