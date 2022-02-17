package integration_tests

import (
	testframework "github.com/cloudfoundry-incubator/cloud-service-broker/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var postgresNoOverridesPlan = map[string]interface{}{
	"name":         "no-overrides",
	"id":           "5f60d632-8f1e-11ec-9832-7bd519d660a9",
	"description":  "no-override-description",
	"display_name": "no-overrides-plan-display-name",
}

var postgresAllOverridesPlan = map[string]interface{}{
	"name":                  "all-overrides",
	"id":                    "4be43944-8f20-11ec-9ea5-834eb2499c32",
	"description":           "all-override-description",
	"display_name":          "all-overrides-plan-display-name",
	"cores":                 float64(10),
	"postgres_version":      "POSTGRES_14",
	"storage_gb":            float64(20),
	"credentials":           "plan_cred",
	"project":               "plan_project",
	"db_name":               "plan_db_name",
	"region":                "europe-west3",
	"authorized_network":    "plan_authorized_network",
	"authorized_network_id": "plan_authorized_network_id",
}

var postgresPlans = []map[string]interface{}{
	postgresNoOverridesPlan,
	postgresAllOverridesPlan,
}

var _ = Describe("postgres", func() {
	AfterEach(func() {
		Expect(mockTerraform.Reset()).NotTo(HaveOccurred())
	})

	It("publish postgres in the catalog", func() {
		catalog, err := broker.Catalog()
		Expect(err).NotTo(HaveOccurred())
		service := testframework.FindService(catalog, "csb-google-postgres")
		Expect(service.Plans).To(HaveLen(2))
		Expect(service.Metadata.ImageUrl).NotTo(BeEmpty())
		Expect(service.Metadata.DocumentationUrl).NotTo(BeEmpty())
		Expect(service.Metadata.SupportUrl).NotTo(BeEmpty())

		planMetadata := testframework.FindServicePlan(catalog, "csb-google-postgres", postgresNoOverridesPlan["name"].(string))
		Expect(planMetadata.Description).NotTo(BeEmpty())

	})
	Context("no properties overridden from the plan", func() {
		It("provision instance with defaults", func() {
			broker.Provision("csb-google-postgres", postgresNoOverridesPlan["name"].(string), map[string]interface{}{"cores": 1})

			invocations, err := mockTerraform.ApplyInvocations()
			Expect(err).NotTo(HaveOccurred())
			Expect(invocations).To(HaveLen(1))

			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("db_name", "csb-db"))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("database_version", "POSTGRES_11"))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("cores", float64(1)))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("storage_gb", float64(10)))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("credentials", BrokerGCPCreds))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("project", BrokerGCPProject))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("authorized_network", "default"))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("authorized_network_id", ""))
		})

		It("provision instance with user parameters", func() {
			parameters := map[string]interface{}{
				"cores":                 float64(10),
				"postgres_version":      "POSTGRES_14",
				"storage_gb":            float64(20),
				"credentials":           "params_cred",
				"project":               "params_project",
				"db_name":               "params_db_name",
				"region":                "europe-west3",
				"authorized_network":    "params_authorized_network",
				"authorized_network_id": "params_authorized_network_id",
			}
			broker.Provision("csb-google-postgres", postgresNoOverridesPlan["name"].(string), parameters)

			invocations, err := mockTerraform.ApplyInvocations()
			Expect(err).NotTo(HaveOccurred())
			Expect(invocations).To(HaveLen(1))

			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("db_name", parameters["db_name"]))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("database_version", parameters["postgres_version"]))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("cores", float64(10)))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("storage_gb", float64(20)))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("credentials", parameters["credentials"]))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("project", parameters["project"]))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("authorized_network", parameters["authorized_network"]))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("authorized_network_id", parameters["authorized_network_id"]))
		})
	})

	Context("properties have been overridden from the plan", func() {
		It("should use properties from the plan", func() {
			broker.Provision("csb-google-postgres", postgresAllOverridesPlan["name"].(string), nil)

			invocations, err := mockTerraform.ApplyInvocations()
			Expect(err).NotTo(HaveOccurred())
			Expect(invocations).To(HaveLen(1))

			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("db_name", postgresAllOverridesPlan["db_name"]))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("database_version", postgresAllOverridesPlan["postgres_version"]))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("cores", postgresAllOverridesPlan["cores"]))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("storage_gb", postgresAllOverridesPlan["storage_gb"]))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("credentials", postgresAllOverridesPlan["credentials"]))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("project", postgresAllOverridesPlan["project"]))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("authorized_network", postgresAllOverridesPlan["authorized_network"]))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("authorized_network_id", postgresAllOverridesPlan["authorized_network_id"]))
		})
	})
})
