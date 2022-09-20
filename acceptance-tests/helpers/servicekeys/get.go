package servicekeys

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"

	"csbbrokerpakgcp/acceptance-tests/helpers/cf"

	. "github.com/onsi/gomega"
)

func (s *ServiceKey) Get(receiver any) {
	Expect(reflect.ValueOf(receiver).Kind()).To(Equal(reflect.Ptr), "receiver must be a pointer")
	keyGUID, errStr := cf.Run("service-key", s.serviceInstanceName, s.name, "--guid")
	Expect(errStr).To(BeEmpty())
	nonGUID := regexp.MustCompile(`[^a-zA-Z0-9-]`)
	keyGUID = nonGUID.ReplaceAllString(keyGUID, "")
	keyURL := fmt.Sprintf("/v3/service_credential_bindings/%s/details", keyGUID)
	keyJSON, _ := cf.Run("curl", keyURL)
	if errStr != "" {
		deprecatedKeyURL := fmt.Sprintf("/v2/service_keys/%s", keyGUID)
		keyJSON, errStr = cf.Run("curl", deprecatedKeyURL)
		Expect(errStr).To(BeEmpty(), "unable to fetch the service key", s)
		keyEntity := struct {
			Entity map[string]any `json:"entity"`
		}{}
		Expect(json.Unmarshal([]byte(keyJSON), &keyEntity)).NotTo(HaveOccurred())
		keyJSONBytes, err := json.Marshal(keyEntity.Entity)
		Expect(err).NotTo(HaveOccurred())
		keyJSON = string(keyJSONBytes)
	}

	Expect(json.Unmarshal([]byte(keyJSON), receiver)).NotTo(HaveOccurred())
}
