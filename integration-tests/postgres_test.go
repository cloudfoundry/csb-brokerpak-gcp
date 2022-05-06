package integration_test

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
	"tier":                  "db-f1-micro",
	"postgres_version":      "POSTGRES_14",
	"storage_gb":            float64(20),
	"credentials":           "plan_cred",
	"project":               "plan_project",
	"db_name":               "plan_db_name",
	"region":                "europe-west3",
	"authorized_network":    "plan_authorized_network",
	"authorized_network_id": "plan_authorized_network_id",
	"require_ssl":           false,
}

var postgresPlans = []map[string]interface{}{
	postgresNoOverridesPlan,
	postgresAllOverridesPlan,
}

var _ = Describe("postgres", func() {
	AfterEach(func() {
		Expect(mockTerraform.Reset()).NotTo(HaveOccurred())
	})

	It("publishes postgres in the catalog", func() {
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

	Context("updating properties of the service instance", func() {
		var instanceGUID string

		BeforeEach(func() {
			var err error
			instanceGUID, _ = broker.Provision("csb-google-postgres", postgresNoOverridesPlan["name"].(string), map[string]interface{}{"tier": "db-f1-micro"})

			invocations, err := mockTerraform.ApplyInvocations()
			Expect(err).NotTo(HaveOccurred())
			Expect(invocations).To(HaveLen(1))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("tier", "db-f1-micro"))
			mockTerraform.Reset()
		})

		DescribeTable(
			"should prevent users from updating",
			func(key string, value interface{}) {
				err := broker.Update(instanceGUID, "csb-google-postgres", postgresNoOverridesPlan["name"].(string), map[string]interface{}{key: value})

				Expect(err).To(MatchError(ContainSubstring("attempt to update parameter that may result in service instance re-creation and data loss")))

				Expect(mockTerraform.ApplyInvocations()).To(HaveLen(0))
			},
			Entry("postgres_version", "postgres_version", "POSTGRES_12"),
			Entry("instance_name", "instance_name", "name"),
			Entry("project", "project", "project_name"),
			Entry("db_name", "db_name", "new_db_name"),
			Entry("region", "region", "asia-northeast1"),
		)

		DescribeTable(
			"some allowed update",
			func(key string, value interface{}) {
				err := broker.Update(instanceGUID, "csb-google-postgres", postgresNoOverridesPlan["name"].(string), map[string]interface{}{key: value})
				Expect(err).NotTo(HaveOccurred())
			},
			Entry("tier", "tier", "db-g1-small"),
			Entry("storage_gb", "storage_gb", 11),
			Entry("authorized_network", "authorized_network", "new_network"),
			Entry("authorized_network_id", "authorized_network_id", "new_network_id"),
			Entry("authorized_networks_cidrs", "authorized_networks_cidrs", []string{"new cidr"}),
			Entry("public_ip", "public_ip", true),
			Entry("credentials", "credentials", "creds"),
		)
	})

	Context("versions of postgres", func() {
		It("defaults to postgres postgresql_13", func() {
			broker.Provision("csb-google-postgres", postgresNoOverridesPlan["name"].(string), map[string]interface{}{"tier": "db-f1-micro"})

			invocations, err := mockTerraform.ApplyInvocations()
			Expect(err).NotTo(HaveOccurred())
			Expect(invocations).To(HaveLen(1))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("database_version", "POSTGRES_13"))
		})

		DescribeTable(
			"supports custom postgres versions",
			func(version interface{}) {
				broker.Provision("csb-google-postgres", postgresNoOverridesPlan["name"].(string), map[string]interface{}{"tier": "db-f1-micro", "postgres_version": version})

				invocations, err := mockTerraform.ApplyInvocations()
				Expect(err).NotTo(HaveOccurred())
				Expect(invocations).To(HaveLen(1))
				Expect(invocations[0].TFVars()).To(HaveKeyWithValue("database_version", version))
			},
			Entry("11", "POSTGRES_11"),
			Entry("12", "POSTGRES_12"),
			Entry("13", "POSTGRES_13"),
			Entry("14", "POSTGRES_14"),
		)

		DescribeTable(
			"does not allow versions other than 11-14",
			func(version interface{}) {
				_, err := broker.Provision("csb-google-postgres", postgresNoOverridesPlan["name"].(string), map[string]interface{}{"tier": "db-f1-micro", "postgres_version": version})
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(ContainSubstring("postgres_version: postgres_version must be one of the following")))

				Expect(mockTerraform.ApplyInvocations()).To(HaveLen(0))
			},
			Entry("10", "POSTGRES_10"),
			Entry("15", "POSTGRES_15"),
			Entry("16", "POSTGRES_16"),
		)
	})

	Context("no properties overridden from the plan", func() {
		It("provision instance with defaults", func() {
			broker.Provision("csb-google-postgres", postgresNoOverridesPlan["name"].(string), map[string]interface{}{"tier": "db-f1-micro"})

			invocations, err := mockTerraform.ApplyInvocations()
			Expect(err).NotTo(HaveOccurred())
			Expect(invocations).To(HaveLen(1))

			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("db_name", "csb-db"))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("database_version", "POSTGRES_13"))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("tier", "db-f1-micro"))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("storage_gb", float64(10)))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("credentials", BrokerGCPCreds))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("project", BrokerGCPProject))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("authorized_network", "default"))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("authorized_network_id", ""))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("public_ip", false))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("authorized_networks_cidrs", make([]interface{}, 0)))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("require_ssl", true))
		})

		It("provisions instance with user parameters", func() {
			parameters := map[string]interface{}{
				"tier":                      "db-f1-micro",
				"postgres_version":          "POSTGRES_14",
				"storage_gb":                float64(20),
				"credentials":               "params_cred",
				"project":                   "params_project",
				"db_name":                   "params_db_name",
				"region":                    "europe-west3",
				"authorized_network":        "params_authorized_network",
				"authorized_network_id":     "params_authorized_network_id",
				"public_ip":                 true,
				"authorized_networks_cidrs": []string{"params_authorized_network_cidr1", "params_authorized_network_cidr2"},
				"require_ssl":               false,
			}
			broker.Provision("csb-google-postgres", postgresNoOverridesPlan["name"].(string), parameters)

			invocations, err := mockTerraform.ApplyInvocations()
			Expect(err).NotTo(HaveOccurred())
			Expect(invocations).To(HaveLen(1))

			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("db_name", parameters["db_name"]))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("database_version", parameters["postgres_version"]))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("tier", parameters["tier"]))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("storage_gb", float64(20)))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("credentials", parameters["credentials"]))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("project", parameters["project"]))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("authorized_network", parameters["authorized_network"]))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("authorized_network_id", parameters["authorized_network_id"]))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("public_ip", true))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("require_ssl", false))
			tfVars, _ := invocations[0].TFVars()
			Expect(tfVars["authorized_networks_cidrs"]).To(ConsistOf("params_authorized_network_cidr1", "params_authorized_network_cidr2"))
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
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("tier", postgresAllOverridesPlan["tier"]))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("storage_gb", postgresAllOverridesPlan["storage_gb"]))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("credentials", postgresAllOverridesPlan["credentials"]))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("project", postgresAllOverridesPlan["project"]))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("authorized_network", postgresAllOverridesPlan["authorized_network"]))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("authorized_network_id", postgresAllOverridesPlan["authorized_network_id"]))
			Expect(invocations[0].TFVars()).To(HaveKeyWithValue("require_ssl", postgresAllOverridesPlan["require_ssl"]))
		})
	})

	Context("bind a service ", func() {
		It("return the bind values from terraform output", func() {
			const (
				fakeSSLRoot    = "REAL_SSL_ROOT_CERT"
				fakeClientCert = "REAL_SSL_CLIENT_CERT"
				fakeClientKey  = "REAL_SSL_CLIENT_KEY"
			)
			mockTerraform.ReturnTFState([]testframework.TFStateValue{
				{Name: "hostname", Type: "string", Value: "create.hostname.gcp.test"},
				{Name: "username", Type: "string", Value: "create.test.username"},
				{Name: "password", Type: "string", Value: "create.test.password"},
				{Name: "name", Type: "string", Value: "create.test.instancename"},
				{Name: "require_ssl", Type: "bool", Value: false},
				{Name: "sslrootcert", Type: "string", Value: fakeSSLRoot},
				{Name: "sslcert", Type: "string", Value: fakeClientCert},
				{Name: "sslkey", Type: "string", Value: fakeClientKey},
			})

			instanceID, err := broker.Provision("csb-google-postgres", postgresAllOverridesPlan["name"].(string), nil)
			Expect(err).NotTo(HaveOccurred())

			mockTerraform.ReturnTFState([]testframework.TFStateValue{
				{Name: "username", Type: "string", Value: "bind.test.username"},
				{Name: "password", Type: "string", Value: "bind.test.password"},
				{Name: "uri", Type: "string", Value: "bind.test.uri"},
				{Name: "jdbcUrl", Type: "string", Value: "bind.test.jdbcUrl"},
			})
			bindResult, err := broker.Bind("csb-google-postgres", postgresAllOverridesPlan["name"].(string), instanceID, nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(bindResult).To(Equal(map[string]interface{}{
				"username":    "bind.test.username",
				"hostname":    "create.hostname.gcp.test",
				"jdbcUrl":     "bind.test.jdbcUrl",
				"name":        "create.test.instancename",
				"password":    "bind.test.password",
				"uri":         "bind.test.uri",
				"require_ssl": false,
				"sslrootcert": fakeSSLRoot,
				"sslcert":     fakeClientCert,
				"sslkey":      fakeClientKey,
			}))
		})
	})

	Context("property validation", func() {
		Describe("region", func() {
			It("should validate the region", func() {
				_, err := broker.Provision("csb-google-postgres", postgresNoOverridesPlan["name"].(string), map[string]interface{}{"region": "non-existent-region"})

				Expect(err).To(MatchError(ContainSubstring("region must be one of the following:")))
			})
		})
	})

	Describe("backup", func() {
		It("enables backup by default", func() {
			broker.Provision("csb-google-postgres", postgresNoOverridesPlan["name"].(string), map[string]interface{}{"tier": "db-f1-micro"})

			invocations, err := mockTerraform.ApplyInvocations()
			Expect(err).NotTo(HaveOccurred())
			Expect(invocations).To(HaveLen(1))
			Expect(invocations[0].TFVars()).To(SatisfyAll(
				HaveKeyWithValue("backups_retain_number", float64(7)),
				HaveKeyWithValue("backups_location", "us"),
				HaveKeyWithValue("backups_start_time", "07:00"),
				HaveKeyWithValue("backups_point_in_time_log_retain_days", float64(7)),
			))
		})

		It("allows backup to be configured", func() {
			broker.Provision("csb-google-postgres", postgresNoOverridesPlan["name"].(string), map[string]interface{}{
				"tier":                                  "db-f1-micro",
				"backups_retain_number":                 0,
				"backups_location":                      "eu",
				"backups_start_time":                    "09:15",
				"backups_point_in_time_log_retain_days": 0,
			})

			invocations, err := mockTerraform.ApplyInvocations()
			Expect(err).NotTo(HaveOccurred())
			Expect(invocations).To(HaveLen(1))
			Expect(invocations[0].TFVars()).To(SatisfyAll(
				HaveKeyWithValue("backups_retain_number", float64(0)),
				HaveKeyWithValue("backups_location", "eu"),
				HaveKeyWithValue("backups_start_time", "09:15"),
				HaveKeyWithValue("backups_point_in_time_log_retain_days", float64(0)),
			))
		})

		DescribeTable(
			"validation of backup properties",
			func(prop string, value interface{}, substring string) {
				_, err := broker.Provision("csb-google-postgres", postgresNoOverridesPlan["name"].(string), map[string]interface{}{"tier": "db-f1-micro", prop: value})
				Expect(err).To(MatchError(ContainSubstring(substring)))
			},
			Entry("min backups_retain_number", "backups_retain_number", -1, "backups_retain_number: Must be greater than or equal to 0"),
			Entry("max backups_retain_number", "backups_retain_number", 1001, "backups_retain_number: Must be less than or equal to 1000"),
			Entry("invalid backups_location", "backups_location", "moon", `backups_location must be one of the following:`),
			Entry("invalid backups_start_time", "backups_start_time", "34:91", `backups_start_time: Does not match pattern`),
			Entry("min backups_point_in_time_log_retain_days", "backups_point_in_time_log_retain_days", -1, "backups_point_in_time_log_retain_days: Must be greater than or equal to 0"),
			Entry("max backups_point_in_time_log_retain_days", "backups_point_in_time_log_retain_days", 8, "backups_point_in_time_log_retain_days: Must be less than or equal to 7"),
		)
	})
})
