package brokers

import (
	"csbbrokerpakgcp/acceptance-tests/helpers/cf"
	"encoding/json"
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// bindDatabase will bind the CSB app to an appropriate state database
// The aim is for this to be transparent to anyone running the tests
// to reduce the friction of running tests, and therefore hopefully
// have folks run tests more often.
func bindDatabase(app string) {
	GinkgoHelper()

	switch {
	case ourSQLPresent():
		name := fmt.Sprintf("oursql-%s", app)
		cf.Run("create-service", "oursql", "standard", name)
		cf.Run("bind-service", app, name)
	case pmysqlPresent():
		name := strings.ReplaceAll(app, "-", "_")
		cf.Run("bind-service", app, "csb-sql", "-c", fmt.Sprintf(`{"schema":"%s"}`, name))
	default:
		Fail("can't work out which database to bind to")
	}
}

func deleteDatabase(app string) {
	name := fmt.Sprintf("oursql-%s", app)
	cf.Run("delete-service", "-f", name) // idempotent, so ok if it doesn't exist
}

func ourSQLPresent() bool {
	out, _ := cf.Run("curl", "/v3/service_offerings?names=oursql")

	var receiver struct {
		Resources []struct{} `json:"resources"`
	}
	Expect(json.Unmarshal([]byte(out), &receiver)).To(Succeed())
	return len(receiver.Resources) > 0
}

func pmysqlPresent() bool {
	out, _ := cf.Run("curl", "/v3/service_instances?names=csb-sql")

	var receiver struct {
		Resources []struct{} `json:"resources"`
	}
	Expect(json.Unmarshal([]byte(out), &receiver)).To(Succeed())
	return len(receiver.Resources) > 0
}
