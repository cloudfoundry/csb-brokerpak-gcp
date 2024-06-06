package credentials

import (
	"encoding/base64"
	"fmt"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/mitchellh/mapstructure"
)

type PubSubCredentials struct {
	Credentials      string `mapstructure:"credentials"`
	TopicName        string `mapstructure:"topic_name"`
	SubscriptionName string `mapstructure:"subscription_name"`
	ProjectID        string `mapstructure:"ProjectId"`
	PrivateKeyData   string `mapstructure:"PrivateKeyData"`
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

	if r.PrivateKeyData == "" || r.TopicName == "" || r.ProjectID == "" {
		return r, fmt.Errorf("parsed credentials are not valid")
	}

	cred, err := base64.StdEncoding.DecodeString(r.PrivateKeyData)
	if err != nil {
		return r, fmt.Errorf("error decoding private key in JSON format, base64 encoded: %s", err.Error())
	}
	r.Credentials = string(cred)

	return r, nil
}
