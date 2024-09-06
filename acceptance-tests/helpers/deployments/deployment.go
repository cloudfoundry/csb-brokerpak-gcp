package deployments

import (
	"time"
)

type Deployment struct {
	Name           string
	DeploymentName string
	disk           string
	manifest       string
	start          bool
}

const deployWaitTime = 20 * time.Minute

type Option func(*Deployment)

func (i *Deployment) Deploy(opts ...Option) {

}
