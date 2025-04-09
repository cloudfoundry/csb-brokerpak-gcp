package integration_test

import (
	"fmt"

	testframework "github.com/cloudfoundry/cloud-service-broker/v2/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

const (
	mySQLServiceName        = "csb-google-mysql"
	mySQLServiceID          = "fe6a86d5-ce06-4c58-96f8-43aef1ac8a31"
	mySQLServiceDisplayName = "Google Cloud MySQL"
	mySQLServiceDescription = "MySQL is a fully managed service for the Google Cloud Platform."
	mySQLServiceSupportURL  = "https://cloud.google.com/support/"
	mySQLCustomPlanName     = "custom-plan"
	mySQLCustomPlanID       = "9daa07f1-78e8-4bda-9efe-91576102c30d"
)

var customMySQLPlans = []map[string]any{
	customMySQLPlan,
}

var customMySQLPlan = map[string]any{
	"name":          mySQLCustomPlanName,
	"id":            mySQLCustomPlanID,
	"description":   "custom plan defined by customer",
	"mysql_version": "MYSQL_8_0",
	"metadata": map[string]any{
		"displayName": "custom plan defined by customer (beta)",
	},
}

var _ = Describe("MySQL", Label("MySQL"), func() {
	BeforeEach(func() {
		Expect(mockTerraform.SetTFState([]testframework.TFStateValue{})).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(mockTerraform.Reset()).To(Succeed())
	})

	It("publishes in the catalog", func() {
		catalog, err := broker.Catalog()
		Expect(err).NotTo(HaveOccurred())

		service := testframework.FindService(catalog, mySQLServiceName)
		Expect(service.ID).To(Equal(mySQLServiceID))
		Expect(service.Description).To(Equal(mySQLServiceDescription))
		Expect(service.Tags).To(ConsistOf("gcp", "mysql"))
		Expect(service.Metadata.ImageUrl).To(ContainSubstring("data:image/png;base64,"))
		Expect(service.Metadata.DisplayName).To(Equal(mySQLServiceDisplayName))
		Expect(service.Metadata.DocumentationUrl).To(Equal(cloudServiceBrokerDocumentationURL))
		Expect(service.Metadata.ProviderDisplayName).To(Equal(providerDisplayName))
		Expect(service.Metadata.SupportUrl).To(Equal(mySQLServiceSupportURL))
		Expect(service.Plans).To(
			ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					ID:   Equal(mySQLCustomPlanID),
					Name: Equal(mySQLCustomPlanName),
				}),
			),
		)
	})

	Describe("provisioning", func() {
		It("should provision a plan", func() {
			instanceID, err := broker.Provision(mySQLServiceName, mySQLCustomPlanName, map[string]any{"tier": "db-n1-standard-1"})

			Expect(err).NotTo(HaveOccurred())
			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("db_name", "csb-db"),
					HaveKeyWithValue("authorized_network_id", BeAssignableToTypeOf("")),
					HaveKeyWithValue("public_ip", false),
					HaveKeyWithValue("authorized_networks_cidrs", []any{}),
					HaveKeyWithValue("credentials", "broker-gcp-creds"),
					HaveKeyWithValue("mysql_version", "MYSQL_8_0"),
					HaveKeyWithValue("db_name", "csb-db"),
					HaveKeyWithValue("instance_name", "csb-mysql-"+instanceID),
					HaveKeyWithValue("project", "broker-gcp-project"),
					HaveKeyWithValue("region", "us-central1"),
					HaveKeyWithValue("storage_gb", BeNumerically("==", 10)),

					HaveKeyWithValue("tier", "db-n1-standard-1"),
					HaveKeyWithValue("disk_autoresize", true),
					HaveKeyWithValue("disk_autoresize_limit", BeNumerically("==", 0)),
					HaveKeyWithValue("deletion_protection", false),
					HaveKeyWithValue("backups_start_time", "07:00"),
					HaveKeyWithValue("backups_location", BeNil()),
					HaveKeyWithValue("backups_retain_number", BeNumerically("==", 7)),
					HaveKeyWithValue("backups_transaction_log_retention_days", BeNumerically("==", 0)),
				),
			)
		})

		It("should allow setting properties not defined in the plan", func() {
			_, err := broker.Provision(mySQLServiceName, mySQLCustomPlanName, map[string]any{
				"credentials":                            "fake-credentials",
				"project":                                "fake-project",
				"instance_name":                          "fakeinstancename",
				"db_name":                                "fake-db_name",
				"region":                                 "asia-northeast1",
				"authorized_network_id":                  "fake-authorized_network_id",
				"tier":                                   "fake-tier",
				"disk_autoresize":                        true,
				"disk_autoresize_limit":                  400,
				"deletion_protection":                    true,
				"backups_start_time":                     "12:34",
				"backups_location":                       "somewhere-over-the-rainbow12",
				"backups_retain_number":                  5,
				"backups_transaction_log_retention_days": 6,
				"public_ip":                              true,
				"authorized_networks_cidrs":              []any{"one", "two"},
				"highly_available":                       true,
				"location_preference_zone":               "a",
				"location_preference_secondary_zone":     "c",
				"maintenance_day":                        2,
				"maintenance_hour":                       7,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("credentials", "fake-credentials"),
					HaveKeyWithValue("project", "fake-project"),
					HaveKeyWithValue("instance_name", "fakeinstancename"),
					HaveKeyWithValue("db_name", "fake-db_name"),
					HaveKeyWithValue("region", "asia-northeast1"),
					HaveKeyWithValue("authorized_network_id", "fake-authorized_network_id"),
					HaveKeyWithValue("tier", "fake-tier"),
					HaveKeyWithValue("disk_autoresize", true),
					HaveKeyWithValue("disk_autoresize_limit", BeNumerically("==", 400)),
					HaveKeyWithValue("deletion_protection", BeTrue()),
					HaveKeyWithValue("backups_start_time", BeEquivalentTo("12:34")),
					HaveKeyWithValue("backups_location", Equal("somewhere-over-the-rainbow12")),
					HaveKeyWithValue("backups_retain_number", BeNumerically("==", 5)),
					HaveKeyWithValue("backups_transaction_log_retention_days", BeNumerically("==", 6)),
					HaveKeyWithValue("public_ip", BeTrue()),
					HaveKeyWithValue("authorized_networks_cidrs", ConsistOf("one", "two")),
					HaveKeyWithValue("highly_available", BeTrue()),
					HaveKeyWithValue("location_preference_zone", Equal("a")),
					HaveKeyWithValue("location_preference_secondary_zone", Equal("c")),
					HaveKeyWithValue("maintenance_day", BeNumerically("==", 2)),
					HaveKeyWithValue("maintenance_hour", BeNumerically("==", 7)),
				),
			)
		})

		It("should not allow changing of plan defined properties", func() {
			_, err := broker.Provision(mySQLServiceName, mySQLCustomPlanName, map[string]any{"mysql_version": "MYSQL_8_0"})

			Expect(err).To(MatchError(ContainSubstring("plan defined properties cannot be changed: mysql_version")))
		})

		DescribeTable("property constraints",
			func(params map[string]any, expectedErrorMsg string) {
				_, err := broker.Provision(mySQLServiceName, mySQLCustomPlanName, params)

				Expect(err).To(MatchError(ContainSubstring(expectedErrorMsg)))
			},
			Entry(
				"storage capacity minimum value is 10",
				map[string]any{"storage_gb": 9},
				"storage_gb: Must be greater than or equal to 10",
			),
			Entry(
				"instance name maximum length is 98 characters",
				map[string]any{"instance_name": stringOfLen(99)},
				"instance_name: String length must be less than or equal to 98",
			),
			Entry(
				"instance name minimum length is 6 characters",
				map[string]any{"instance_name": stringOfLen(5)},
				"instance_name: String length must be greater than or equal to 6",
			),
			Entry(
				"instance name invalid characters",
				map[string]any{"instance_name": ".aaaaa"},
				"instance_name: Does not match pattern '^[a-z][a-z0-9-]+[a-z0-9]$'",
			),
			Entry(
				"database name maximum length is 64 characters",
				map[string]any{"db_name": stringOfLen(65)},
				"db_name: String length must be less than or equal to 64",
			),
			Entry(
				"invalid region",
				map[string]any{"region": "-Asia-northeast1"},
				"region: Does not match pattern '^[a-z][a-z0-9-]+[a-z0-9]$'",
			),
			Entry(
				"tier invalid characters",
				map[string]any{"tier": ".aaaaa"},
				"tier: Does not match pattern '^[a-z][a-z0-9-]+[a-z0-9]$'",
			),
			Entry(
				"invalid backup location",
				map[string]any{"backups_location": "australia-central-"},
				"backups_location: Does not match pattern '^[a-z][a-z0-9-]+[a-z0-9]$'",
			),
			Entry(
				"invalid backups retain number",
				map[string]any{"backups_retain_number": -7},
				"backups_retain_number: Must be greater than or equal to 0",
			),
			Entry(
				"invalid transaction log retention days",
				map[string]any{"backups_transaction_log_retention_days": -1},
				"backups_transaction_log_retention_days: Must be greater than or equal to 0",
			),
			Entry(
				"invalid transaction log retention days",
				map[string]any{"backups_transaction_log_retention_days": 8},
				"backups_transaction_log_retention_days: Must be less than or equal to 7",
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
				"invalid maintenance day",
				map[string]any{"maintenance_day": 8},
				"maintenance_day: Must be less than or equal to 7",
			),
			Entry(
				"invalid maintenance hour",
				map[string]any{"maintenance_hour": 24},
				"maintenance_hour: Must be less than or equal to 23",
			),
		)
	})

	Describe("updating instance", func() {
		var instanceID string

		BeforeEach(func() {
			var err error
			instanceID, err = broker.Provision(mySQLServiceName, mySQLCustomPlanName, map[string]any{"tier": "db-n1-standard-1"})

			Expect(err).NotTo(HaveOccurred())
		})

		DescribeTable("should allow updating properties not flagged as `prohibit_update` and not specified in the plan",
			func(params map[string]any) {
				err := broker.Update(instanceID, mySQLServiceName, mySQLCustomPlanName, params)

				Expect(err).NotTo(HaveOccurred())
			},
			Entry("update credentials", map[string]any{"credentials": "other-credentials"}),
			Entry("update project", map[string]any{"project": "another-project"}),
			Entry("update tier", map[string]any{"tier": "db-n1-standard-16"}),
			Entry("update disk_autoresize", map[string]any{"disk_autoresize": true}),
			Entry("update disk_autoresize_limit", map[string]any{"disk_autoresize_limit": 400}),
			Entry("update storage_gb", map[string]any{"storage_gb": 100}),
			Entry("update deletion_protection", map[string]any{"deletion_protection": true}),
			Entry("update backups_start_time", map[string]any{"backups_start_time": "22:33"}),
			Entry("update backups_location", map[string]any{"backups_location": "safety-deposit-box"}),
			Entry("update backups_retain_number", map[string]any{"backups_retain_number": 0}),
			Entry("update backups_transaction_log_retention_days", map[string]any{"backups_transaction_log_retention_days": 1}),
			Entry("update public_ip", map[string]any{"public_ip": true}),
			Entry("update authorized_networks_cidrs", map[string]any{"authorized_networks_cidrs": []any{"three", "four"}}),
			Entry("update highly_available", map[string]any{"highly_available": true}),
			Entry("update location_preference_zone", map[string]any{"location_preference_zone": "a"}),
			Entry("update location_preference_secondary_zone", map[string]any{"location_preference_secondary_zone": "c"}),
			Entry("update maintenance_day", map[string]any{"maintenance_day": 3}),
			Entry("update maintenance_hour", map[string]any{"maintenance_hour": 8}),
		)

		DescribeTable("should prevent updating properties flagged as `prohibit_update` because it can result in the recreation of the service instance and lost data",
			func(params map[string]any) {
				err := broker.Update(instanceID, mySQLServiceName, mySQLCustomPlanName, params)

				Expect(err).To(MatchError(
					ContainSubstring(
						"attempt to update parameter that may result in service instance re-creation and data loss",
					),
				))

				const initialProvisionInvocation = 1
				Expect(mockTerraform.ApplyInvocations()).To(HaveLen(initialProvisionInvocation))
			},
			Entry("update instance_name", map[string]any{"instance_name": "another-instance-name"}),
			Entry("update db_name", map[string]any{"db_name": "another-db-name"}),
			Entry("update region", map[string]any{"region": "australia-southeast1"}),
			Entry("update authorized_network_id", map[string]any{"authorized_network_id": "another-authorized-network_id"}),
			Entry("update allow_insecure_connections", map[string]any{"allow_insecure_connections": true}),
		)

		DescribeTable("should not allow updating properties that are specified in the plan",
			func(key string, value any) {
				err := broker.Update(instanceID, mySQLServiceName, mySQLCustomPlanName, map[string]any{key: value})

				Expect(err).To(
					MatchError(
						ContainSubstring(
							fmt.Sprintf("plan defined properties cannot be changed: %s", key),
						),
					),
				)
			},
			Entry("update mysql_version", "mysql_version", "MYSQL_8_0"),
		)

		DescribeTable("should not allow updating additional properties",
			func(key string, value any) {
				err := broker.Update(instanceID, mySQLServiceName, mySQLCustomPlanName, map[string]any{key: value})

				Expect(err).To(
					MatchError(
						ContainSubstring(
							fmt.Sprintf("additional properties are not allowed: %s", key),
						),
					),
				)
			},
			Entry("update name", "name", "fake-name"),
			Entry("update id", "id", "fake-id"),
		)
	})

	DescribeTable("bind a service",
		func(bindParams map[string]any) {
			const (
				fakeSSLRoot    = "CREATED_SSL_ROOT_CERT"
				fakeClientCert = "CREATED_SSL_CLIENT_CERT"
				fakeClientKey  = "CREATED_SSL_CLIENT_KEY"
				fakePrivateIP  = "CREATED_PRIVATE_IP"
			)
			instanceID := provisionInstanceForBinding(
				fakeSSLRoot,
				fakeClientCert,
				fakeClientKey,
				fakePrivateIP,
			)

			// mocking Terraform state for binding propose
			err := mockTerraform.SetTFState([]testframework.TFStateValue{
				{Name: "username", Type: "string", Value: "bind.test.username"},
				{Name: "password", Type: "string", Value: "bind.test.password"},
				{Name: "uri", Type: "string", Value: "bind.test.uri"},
				{Name: "jdbcUrl", Type: "string", Value: "bind.test.jdbcUrl"},
				{Name: "port", Type: "number", Value: 3306},
			})
			Expect(err).NotTo(HaveOccurred())

			bindResult, err := broker.Bind(mySQLServiceName, mySQLCustomPlanName, instanceID, bindParams)
			Expect(err).NotTo(HaveOccurred())

			Expect(bindResult).To(Equal(map[string]any{
				"username":                   "bind.test.username",
				"hostname":                   "created.hostname.gcp.test",
				"jdbcUrl":                    "bind.test.jdbcUrl",
				"name":                       "created.test.instancename",
				"password":                   "bind.test.password",
				"uri":                        "bind.test.uri",
				"sslrootcert":                fakeSSLRoot,
				"sslcert":                    fakeClientCert,
				"sslkey":                     fakeClientKey,
				"private_ip":                 fakePrivateIP,
				"port":                       float64(3306),
				"allow_insecure_connections": false,
			}))
		},
		Entry(
			"bind with default parameters",
			nil,
		),
		Entry(
			"bind with read-only = true",
			map[string]any{"read_only": true},
		),
	)
})

func provisionInstanceForBinding(
	fakeSSLRoot,
	fakeClientCert,
	fakeClientKey,
	fakePrivateIP string,
) string {
	// used in computed outputs in binding definition
	err := mockTerraform.SetTFState([]testframework.TFStateValue{
		{Name: "hostname", Type: "string", Value: "created.hostname.gcp.test"},
		{Name: "username", Type: "string", Value: "created.test.username"},
		{Name: "password", Type: "string", Value: "created.test.password"},
		{Name: "name", Type: "string", Value: "created.test.instancename"},
		{Name: "sslrootcert", Type: "string", Value: fakeSSLRoot},
		{Name: "sslcert", Type: "string", Value: fakeClientCert},
		{Name: "sslkey", Type: "string", Value: fakeClientKey},
		{Name: "private_ip", Type: "string", Value: fakePrivateIP},
		{Name: "allow_insecure_connections", Type: "boolean", Value: false},
	})
	Expect(err).NotTo(HaveOccurred())

	instanceID, err := broker.Provision(mySQLServiceName, mySQLCustomPlanName, map[string]any{"tier": "db-n1-standard-1"})
	Expect(err).NotTo(HaveOccurred())
	return instanceID
}
