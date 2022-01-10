package credentials

import (
	"fmt"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/go-redis/redis/v8"
	"github.com/mitchellh/mapstructure"
)

func Read() (*redis.Options, error) {
	app, err := cfenv.Current()
	if err != nil {
		return nil, fmt.Errorf("error reading app env: %w", err)
	}
	svs, err := app.Services.WithTag("redis")
	if err != nil {
		return nil, fmt.Errorf("error reading Redis service details")
	}

	var r struct {
		Host string `mapstructure:"host"`
		Port int    `mapstructure:"port"`
	}

	if err := mapstructure.Decode(svs[0].Credentials, &r); err != nil {
		return nil, fmt.Errorf("failed to decode credentials: %w", err)
	}

	if r.Host == "" || r.Port == 0 {
		return nil, fmt.Errorf("parsed credentials are not valid")
	}

	return &redis.Options{
		Addr: fmt.Sprintf("%s:%d", r.Host, r.Port),
		DB:   0,
	}, nil
}
