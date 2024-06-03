package credentials

import (
	"fmt"
	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/mitchellh/mapstructure"
)

type PubSubCredentials struct {
	Credentials      string `mapstructure:"credentials"`
	TopicName        string `mapstructure:"topic_name"`
	SubscriptionName string `mapstructure:"subscription_name"`
	ProjectID        string `mapstructure:"ProjectId"`
}

func Read() (PubSubCredentials, error) {
	app, err := cfenv.Current()
	if err != nil {
		return PubSubCredentials{}, fmt.Errorf("error reading app env: %w", err)
	}
	svs, err := app.Services.WithTag("pubsub")
	if err != nil {
		return PubSubCredentials{}, fmt.Errorf("error reading PubSub service details")
	}

	return readBinding(svs[0].Credentials)
}

func readBinding(creds any) (PubSubCredentials, error) {
	var r PubSubCredentials
	if err := mapstructure.Decode(creds, &r); err != nil {
		return r, fmt.Errorf("failed to decode credentials: %w", err)
	}

	if r.Credentials == "" || r.TopicName == "" || r.ProjectID == "" {
		return r, fmt.Errorf("parsed credentials are not valid")
	}

	return r, nil
}
