package integration_tests

import (
	testframework "github.com/cloudfoundry/cloud-service-broker/brokerpaktestframework"
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

	Context("prevent updating properties of the service instance", func() {
		var instanceGUID string
		var err error

		BeforeEach(func() {
			instanceGUID, err = broker.Provision("csb-google-postgres", postgresNoOverridesPlan["name"].(string), map[string]interface{}{"cores": 1})

			invocations, err := mockTerraform.ApplyInvocations()
			Expect(err).NotTo(HaveOccurred())
			Expect(invocations).To(HaveLen(1))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("cores", float64(1)))
			mockTerraform.Reset()
		})

		DescribeTable(
			"should prevent users from updating",
			func(key string, value interface{}) {
				err = broker.Update(instanceGUID, "csb-google-postgres", postgresNoOverridesPlan["name"].(string), map[string]interface{}{key: value})

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(ContainSubstring("attempt to update parameter that may result in service instance re-creation and data loss")))

				Expect(mockTerraform.ApplyInvocations()).To(HaveLen(0))
			},
			Entry("cores", "cores", 5),
			Entry("postgres_version", "postgres_version", "POSTGRES_12"),
			Entry("storage_gb", "storage_gb", 11),
			Entry("credentials", "credentials", "creds"),
			Entry("instance_name", "instance_name", "name"),
			Entry("project", "project", "project_name"),
			Entry("db_name", "db_name", "new_db_name"),
			Entry("region", "region", "asia-northeast1"),
			Entry("authorized_network", "authorized_network", "new_network"),
			Entry("authorized_network", "authorized_network", "new_network_id"),
		)
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

	Context("bind a service ", func() {
		It("return the bind values from terraform output", func() {
			mockTerraform.ReturnTFState([]testframework.TFStateValue{
				{"hostname", "string", "create.hostname.gcp.test"},
				{"use_tls", "bool", false},
				{"username", "string", "create.test.username"},
				{"password", "string", "create.test.password"},
				{"port", "int", 9999},
				{"name", "string", "create.test.instancename"},
			})

			instanceID, err := broker.Provision("csb-google-postgres", postgresAllOverridesPlan["name"].(string), nil)
			Expect(err).NotTo(HaveOccurred())

			mockTerraform.ReturnTFState([]testframework.TFStateValue{
				{"username", "string", "bind.test.username"},
				{"password", "string", "bind.test.password"},
				{"uri", "string", "bind.test.uri"},
				{"jdbcUrl", "string", "bind.test.jdbcUrl"},
			})
			bindResult, err := broker.Bind("csb-google-postgres", postgresAllOverridesPlan["name"].(string), instanceID, nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(bindResult).To(Equal(map[string]interface{}{
				"username": "bind.test.username",
				"hostname": "create.hostname.gcp.test",
				"jdbcUrl":  "bind.test.jdbcUrl",
				"name":     "create.test.instancename",
				"password": "bind.test.password",
				"port":     float64(9999),
				"uri":      "bind.test.uri",
				"use_tls":  false,
			}))
		})
	})
})
