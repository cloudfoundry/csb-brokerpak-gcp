package integration_test

import (
	testframework "github.com/cloudfoundry/cloud-service-broker/v2/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"golang.org/x/exp/maps"
)

const (
	postgresServiceName        = "csb-google-postgres"
	postgresServiceID          = "40501b82-cb90-11ec-b1c2-e3a703778055"
	postgresServiceDescription = "PostgreSQL is a fully managed service for the Google Cloud Platform."
	postgresServiceDisplayName = "Google Cloud PostgreSQL"
	postgresServiceSupportURL  = "https://cloud.google.com/support/"
)

var postgresNoOverridesPlan = map[string]any{
	"name":        "no-overrides",
	"id":          "5f60d632-8f1e-11ec-9832-7bd519d660a9",
	"description": "no-override-description",
}

var postgresAllOverridesPlan = map[string]any{
	"name":                  "all-overrides",
	"id":                    "4be43944-8f20-11ec-9ea5-834eb2499c32",
	"description":           "all-override-description",
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

var postgresPlans = []map[string]any{
	postgresNoOverridesPlan,
	postgresAllOverridesPlan,
}

var _ = Describe("postgres", Label("postgres"), func() {
	BeforeEach(func() {
		Expect(mockTerraform.SetTFState([]testframework.TFStateValue{})).To(Succeed())
	})

	AfterEach(func() {
		Expect(mockTerraform.Reset()).To(Succeed())
	})

	It("publishes in the catalog", func() {
		catalog, err := broker.Catalog()
		Expect(err).NotTo(HaveOccurred())

		service := testframework.FindService(catalog, postgresServiceName)
		Expect(service.ID).To(Equal(postgresServiceID))
		Expect(service.Description).To(Equal(postgresServiceDescription))
		Expect(service.Tags).To(ConsistOf("gcp", "postgresql", "postgres"))
		Expect(service.Metadata.ImageUrl).To(ContainSubstring("data:image/png;base64,"))
		Expect(service.Metadata.DisplayName).To(Equal(postgresServiceDisplayName))
		Expect(service.Metadata.DocumentationUrl).To(Equal(cloudServiceBrokerDocumentationURL))
		Expect(service.Metadata.ProviderDisplayName).To(Equal(providerDisplayName))
		Expect(service.Metadata.SupportUrl).To(Equal(postgresServiceSupportURL))
		Expect(service.Plans).To(
			ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					ID:   Equal("4be43944-8f20-11ec-9ea5-834eb2499c32"),
					Name: Equal("all-overrides"),
				}),
				MatchFields(IgnoreExtras, Fields{
					ID:   Equal("5f60d632-8f1e-11ec-9832-7bd519d660a9"),
					Name: Equal("no-overrides"),
				}),
			),
		)
	})

	Context("updating properties of the service instance", func() {
		var instanceGUID string

		BeforeEach(func() {
			var err error
			instanceGUID, err = broker.Provision("csb-google-postgres", postgresNoOverridesPlan["name"].(string), map[string]any{"tier": "db-f1-micro"})

			Expect(err).NotTo(HaveOccurred())
			Expect(mockTerraform.FirstTerraformInvocationVars()).To(HaveKeyWithValue("tier", "db-f1-micro"))
			Expect(mockTerraform.Reset()).To(Succeed())
		})

		DescribeTable(
			"should prevent users from updating",
			func(key string, value any) {
				err := broker.Update(instanceGUID, "csb-google-postgres", postgresNoOverridesPlan["name"].(string), map[string]any{key: value})

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
			func(key string, value any) {
				err := broker.Update(instanceGUID, "csb-google-postgres", postgresNoOverridesPlan["name"].(string), map[string]any{key: value})

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
			_, err := broker.Provision("csb-google-postgres", postgresNoOverridesPlan["name"].(string), map[string]any{"tier": "db-f1-micro"})

			Expect(err).NotTo(HaveOccurred())
			Expect(mockTerraform.FirstTerraformInvocationVars()).To(HaveKeyWithValue("database_version", "POSTGRES_13"))
		})

		DescribeTable(
			"supports custom postgres versions",
			func(version any) {
				_, err := broker.Provision("csb-google-postgres", postgresNoOverridesPlan["name"].(string), map[string]any{"tier": "db-f1-micro", "postgres_version": version})

				Expect(err).NotTo(HaveOccurred())
				Expect(mockTerraform.FirstTerraformInvocationVars()).To(HaveKeyWithValue("database_version", version))
			},
			Entry("11", "POSTGRES_11"),
			Entry("12", "POSTGRES_12"),
			Entry("13", "POSTGRES_13"),
			Entry("14", "POSTGRES_14"),
			Entry("15", "POSTGRES_15"),
		)
	})

	Context("no properties overridden from the plan", func() {
		It("provision instance with defaults", func() {
			_, err := broker.Provision("csb-google-postgres", postgresNoOverridesPlan["name"].(string), map[string]any{"tier": "db-f1-micro"})

			Expect(err).NotTo(HaveOccurred())
			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("db_name", "csb-db"),
					HaveKeyWithValue("database_version", "POSTGRES_13"),
					HaveKeyWithValue("tier", "db-f1-micro"),
					HaveKeyWithValue("storage_gb", float64(10)),
					HaveKeyWithValue("credentials", brokerGCPCreds),
					HaveKeyWithValue("project", brokerGCPProject),
					HaveKeyWithValue("authorized_network", "default"),
					HaveKeyWithValue("authorized_network_id", ""),
					HaveKeyWithValue("public_ip", false),
					HaveKeyWithValue("authorized_networks_cidrs", make([]any, 0)),
					HaveKeyWithValue("require_ssl", true),
				),
			)
		})

		It("provisions instance with user parameters", func() {
			parameters := map[string]any{
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
			_, err := broker.Provision("csb-google-postgres", postgresNoOverridesPlan["name"].(string), parameters)

			Expect(err).NotTo(HaveOccurred())
			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("db_name", parameters["db_name"]),
					HaveKeyWithValue("database_version", parameters["postgres_version"]),
					HaveKeyWithValue("tier", parameters["tier"]),
					HaveKeyWithValue("storage_gb", float64(20)),
					HaveKeyWithValue("credentials", parameters["credentials"]),
					HaveKeyWithValue("project", parameters["project"]),
					HaveKeyWithValue("authorized_network", parameters["authorized_network"]),
					HaveKeyWithValue("authorized_network_id", parameters["authorized_network_id"]),
					HaveKeyWithValue("public_ip", true),
					HaveKeyWithValue("require_ssl", false),
					HaveKeyWithValue("authorized_networks_cidrs", ConsistOf(
						"params_authorized_network_cidr1",
						"params_authorized_network_cidr2",
					)),
					HaveKeyWithValue("highly_available", BeFalse()),
					HaveKeyWithValue("location_preference_zone", BeEmpty()),
					HaveKeyWithValue("location_preference_secondary_zone", BeEmpty()),
				),
			)
		})
	})

	Context("properties have been overridden from the plan", func() {
		It("should use properties from the plan", func() {
			_, err := broker.Provision("csb-google-postgres", postgresAllOverridesPlan["name"].(string), nil)

			Expect(err).NotTo(HaveOccurred())
			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("db_name", postgresAllOverridesPlan["db_name"]),
					HaveKeyWithValue("database_version", postgresAllOverridesPlan["postgres_version"]),
					HaveKeyWithValue("tier", postgresAllOverridesPlan["tier"]),
					HaveKeyWithValue("storage_gb", postgresAllOverridesPlan["storage_gb"]),
					HaveKeyWithValue("credentials", postgresAllOverridesPlan["credentials"]),
					HaveKeyWithValue("project", postgresAllOverridesPlan["project"]),
					HaveKeyWithValue("authorized_network", postgresAllOverridesPlan["authorized_network"]),
					HaveKeyWithValue("authorized_network_id", postgresAllOverridesPlan["authorized_network_id"]),
					HaveKeyWithValue("require_ssl", postgresAllOverridesPlan["require_ssl"]),
				),
			)
		})
	})

	Context("bind a service ", func() {
		It("return the bind values from terraform output", func() {
			const (
				fakeSSLRoot    = "REAL_SSL_ROOT_CERT"
				fakeClientCert = "REAL_SSL_CLIENT_CERT"
				fakeClientKey  = "REAL_SSL_CLIENT_KEY"
				fakePrivateIP  = "REAL_PRIVATE_IP"
			)
			err := mockTerraform.SetTFState([]testframework.TFStateValue{
				{Name: "hostname", Type: "string", Value: "create.hostname.gcp.test"},
				{Name: "username", Type: "string", Value: "create.test.username"},
				{Name: "password", Type: "string", Value: "create.test.password"},
				{Name: "name", Type: "string", Value: "create.test.instancename"},
				{Name: "require_ssl", Type: "bool", Value: false},
				{Name: "sslrootcert", Type: "string", Value: fakeSSLRoot},
				{Name: "sslcert", Type: "string", Value: fakeClientCert},
				{Name: "sslkey", Type: "string", Value: fakeClientKey},
				{Name: "private_ip", Type: "string", Value: fakePrivateIP},
			})
			Expect(err).NotTo(HaveOccurred())

			instanceID, err := broker.Provision("csb-google-postgres", postgresAllOverridesPlan["name"].(string), nil)
			Expect(err).NotTo(HaveOccurred())

			err = mockTerraform.SetTFState([]testframework.TFStateValue{
				{Name: "username", Type: "string", Value: "bind.test.username"},
				{Name: "password", Type: "string", Value: "bind.test.password"},
				{Name: "uri", Type: "string", Value: "bind.test.uri"},
				{Name: "jdbcUrl", Type: "string", Value: "bind.test.jdbcUrl"},
			})
			Expect(err).NotTo(HaveOccurred())
			bindResult, err := broker.Bind("csb-google-postgres", postgresAllOverridesPlan["name"].(string), instanceID, nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(bindResult).To(Equal(map[string]any{
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
				"private_ip":  fakePrivateIP,
			}))
		})
	})

	Context("property validation", func() {
		Describe("region", func() {
			It("should validate the region", func() {
				_, err := broker.Provision("csb-google-postgres", postgresNoOverridesPlan["name"].(string), map[string]any{"region": "-Asia-northeast1"})

				Expect(err).To(MatchError(ContainSubstring("region: Does not match pattern '^[a-z][a-z0-9-]+$'")))
			})
		})

		Describe("postgres_version", func() {
			It("should validate the postgres_version", func() {
				_, err := broker.Provision("csb-google-postgres", postgresNoOverridesPlan["name"].(string), map[string]any{"postgres_version": "15"})

				Expect(err).To(MatchError(ContainSubstring("postgres_version: Does not match pattern '^POSTGRES_[0-9]+$'")))
			})
		})

		DescribeTable("property constraints",
			func(params map[string]any, expectedErrorMsg string) {
				mandatoryParams := map[string]any{"tier": "db-f1-micro"}
				maps.Copy(mandatoryParams, params)
				_, err := broker.Provision("csb-google-postgres", postgresNoOverridesPlan["name"].(string), mandatoryParams)

				Expect(err).To(MatchError(ContainSubstring(expectedErrorMsg)))
			},
			Entry(
				"storage capacity minimum value is 10",
				map[string]any{"storage_gb": 9},
				"storage_gb: Must be greater than or equal to 10",
			),
			Entry(
				"invalid region",
				map[string]any{"region": "-Asia-northeast1"},
				"region: Does not match pattern '^[a-z][a-z0-9-]+$'",
			),
			Entry(
				"invalid backup location",
				map[string]any{"backups_location": "australia-central."},
				"backups_location: Does not match pattern '^[a-z][a-z0-9-]+$'",
			),
			Entry(
				"invalid backups retain number",
				map[string]any{"backups_retain_number": -7},
				"backups_retain_number: Must be greater than or equal to 0",
			),
			Entry(
				"invalid preferred primary zone",
				map[string]any{"location_preference_zone": "abc"},
				"location_preference_zone: Does not match pattern '^[a-z]?$'",
			),
			Entry(
				"invalid preferred secondary zone",
				map[string]any{"location_preference_secondary_zone": "abc"},
				"location_preference_secondary_zone: Does not match pattern '^[a-z]?$'",
			),
			Entry(
				"invalid postgres_version",
				map[string]any{"postgres_version": "15"},
				"postgres_version: Does not match pattern '^POSTGRES_[0-9]+$'",
			),
		)
	})

	Describe("backup", func() {
		It("enables backup by default", func() {
			_, err := broker.Provision("csb-google-postgres", postgresNoOverridesPlan["name"].(string), map[string]any{"tier": "db-f1-micro"})
			Expect(err).NotTo(HaveOccurred())

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
			_, err := broker.Provision("csb-google-postgres", postgresNoOverridesPlan["name"].(string), map[string]any{
				"tier":                                  "db-f1-micro",
				"backups_retain_number":                 0,
				"backups_location":                      "eu",
				"backups_start_time":                    "09:15",
				"backups_point_in_time_log_retain_days": 0,
			})

			Expect(err).NotTo(HaveOccurred())
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
			func(prop string, value any, substring string) {
				_, err := broker.Provision("csb-google-postgres", postgresNoOverridesPlan["name"].(string), map[string]any{"tier": "db-f1-micro", prop: value})
				Expect(err).To(MatchError(ContainSubstring(substring)))
			},
			Entry("min backups_retain_number", "backups_retain_number", -1, "backups_retain_number: Must be greater than or equal to 0"),
			Entry("max backups_retain_number", "backups_retain_number", 1001, "backups_retain_number: Must be less than or equal to 1000"),
			Entry("invalid backups_location", "backups_location", "--moon", "backups_location: Does not match pattern '^[a-z][a-z0-9-]+$'"),
			Entry("invalid backups_start_time", "backups_start_time", "34:91", `backups_start_time: Does not match pattern`),
			Entry("min backups_point_in_time_log_retain_days", "backups_point_in_time_log_retain_days", -1, "backups_point_in_time_log_retain_days: Must be greater than or equal to 0"),
			Entry("max backups_point_in_time_log_retain_days", "backups_point_in_time_log_retain_days", 8, "backups_point_in_time_log_retain_days: Must be less than or equal to 7"),
		)
	})

	Describe("high-availability", func() {
		It("allows high-availability to be configured", func() {
			_, err := broker.Provision("csb-google-postgres", postgresNoOverridesPlan["name"].(string), map[string]any{
				"tier":                                  "db-f1-micro",
				"backups_retain_number":                 7,
				"backups_point_in_time_log_retain_days": 7,
				"highly_available":                      true,
			})

			Expect(err).NotTo(HaveOccurred())
			invocations, err := mockTerraform.ApplyInvocations()
			Expect(err).NotTo(HaveOccurred())
			Expect(invocations).To(HaveLen(1))
			Expect(invocations[0].TFVars()).To(SatisfyAll(
				HaveKeyWithValue("backups_retain_number", BeNumerically("==", 7)),
				HaveKeyWithValue("backups_point_in_time_log_retain_days", BeNumerically("==", 7)),
				HaveKeyWithValue("highly_available", BeTrue()),
			))
		})
	})
})
