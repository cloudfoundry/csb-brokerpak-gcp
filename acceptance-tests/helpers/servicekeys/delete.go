// Package servicekeys manages service keys
package servicekeys

import "csbbrokerpakgcp/acceptance-tests/helpers/cf"

func (s *ServiceKey) Delete() {
	cf.Run("delete-service-key", "-f", s.serviceInstanceName, s.name)
}
